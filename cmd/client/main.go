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

package main

import (
	"encoding/json"
	"flag"
	"os"

	"github.com/bestchains/bc-explorer/pkg/network"

	gwclient "github.com/hyperledger/fabric-gateway/pkg/client"

	"k8s.io/klog/v2"

	"google.golang.org/grpc/status"
)

var (
	profile  = flag.String("profile", "./network.json", "profile to connect with blockchain network")
	contract = flag.String("contract", "samplecc", "contract name")
	method   = flag.String("method", "PutValue", "contract method")
	args     = new(sliceArgs)
)

func main() {
	flag.Var(args, "args", "a list of arguments for contract call")
	flag.Parse()

	raw, err := os.ReadFile(*profile)
	if err != nil {
		panic(err)
	}
	profile := &network.Network{}
	err = json.Unmarshal(raw, profile)
	if err != nil {
		panic(err)
	}
	client, err := network.NewFabricClient(profile)
	if err != nil {
		panic(err)
	}
	contract := client.Channel("").GetContract(*contract)
	resp, err := contract.SubmitTransaction(*method, *args...)
	if err != nil {
		switch err := err.(type) {
		case *gwclient.EndorseError, *gwclient.SubmitError, *gwclient.CommitStatusError:
			s := status.Convert(err)
			klog.Errorf("StatusCode: %d Details: %v \n", s.Code(), s.Details())
		case *gwclient.CommitError:
			klog.Errorf("TxId: %s ValidationCode: %d Message: %s \n", err.TransactionID, err.Code, err.Error())
		default:
			klog.Error(err)
		}
		return
	}
	klog.Infof("Result: %s", resp)
}
