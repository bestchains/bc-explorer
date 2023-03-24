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

package listener

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/bestchains/bc-explorer/pkg/errorsq"
	"github.com/bestchains/bc-explorer/pkg/models"
	"github.com/bestchains/bc-explorer/pkg/network"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
)

var (
	errListenerMissingField  = errors.New("listener missing important fields")
	errListNetworks          = errors.New("failed to list networks")
	errNetworkTypeUnknown    = errors.New("unknown network type")
	errInvalidNetworkProfile = errors.New("invalid network profile")
	errNetworkAlreadyExists  = errors.New("network with this id already exists in this listener")
)

type Listener interface {
	Selector() Selector
	Register(*network.Network) error
	Deregister(string) error
	Delete(string) error
}

type listener struct {
	lock sync.Mutex
	ctx  context.Context

	errq errorsq.Errorsq

	injector Injector
	selector Selector

	networks map[string]BlockEventListener
}

func NewListener(ctx context.Context, errq errorsq.Errorsq, injector Injector, selector Selector) (Listener, error) {
	if errq == nil || injector == nil || selector == nil {
		return nil, errListenerMissingField
	}
	l := &listener{
		lock:     sync.Mutex{},
		ctx:      ctx,
		errq:     errq,
		injector: injector,
		selector: selector,
		networks: map[string]BlockEventListener{},
	}

	nets, err := selector.Networks()
	if err != nil {
		return nil, errors.Wrap(errListNetworks, err.Error())
	}
	klog.Infof("Pre-register %d networks", len(nets))
	for _, net := range nets {
		if net.Status != models.Registered {
			klog.V(5).Infof("Skip pre-register network %s which at status %s", net.ID, net.Status)
			continue
		}
		err = l.preRegister(&net)
		if err != nil {
			errq.Send(errors.Wrap(err, ""))
			continue
		}
	}

	return l, nil
}

func (l *listener) preRegister(net *models.Network) error {
	klog.Infof("Pre-register network %s", net.ID)

	n := &network.Network{
		ID:       net.ID,
		Platform: network.Platform(net.Platform),
	}

	var err error
	var blkListener BlockEventListener

	switch net.Type {
	case string(network.FABRIC):
		var fabProfile = new(network.FabProfile)
		err = json.Unmarshal(net.Profile, fabProfile)
		if err != nil {
			return errors.Wrap(errInvalidNetworkProfile, err.Error())
		}
		n.FabProfile = fabProfile
		startBlock, err := l.selector.NetworkStartAt(n.ID)
		if err != nil {
			l.errq.Send(err)
		}
		blkListener, err = newFabEventListener(l.ctx, l.errq, l.injector, n, startBlock)
		if err != nil {
			return err
		}
	default:
		return errNetworkTypeUnknown
	}

	go blkListener.Events()
	l.networks[n.ID] = blkListener

	return nil
}

func (l *listener) Selector() Selector {
	return l.selector
}

func (l *listener) Register(n *network.Network) error {
	l.lock.Lock()
	defer l.lock.Unlock()

	// Use {network}_{channel} to identity a blockchain uniquely
	if n.Type() == network.FABRIC && n.FabProfile.Channel != "" {
		n.ID = fmt.Sprintf("%s_%s", n.ID, n.FabProfile.Channel)
	}

	if _, ok := l.networks[n.ID]; ok {
		return errNetworkAlreadyExists
	}

	var blkListener BlockEventListener
	var err error

	var profile = make([]byte, 0)
	switch n.Type() {
	case network.FABRIC:
		klog.Infof("Registering a new fabric network: %s", n.ID)
		profile, err = json.Marshal(n.FabProfile)
		if err != nil {
			l.errq.Send(err)
			return err
		}
		blkListener, err = newFabEventListener(l.ctx, l.errq, l.injector, n, 0)
	default:
		return errNetworkTypeUnknown
	}

	if err != nil {
		l.errq.Send(err)
		return err
	}

	go blkListener.Events()
	l.networks[n.ID] = blkListener

	if l.injector != nil {
		err = l.injector.InjectNetworks(&models.Network{
			ID:       n.ID,
			Platform: string(n.Platform),
			Type:     string(n.Type()),
			Profile:  profile,
			Status:   models.Registered,
		})
		if err != nil {
			l.errq.Send(err)
			return err
		}
	}

	return nil
}

func (l *listener) Deregister(nid string) error {
	l.lock.Lock()
	defer l.lock.Unlock()
	klog.Infof("Deregistering network: %s", nid)

	// change network's status to `Deregistered`
	if l.injector != nil && l.selector != nil {
		// do stats update
		net, err := l.selector.Network(nid)
		if err != nil {
			l.errq.Send(err)
			return err
		}
		if net.Status == models.Registered {
			net.Status = models.Deregistered
			err = l.injector.InjectNetworks(net)
			if err != nil {
				l.errq.Send(err)
				return err
			}
		}
	}

	blkListener, ok := l.networks[nid]
	if ok {
		blkListener.Close()
		delete(l.networks, nid)
	}

	return nil
}

func (l *listener) Delete(nid string) error {
	l.lock.Lock()
	defer l.lock.Unlock()
	klog.Infof("Deleting network: %s", nid)

	if l.injector != nil {
		err := l.injector.DeleteNetwork(nid)
		if err != nil {
			l.errq.Send(err)
			return err
		}
	}

	blkListener, ok := l.networks[nid]
	if ok {
		blkListener.Close()
		delete(l.networks, nid)
	}

	return nil
}
