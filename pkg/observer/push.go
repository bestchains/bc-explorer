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
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/bestchains/bc-explorer/pkg/network"

	"k8s.io/klog/v2"
)

type Pusher struct {
	Host string
	Msg  <-chan Msg
}

func NewPusher(host string, msg <-chan Msg) *Pusher {
	return &Pusher{
		Host: strings.TrimSuffix(host, "/"),
		Msg:  msg,
	}
}

func (p *Pusher) Run(ctx context.Context) {
	for {
		data := <-p.Msg
		key := key(data.NetworkName, data.ChannelName)
		for i := 0; i < 2; i++ {
			// If the HTTP request fails, try again after 1 second.
			time.Sleep(time.Duration(i) * time.Second)
			switch data.Type {
			case Register:
				if err := p.Register(key, data.Data); err != nil {
					klog.ErrorS(err, "register error", "key", key)
					continue
				}
				klog.InfoS("register done.", "key", key)
			case Delete:
				if err := p.Delete(key); err != nil {
					klog.ErrorS(err, "delete error", "key", key)
					continue
				}
				klog.InfoS("delete done.", "key", key)
			case Deregister:
				if err := p.DeRegister(key); err != nil {
					klog.ErrorS(err, "deregister error", "key", key)
					continue
				}
				klog.InfoS("deregister done.", "key", key)
			}
			break
		}
	}
}

func (p *Pusher) Register(key string, data *network.Network) (err error) {
	url := p.Host + "/network/register"
	jsonStr, _ := json.Marshal(data)
	klog.V(5).InfoS(fmt.Sprintf("register url:%s post body:%s", url, string(jsonStr)), "key", key)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonStr))
	if err != nil {
		return err
	}
	return p.getResp(key, resp)
}

func (p *Pusher) Delete(name string) (err error) {
	url := p.Host + fmt.Sprintf("/network/%s", name)
	klog.V(5).InfoS(fmt.Sprintf("delete url:%s", url), "key", name)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	return p.getResp(name, resp)
}

func (p *Pusher) DeRegister(name string) (err error) {
	url := p.Host + fmt.Sprintf("/network/deregister/%s", name)
	klog.V(5).InfoS(fmt.Sprintf("deregister url:%s", url), "key", name)
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return err
	}
	return p.getResp(name, resp)
}

func (p *Pusher) getResp(key string, resp *http.Response) error {
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyString := string(bodyBytes)
	klog.V(5).InfoS(fmt.Sprintf("resp code:%d body:%s", resp.StatusCode, bodyString), "key", key)
	return fmt.Errorf("code: [%d] %s, resp:%s", resp.StatusCode, http.StatusText(resp.StatusCode), bodyString)
}
