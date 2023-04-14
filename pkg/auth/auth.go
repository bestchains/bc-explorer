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

package auth

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"
	"k8s.io/klog/v2"
)

// New creates a new middleware handler
func New(ctx context.Context, config Config) fiber.Handler {
	var a auth
	switch strings.ToLower(config.AuthMethod) {
	case "none":
		a = &NoneAuthor{}
	case "oidc":
		a = &OIDCAuthor{
			KubernetesAuthor: &KubernetesAuthor{},
		}
	default:
		a = &KubernetesAuthor{}
	}
	if err := a.New(ctx); err != nil {
		// If auth init fails, should panic as soon as possible.
		panic(err)
	}
	klog.Infoln("auth init success", "authMethod", config.AuthMethod)
	return a.Run()
}

type auth interface {
	New(ctx context.Context) error
	Run() fiber.Handler
}
