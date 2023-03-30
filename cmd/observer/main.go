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
	"context"
	"flag"

	"github.com/bestchains/bc-explorer/pkg/observer"
	"k8s.io/apiserver/pkg/server"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	host              = flag.String("host", "localhost:9999", "the host of listener")
	operatorNamespace = flag.String("operator-namespace", "baas-system", "the ns of fabric-operator")
)

func main() {
	logf.SetLogger(zap.New(zap.UseDevMode(true)))
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	flag.Parse()
	flag.VisitAll(func(flag *flag.Flag) {
		klog.V(1).Infof("FLAG: --%s=%q", flag.Name, flag.Value)
	})

	if err := run(); err != nil {
		klog.Error(err)
	}
}

func run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		stopCh := server.SetupSignalHandler()
		<-stopCh
		cancel()
	}()
	restConfig := config.GetConfigOrDie()
	if err := observer.Run(ctx, restConfig, *host, *operatorNamespace); err != nil {
		return err
	}
	return nil
}
