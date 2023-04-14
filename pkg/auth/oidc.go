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
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"k8s.io/apiserver/pkg/authentication/request/bearertoken"
	"k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/server/dynamiccertificates"
	"k8s.io/apiserver/plugin/pkg/authenticator/token/oidc"
	"k8s.io/klog/v2"
)

type OIDCAuthor struct {
	*KubernetesAuthor
}

func (o *OIDCAuthor) New(ctx context.Context) (err error) {
	sarAuthorizer, err := getRequestAuthorizer()
	if err != nil {
		return fmt.Errorf("failed to create sar authorizer: %w", err)
	}
	o.requestAuthorizer = sarAuthorizer

	fileName := os.Getenv("OIDC_CA_FILE")
	if fileName == "" {
		return errors.New("no ca file")
	}
	pemBlock, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}
	provider, err := dynamiccertificates.NewStaticCAContent("ca", pemBlock)
	if err != nil {
		return err
	}
	tokenAuthenticator, err := oidc.New(oidc.Options{
		CAContentProvider: provider,
		IssuerURL:         os.Getenv("OIDC_ISSUER_URL"),
		ClientID:          os.Getenv("OIDC_CLIENT_ID"),
		UsernameClaim:     os.Getenv("OIDC_USERNAME_CLAIM"),
		GroupsClaim:       os.Getenv("OIDC_GROUPS_CLAIM"),
	})
	if err != nil {
		return err
	}
	o.requestAuthenticator = bearertoken.New(tokenAuthenticator)
	o.NetworkLister, o.ChannelLister, err = getListers(ctx)
	return err
}

func (o *OIDCAuthor) Authentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		klog.V(5).InfoS("try to get user")
		res, ok, err := o.requestAuthenticator.AuthenticateRequest(req)
		if err != nil {
			klog.Errorf("Unable to authenticate the request due to an error: %v", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		klog.V(5).InfoS("get user in req", "user", res.User)
		req = req.WithContext(request.WithUser(req.Context(), res.User))
		next.ServeHTTP(w, req)
	})
}

func (o *OIDCAuthor) Run() fiber.Handler {
	return adaptor.HTTPMiddleware(func(handler http.Handler) http.Handler {
		handler = o.Authorizer(handler)
		return o.Authentication(handler)
	})
}
