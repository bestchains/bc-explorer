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
	"encoding/json"
	"fmt"
	"sort"
	"sync"

	"github.com/bestchains/bc-explorer/pkg/network"
	"github.com/pkg/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/IBM-Blockchain/fabric-operator/api/v1beta1"
	"github.com/IBM-Blockchain/fabric-operator/pkg/generated/clientset/versioned"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

var (
	WrongTypeChannelErr = errors.New("wrong type of channel")
)

type Watcher struct {
	Profile     chan<- network.Network
	DeleteNames chan<- string
	Client      *kubernetes.Clientset
	VClient     *versioned.Clientset
	Send        sync.Map
}

func NewWatcher(profile chan<- network.Network, deleteNames chan<- string, client *kubernetes.Clientset, vclient *versioned.Clientset) *Watcher {
	return &Watcher{
		Profile:     profile,
		DeleteNames: deleteNames,
		Client:      client,
		VClient:     vclient,
		Send:        sync.Map{},
	}
}

func (w *Watcher) ChannelCreate(obj interface{}) {
	klog.V(5).Infoln("get new channel")
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
		return
	}
	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(err)
		return
	}
	channel, ok := obj.(*v1beta1.Channel)
	if !ok {
		klog.ErrorS(WrongTypeChannelErr, "get wrong type of network", "obj", name)
		return
	}
	if err := w.GetProfile(context.TODO(), channel); err != nil {
		runtime.HandleError(err)
		return
	}
}

func (w *Watcher) ChannelUpdate(oldObj interface{}, newObj interface{}) {
	klog.V(5).Infoln("update channel")
	w.ChannelCreate(newObj)
}

func (w *Watcher) ChannelDelete(obj interface{}) {
	klog.V(5).Infoln("delete channel")
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
		return
	}
	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(err)
		return
	}
	_, ok := obj.(*v1beta1.Channel)
	if !ok {
		klog.ErrorS(WrongTypeChannelErr, "get wrong type of channel", "obj", name)
		return
	}
	w.DeleteNames <- name
	w.Send.Delete(name)
}

func (w *Watcher) GetProfile(ctx context.Context, channel *v1beta1.Channel) (err error) {
	channelName := channel.GetName()
	if _, exist := w.Send.Load(channelName); exist {
		klog.V(5).Infof("channel %s has send to listener", channelName)
		return
	}
	if len(channel.Spec.Peers) == 0 {
		klog.V(5).Infof("skip channel:%s send because of no peers", channelName)
		return
	}
	sort.Slice(channel.Spec.Peers, func(i, j int) bool {
		return channel.Spec.Peers[i].Name > channel.Spec.Peers[j].Name
	})
	cmName := channel.GetConnectionPorfile()
	cmNs := channel.Spec.Peers[0].Namespace
	cm, err := w.Client.CoreV1().ConfigMaps(cmNs).Get(ctx, cmName, metav1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, "cant get channel connection profile configmap")
	}
	data := cm.BinaryData["profile.json"]
	if data == nil {
		return fmt.Errorf("no profile.json in configmap:%s in ns:%s", cmName, cmNs)
	}
	profile := &Profile{}
	if err := json.Unmarshal(data, profile); err != nil {
		return errors.Wrap(err, fmt.Sprintf("configmap.BinaryData.'profile.json' json unmarshal error, configmap:%s in ns:%s", cmName, cmNs))
	}
	fabProfile := &network.FabProfile{}
	fabProfile.Channel = channelName
	for key, value := range profile.Organizations {
		fabProfile.Organization = key
		for _, v := range value.Users {
			fabProfile.User.Name = v.Name
			fabProfile.User.Key.Pem = v.Key.Pem
			fabProfile.User.Cert.Pem = v.Cert.Pem
			break
		}
		break
	}
	for _, value := range profile.Peers {
		fabProfile.Enpoint.URL = value.URL
		fabProfile.Enpoint.TLSCACerts.Pem = value.TLSCACerts.Pem
		break
	}
	n := network.Network{}
	n.FabProfile = fabProfile
	n.ID = channel.Spec.Network
	n.Platform = "bestchains"
	w.Profile <- n
	w.Send.Load(channelName)
	return
}
