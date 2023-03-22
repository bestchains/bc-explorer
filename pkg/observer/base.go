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

	"github.com/bestchains/bc-explorer/pkg/network"

	"github.com/IBM-Blockchain/fabric-operator/pkg/generated/informers/externalversions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"github.com/IBM-Blockchain/fabric-operator/pkg/generated/clientset/versioned"

	"github.com/pkg/errors"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

func Run(ctx context.Context, host, kubeConfigPath string) (err error) {
	klog.V(5).Infof("observer start...")
	var config *rest.Config
	if kubeConfigPath == "" {
		config, err = rest.InClusterConfig()
		if err != nil {
			return errors.Wrap(err, "create in-cluster client configuration error")
		}
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		if err != nil {
			return errors.Wrap(err, "create out-of-cluster client configuration error")
		}
	}

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
	profile := make(chan network.Network, 100)
	names := make(chan string, 100)
	watcher := NewWatcher(profile, names, client, vclient)
	channelInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    watcher.ChannelCreate,
		UpdateFunc: watcher.ChannelUpdate,
		DeleteFunc: watcher.ChannelDelete,
	})
	informerFactory.Start(ctx.Done())
	if !cache.WaitForCacheSync(ctx.Done(), channelInformer.Informer().HasSynced) {
		err := fmt.Errorf("waitForCacheSync failed")
		klog.ErrorS(err, "cannot sync caches")
		return err
	}
	klog.V(5).Infoln("observer init finish.")
	pusher := NewPusher(host, profile, names)
	pusher.Run(ctx)
	return nil
}
