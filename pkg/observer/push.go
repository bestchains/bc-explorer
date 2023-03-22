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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bestchains/bc-explorer/pkg/network"

	"k8s.io/klog/v2"
)

type Pusher struct {
	Host         string
	RegisterChan <-chan network.Network
	DeleteChan   <-chan string
}

func NewPusher(host string, register <-chan network.Network, delete <-chan string) *Pusher {
	return &Pusher{
		Host:         host,
		RegisterChan: register,
		DeleteChan:   delete,
	}
}

func (p *Pusher) Run(ctx context.Context) {
	for {
		select {
		case data := <-p.RegisterChan:
			if err := p.Register(data); err != nil {
				klog.ErrorS(err, "register error")
			}
		case name := <-p.DeleteChan:
			if err := p.Delete(name); err != nil {
				klog.ErrorS(err, "deregister error")
			}
		}
	}
}

func (p *Pusher) Register(data network.Network) (err error) {
	url := p.Host + "/network/register"
	jsonStr, _ := json.Marshal(data)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonStr))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (p *Pusher) Delete(name string) (err error) {
	url := p.Host + fmt.Sprintf("/network/%s", name)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
