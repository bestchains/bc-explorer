/*
Copyright 2023 The Bestchains Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package observer

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/IBM-Blockchain/fabric-operator/pkg/generated/informers/externalversions"
	"github.com/bestchains/bc-explorer/pkg/network"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"github.com/IBM-Blockchain/fabric-operator/pkg/generated/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	coreInformers "k8s.io/client-go/informers/core/v1"

	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

func Run(ctx context.Context, config *rest.Config, host, operatorNamespace string) (err error) {
	defer runtime.HandleCrash()
	klog.V(5).Infof("observer start...")
	vclient, err := versioned.NewForConfig(config)
	if err != nil {
		return err
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	informerFactory := externalversions.NewSharedInformerFactory(vclient, 0)
	channelInformer := informerFactory.Ibp().V1beta1().Channels()
	msg := make(chan Msg, 100)
	watcher := NewWatcher(msg, client, vclient, operatorNamespace)
	channelInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		// channelUpdate can watch channel status change, eg Archived
		UpdateFunc: watcher.ChannelUpdate,
		// channelDelete can watch channel delete
		DeleteFunc: watcher.ChannelDelete,
	})
	informerFactory.Start(ctx.Done())
	if !cache.WaitForCacheSync(ctx.Done(), channelInformer.Informer().HasSynced) {
		err := fmt.Errorf("waitForCacheSync failed")
		klog.ErrorS(err, "cannot sync caches")
		return err
	}

	conConfigMapInformer := coreInformers.NewConfigMapInformer(client, operatorNamespace, 12*time.Hour, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	conConfigMapInformer.AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			if cast, ok := obj.(*corev1.ConfigMap); ok {
				return IsConfigMapHasProfile(cast.Name)
			}
			if tombstone, ok := obj.(cache.DeletedFinalStateUnknown); ok {
				if cast, ok := tombstone.Obj.(*corev1.ConfigMap); ok {
					return IsConfigMapHasProfile(cast.Name)
				}
			}
			return false
		},
		Handler: cache.ResourceEventHandlerFuncs{
			// profileConfigmapCreate can watch channel create
			AddFunc: watcher.ProfileConfigMapCreate,
			// profileConfigmapUpdate can watch channel update, peers update, user cert update and so on.
			UpdateFunc: watcher.ProfileConfigMapUpdate,
		},
	})
	go conConfigMapInformer.Run(ctx.Done())
	klog.V(5).Infoln("observer init finish.")
	pusher := NewPusher(host, msg)
	pusher.Run(ctx)
	return nil
}

type Msg struct {
	ChannelName string
	NetworkName string
	Type        MsgType
	Data        *network.Network
}

type MsgType int

const (
	Register   MsgType = 1
	Deregister MsgType = 1 << iota
	Delete     MsgType = 1 << iota
)

func IsConfigMapHasProfile(name string) bool {
	return strings.HasPrefix(name, "chan-") && strings.HasSuffix(name, "-connection-profile")
}
