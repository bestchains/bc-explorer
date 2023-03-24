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

package network

import (
	"github.com/pkg/errors"
	"k8s.io/klog/v2"

	"github.com/hyperledger/fabric-gateway/pkg/client"
)

var (
	errMissingFabNetProfile = errors.New("missing fabric network's profile")
	errMissingFabChannel    = errors.New("missing field channel in fabric network's profile")
	errInvalidX509Cert      = errors.New("invalid x509 certificate")
)

type Platform string

const (
	Bestchains Platform = "bestchains"
)

type Type string

const (
	Unknown Type = "Unknown"
	FABRIC  Type = "Fabric"
)

type Network struct {
	ID          string `json:"id"` // network metadat.name
	Platform    `json:"platform"`
	*FabProfile `json:"fabProfile,omitempty"`
}

func (n *Network) Type() Type {
	if n.FabProfile != nil {
		return FABRIC
	}
	return Unknown
}

type FabricClient struct {
	gw *client.Gateway

	primaryChannel *client.Network
}

func NewFabricClient(n *Network) (*FabricClient, error) {
	var err error

	if n.FabProfile == nil {
		klog.Error(errMissingFabNetProfile)
		return nil, errMissingFabNetProfile
	}

	profile := n.FabProfile
	if profile.Channel == "" {
		klog.Error(errMissingFabChannel)
		return nil, errMissingFabChannel
	}

	klog.V(5).Infof("initialize a fabric client conn for network: %s", n.ID)
	clientConn, err := newFabClientConn(profile)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	klog.V(5).Infof("initialize a fabric identity for network: %s", n.ID)
	id, sign, err := profile.User.ToIdentity(profile.Organization)
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	klog.V(5).Infof("connect to network: %s", n.ID)
	gateway, err := client.Connect(id, client.WithSign(sign), client.WithClientConnection(clientConn))
	if err != nil {
		klog.Error(err)
		return nil, err
	}

	return &FabricClient{
		gw:             gateway,
		primaryChannel: gateway.GetNetwork(profile.Channel),
	}, nil
}

func (fabclient *FabricClient) Channel(channel string) *client.Network {
	if channel != "" {
		return fabclient.gw.GetNetwork(channel)
	}
	return fabclient.primaryChannel

}

func (fabclient *FabricClient) Close() {
	fabclient.gw.Close()
}
