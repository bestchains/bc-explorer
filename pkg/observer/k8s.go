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
	"reflect"
	"sort"
	"sync"
	"time"

	"github.com/bestchains/bc-explorer/pkg/network"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/IBM-Blockchain/fabric-operator/api/v1beta1"
	"github.com/IBM-Blockchain/fabric-operator/pkg/generated/clientset/versioned"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"k8s.io/utils/strings/slices"
)

var (
	ErrWrongTypeChannel   = errors.New("wrong type of channel")
	ErrWrongTypeConfigmap = errors.New("wrong type of configmap")
)

type Watcher struct {
	Msg               chan<- Msg
	Client            *kubernetes.Clientset
	VClient           *versioned.Clientset
	Send              sync.Map
	OperatorNamespace string
}

func NewWatcher(msg chan<- Msg, client *kubernetes.Clientset, vclient *versioned.Clientset, operatorNamespace string) *Watcher {
	return &Watcher{
		Msg:               msg,
		Client:            client,
		VClient:           vclient,
		Send:              sync.Map{},
		OperatorNamespace: operatorNamespace,
	}
}

func (w *Watcher) ChannelUpdate(old interface{}, new interface{}) {
	klog.V(5).Infoln("update channel")
	key, err := cache.MetaNamespaceKeyFunc(new)
	if err != nil {
		runtime.HandleError(err)
		return
	}
	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(err)
		return
	}
	channel, ok := new.(*v1beta1.Channel)
	if !ok {
		klog.ErrorS(ErrWrongTypeChannel, "get wrong type of network", "obj", name)
		return
	}
	oldChannel, ok := old.(*v1beta1.Channel)
	if !ok {
		klog.ErrorS(ErrWrongTypeChannel, "get wrong type of network", "obj", name)
		return
	}
	if reflect.DeepEqual(channel.Spec, oldChannel.Spec) && reflect.DeepEqual(channel.Status, oldChannel.Status) {
		klog.V(5).Infof("channel %s updated but has same spec and status, update detail:%s, just skip", channel.GetName(), cmp.Diff(oldChannel, channel))
		return
	}
	if err := w.HandleProfile(context.TODO(), w.OperatorNamespace, channel); err != nil {
		runtime.HandleError(err)
		return
	}
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
	channel, ok := obj.(*v1beta1.Channel)
	if !ok {
		klog.ErrorS(ErrWrongTypeChannel, "get wrong type of channel", "obj", name)
		return
	}
	msg := Msg{
		ChannelID:   channel.GetChannelID(),
		NetworkName: channel.Spec.Network,
		Type:        Delete,
		Data:        nil,
	}
	w.sendMsg(msg)
}

func (w *Watcher) HandleProfile(ctx context.Context, operatorNamespace string, channel *v1beta1.Channel) (err error) {
	if len(channel.Spec.Peers) == 0 {
		klog.V(5).Infof("skip channel:%s send because of no peers", channel.GetName())
		return
	}
	cmName := channel.GetConnectionPorfile()
	cm, err := w.Client.CoreV1().ConfigMaps(operatorNamespace).Get(ctx, cmName, metav1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, "cant get channel connection profile configmap")
	}

	fabProfile, err := w.parseDataFromConfigmap(cm, channel)
	if err != nil {
		return err
	}
	w.SendProfile(fabProfile, channel.Spec.Network, channel.GetChannelID(), channel.Status.Type)
	return
}

func (w *Watcher) parseDataFromConfigmap(configmap *corev1.ConfigMap, channel *v1beta1.Channel) (fabProfile *network.FabProfile, err error) {
	data := configmap.BinaryData["profile.json"]
	if data == nil {
		return nil, fmt.Errorf("no profile.json in configmap:%s in ns:%s", configmap.Name, configmap.Namespace)
	}
	profile := &Profile{}
	if err := json.Unmarshal(data, profile); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("configmap.BinaryData.'profile.json' json unmarshal error, configmap:%s in ns:%s", configmap.Name, configmap.Namespace))
	}
	fabProfile = &network.FabProfile{}
	fabProfile.Channel = channel.GetChannelID()
	// Always use the connection profile of the peer as the first in alphabetical
	peers := make([]string, 0)
	for peerName := range profile.Peers {
		peers = append(peers, peerName)
	}
	if len(peers) == 0 {
		return nil, fmt.Errorf("no peers find in comfigmap:%s in ns:%s for channel:%s", configmap.Name, configmap.Namespace, channel.GetName())
	}
	sort.Strings(peers)
	wantPeer := peers[0]
	for orgName, value := range profile.Organizations {
		if !slices.Contains(value.Peers, wantPeer) {
			continue
		}
		fabProfile.Organization = orgName
		users := make([]string, 0)
		for userName := range value.Users {
			users = append(users, userName)
		}
		if len(users) == 0 {
			return nil, fmt.Errorf("has peer, but no user find for org:%s in comfigmap:%s in ns:%s for channel:%s", orgName, configmap.Name, configmap.Namespace, channel.GetName())
		}
		sort.Strings(users)
		for userName, v := range value.Users {
			if userName != users[0] {
				continue
			}
			fabProfile.User.Name = v.Name
			fabProfile.User.Key.Pem = v.Key.Pem
			fabProfile.User.Cert.Pem = v.Cert.Pem
		}
	}
	for peerName, value := range profile.Peers {
		if wantPeer != peerName {
			continue
		}
		fabProfile.Enpoint.URL = value.URL
		fabProfile.Enpoint.TLSCACerts.Pem = value.TLSCACerts.Pem
	}
	return fabProfile, nil
}

func key(networkName, channelID string) string {
	return networkName + "_" + channelID
}

func (w *Watcher) ProfileConfigMapCreate(obj interface{}) {
	klog.V(5).Infoln("new configmap create")
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
	cm, ok := obj.(*corev1.ConfigMap)
	if !ok {
		klog.ErrorS(ErrWrongTypeConfigmap, "get wrong type of configmap", "obj", name)
		return
	}

	channelName, err := getConfigMapOwnerChannel(cm)
	if err != nil {
		klog.ErrorS(err, "cant get channel name", "configmap", name)
		return
	}
	channel, err := w.VClient.Ibp().Channels().Get(context.TODO(), channelName, metav1.GetOptions{})
	if err != nil {
		klog.ErrorS(err, "cant get channel", "configmap", name)
		return
	}
	fabProfile, err := w.parseDataFromConfigmap(cm, channel)
	if err != nil {
		klog.ErrorS(err, "cant parse data from configmap", "configmap", name)
		return
	}
	w.SendProfile(fabProfile, channel.Spec.Network, channel.GetChannelID(), channel.Status.Type)
}

func (w *Watcher) ProfileConfigMapUpdate(old interface{}, new interface{}) {
	klog.V(5).Infoln("update configmap")
	key, err := cache.MetaNamespaceKeyFunc(new)
	if err != nil {
		runtime.HandleError(err)
		return
	}
	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(err)
		return
	}
	cm, ok := new.(*corev1.ConfigMap)
	if !ok {
		klog.ErrorS(ErrWrongTypeConfigmap, "get wrong type of configmap", "obj", name)
		return
	}
	oldCm, ok := old.(*corev1.ConfigMap)
	if !ok {
		klog.ErrorS(ErrWrongTypeConfigmap, "get wrong type of configmap", "obj", name)
		return
	}
	if reflect.DeepEqual(cm.BinaryData, oldCm.BinaryData) && reflect.DeepEqual(cm.OwnerReferences, oldCm.OwnerReferences) {
		klog.V(5).Infof("configmap %s updated but has same BinaryData and ownerReferences, update detail:%s, just skip", cm.GetName(), cmp.Diff(oldCm, cm))
		return
	}
	w.ProfileConfigMapCreate(new)
}

func getConfigMapOwnerChannel(cm *corev1.ConfigMap) (channelName string, err error) {
	if len(cm.OwnerReferences) == 0 {
		return "", fmt.Errorf("configmap:%s in ns:%s has no owerReference", cm.Name, cm.Namespace)
	}
	for _, owner := range cm.OwnerReferences {
		if owner.Kind == "Channel" && owner.APIVersion == "ibp.com/v1beta1" {
			return owner.Name, nil
		}
	}
	return "", fmt.Errorf("configmap:%s in ns:%s has owerReference, but no one is channel", cm.Name, cm.Namespace)
}

func (w *Watcher) SendProfile(fabProfile *network.FabProfile, networkName, channelID string, channelStatus v1beta1.IBPCRStatusType) {
	if networkName == "" {
		klog.V(5).Infof("channel %s, no networkName, skip send", channelID)
		return
	}
	if channelID == "" {
		klog.V(5).Infof("network %s no channelID, skip send", networkName)
		return
	}
	data := &network.Network{}
	data.FabProfile = fabProfile
	data.ID = networkName
	data.Platform = "bestchains"
	msg := Msg{
		ChannelID:   channelID,
		NetworkName: networkName,
		Data:        data,
	}
	switch channelStatus {
	case v1beta1.ChannelArchived:
		msg.Type = Deregister
	default:
		msg.Type = Register
	}
	w.sendMsg(msg)
}

func (w *Watcher) sendMsg(msg Msg) {
	key := key(msg.NetworkName, msg.ChannelID)
	if value, exist := w.Send.Load(key); exist {
		oldMsg, ok := (value).(Msg)
		if !ok {
			klog.InfoS("cant get msg from sync.Map, skip send", "network", msg.NetworkName, "channel", msg.ChannelID)
			return
		}
		if reflect.DeepEqual(msg, oldMsg) {
			klog.InfoS("has send to listener, skip resend", "network", msg.NetworkName, "channel", msg.ChannelID)
			return
		}
	}
	w.Msg <- msg
	w.Send.Store(key, msg)
	if msg.Type == Delete {
		go func() {
			// When msg.Type is Delete, it means that the channel has been deleted in the cluster.
			// We should also delete the data in sync.Map to prevent memory leaks.
			// In order to filter out possible multiple deletion events, we will delay for 5 minutes before deleting.
			time.Sleep(5 * time.Minute)
			value, exist := w.Send.Load(key)
			if exist {
				msg, ok := (value).(Msg)
				if ok && msg.Type == Delete {
					w.Send.Delete(key)
				}
			}
		}()
	}
}
