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
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/IBM-Blockchain/fabric-operator/pkg/generated/clientset/versioned"
	"github.com/IBM-Blockchain/fabric-operator/pkg/generated/informers/externalversions"
	"github.com/IBM-Blockchain/fabric-operator/pkg/generated/listers/core/v1beta1"
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/authentication/authenticatorfactory"
	"k8s.io/apiserver/pkg/authorization/authorizer"
	"k8s.io/apiserver/pkg/authorization/authorizerfactory"
	"k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/server/options"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type KubernetesAuthor struct {
	requestAuthenticator authenticator.Request
	requestAuthorizer    authorizer.Authorizer
	NetworkLister        v1beta1.NetworkLister
	ChannelLister        v1beta1.ChannelLister
	SkipAuthorize        bool
}

var (
	ErrNoPermission = errors.New("no permission")
)

const (
	// TODO valid the external request to the listener
	ListPath       = "/networks"
	RegisterPath   = "/network/register"
	DeregisterPath = "/network/deregister/"
	CommonPath     = "/networks/"
)

func (k *KubernetesAuthor) New(ctx context.Context) (err error) {
	restConfig := config.GetConfigOrDie()
	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	authenticatorConfig := authenticatorfactory.DelegatingAuthenticatorConfig{
		Anonymous:               false, // always require authentication
		CacheTTL:                2 * time.Minute,
		TokenAccessReviewClient: kubeClient.AuthenticationV1(),
		WebhookRetryBackoff:     options.DefaultAuthWebhookRetryBackoff(),
	}
	authenticatorReq, _, err := authenticatorConfig.New()
	if err != nil {
		return err
	}
	k.requestAuthenticator = authenticatorReq

	if !k.SkipAuthorize {
		authorizerConfig := authorizerfactory.DelegatingAuthorizerConfig{
			SubjectAccessReviewClient: kubeClient.AuthorizationV1(),
			AllowCacheTTL:             5 * time.Minute,
			DenyCacheTTL:              30 * time.Second,
			WebhookRetryBackoff:       options.DefaultAuthWebhookRetryBackoff(),
		}
		k.requestAuthorizer, err = authorizerConfig.New()
		if err != nil {
			return fmt.Errorf("failed to create sar authorizer: %w", err)
		}
		k.NetworkLister, k.ChannelLister, err = getListers(ctx)
	}
	return err
}

func (k *KubernetesAuthor) Authorizer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if k.SkipAuthorize {
			next.ServeHTTP(w, req)
			return
		}
		klog.V(5).InfoS("try to get permission")
		u, ok := request.UserFrom(req.Context())
		if !ok {
			http.Error(w, "user not in context", http.StatusBadRequest)
			return
		}
		if u.GetName() == fmt.Sprintf("system:serviceaccount:%s:%s", os.Getenv("POD_NAMESPACE"), os.Getenv("POD_SA")) {
			klog.V(5).InfoS("local observer, skip")
			next.ServeHTTP(w, req)
			return
		}

		networkName, channelName, err := k.GetReqName(req.RequestURI)
		if err != nil {
			klog.V(2).Infof("parse resource get error:%v", err)
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		apiVerb := "*"
		switch req.Method {
		case "POST":
			apiVerb = "create"
		case "GET":
			apiVerb = "get"
		case "PUT":
			apiVerb = "update"
		case "PATCH":
			apiVerb = "patch"
		case "DELETE":
			apiVerb = "delete"
		}

		attrs := []authorizer.AttributesRecord{
			{
				User:            u,
				Verb:            apiVerb,
				APIGroup:        "ibp.com",
				APIVersion:      "v1beta1",
				Resource:        "networks",
				Subresource:     "",
				Name:            networkName,
				ResourceRequest: true,
			},
			{
				User:            u,
				Verb:            apiVerb,
				APIGroup:        "ibp.com",
				APIVersion:      "v1beta1",
				Resource:        "channels",
				Subresource:     "",
				Name:            channelName,
				ResourceRequest: true,
			},
		}
		for _, attr := range attrs {
			authorized, reason, err := k.requestAuthorizer.Authorize(req.Context(), attr)
			msg := fmt.Sprintf("(user=%s, verb=%s, resource=%s, subresource=%s, resourcename=%s)", u.GetName(), attr.GetVerb(), attr.GetResource(), attr.GetSubresource(), attr.GetName())
			if err != nil {
				msg = "Authorization error" + msg
				klog.Errorf("%s: %s", msg, err)
				http.Error(w, msg, http.StatusInternalServerError)
				return
			}
			if authorized != authorizer.DecisionAllow {
				msg = "Forbidden" + msg
				klog.V(2).Infof("%s. Reason: %q.", msg, reason)
				http.Error(w, msg, http.StatusForbidden)
				return
			}
		}
		next.ServeHTTP(w, req)
	})
}

func (k *KubernetesAuthor) Authentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		klog.V(5).InfoS("try to get user")
		res, ok, err := k.requestAuthenticator.AuthenticateRequest(req)
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

func (k *KubernetesAuthor) GetReqName(rawURL string) (network, channelName string, err error) {
	defer func() {
		klog.V(5).Infof("url:%s, get network:%s channelID:%s", rawURL, network, channelName)
	}()
	u, err := url.Parse(rawURL)
	if err != nil {
		klog.ErrorS(err, "parse url get error", "url", rawURL)
		return "", "", err
	}
	if u.Path == ListPath {
		return "", "", ErrNoPermission
	}
	if u.Path == RegisterPath {
		return "", "", ErrNoPermission
	}
	if strings.HasPrefix(u.Path, DeregisterPath) {
		return "", "", ErrNoPermission
	}
	if strings.HasPrefix(u.Path, CommonPath) {
		t := strings.Split(strings.TrimPrefix(u.Path, CommonPath), "/")
		if len(t) == 0 {
			return "", "", fmt.Errorf("wrong uri:%s", u.Path)
		}
		networkNameChannelID := t[0]
		networkName, channelID, found := strings.Cut(networkNameChannelID, "_")
		if !found {
			return "", "", fmt.Errorf("wrong uri:%s", u.Path)
		}
		if _, err := k.NetworkLister.Get(networkName); err != nil {
			return "", "", err
		}
		list, err := k.ChannelLister.List(labels.Everything())
		if err != nil {
			return "", "", err
		}
		for _, ch := range list {
			if ch.GetChannelID() == channelID && ch.Spec.Network == networkName {
				return networkName, ch.GetName(), nil
			}
		}
		return "", "", fmt.Errorf("from channelID:%s cant find channel", channelID)
	}
	return
}

func (k *KubernetesAuthor) Run() fiber.Handler {
	return adaptor.HTTPMiddleware(func(handler http.Handler) http.Handler {
		handler = k.Authorizer(handler)
		return k.Authentication(handler)
	})
}

func getListers(ctx context.Context) (networkLister v1beta1.NetworkLister, channelLister v1beta1.ChannelLister, err error) {
	restConfig := config.GetConfigOrDie()
	var vclient *versioned.Clientset
	vclient, err = versioned.NewForConfig(restConfig)
	if err != nil {
		return
	}

	informerFactory := externalversions.NewSharedInformerFactory(vclient, 0)
	channelInformer := informerFactory.Ibp().V1beta1().Channels()
	channelLister = channelInformer.Lister()
	networkInformer := informerFactory.Ibp().V1beta1().Networks()
	networkLister = networkInformer.Lister()
	informerFactory.Start(ctx.Done())
	if !cache.WaitForNamedCacheSync("auth", ctx.Done(), channelInformer.Informer().HasSynced, networkInformer.Informer().HasSynced) {
		err = fmt.Errorf("waitForCacheSync failed")
		klog.ErrorS(err, "cannot sync caches")
		return
	}
	return
}
