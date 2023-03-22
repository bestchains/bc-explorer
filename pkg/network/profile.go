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
	"crypto/x509"
	"encoding/pem"
	"net/url"

	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"k8s.io/klog/v2"
)

var (
	errInvalidFabNetEndpoint = errors.New("fabric network's peer endpoint is invalid")
	errInvalidCert           = errors.New("invalid certificate")
	errInvalidPrivateKey     = errors.New("invalid private key")
)

type FabProfile struct {
	Organization string       `yaml:"organization" json:"organization" validate:"required"`
	User         User         `yaml:"user" json:"user" validate:"required"`
	Enpoint      NodeEndpoint `yaml:"endpoint" json:"endpoint" validate:"required"`
}

type User struct {
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	Key  Pem    `yaml:"key,omitempty" json:"key,omitempty"`
	Cert Pem    `yaml:"cert,omitempty" json:"cert,omitempty"`
}

type Pem struct {
	Pem string `yaml:"pem,omitempty" json:"pem,omitempty"`
}

type NodeEndpoint struct {
	URL        string `yaml:"url,omitempty" json:"url,omitempty"`
	TLSCACerts `yaml:"tlsCACerts,omitempty" json:"tlsCACerts,omitempty"`
}

type TLSCACerts struct {
	Pem string `yaml:"pem,omitempty" json:"pem,omitempty"`
}

func newFabClientConn(p *FabProfile, channel string) (*grpc.ClientConn, error) {
	u, err := url.Parse(p.Enpoint.URL)
	if err != nil {
		return nil, errors.Wrap(errInvalidFabNetEndpoint, err.Error())
	}

	// initialize transport credentials
	transportCreds := insecure.NewCredentials()
	if u.Scheme == "grpcs" {
		klog.Infof("ssl enabled in endpoint: %s", u.Host)
		cpb, _ := pem.Decode([]byte(p.Enpoint.TLSCACerts.Pem))
		cert, err := x509.ParseCertificate(cpb.Bytes)
		if err != nil {
			return nil, errors.Wrap(errInvalidX509Cert, err.Error())
		}
		pool := x509.NewCertPool()
		pool.AddCert(cert)
		transportCreds = credentials.NewClientTLSFromCert(pool, "")
	}

	return grpc.Dial(u.Host, grpc.WithTransportCredentials(transportCreds))
}

func (u User) ToIdentity(org string) (identity.Identity, identity.Sign, error) {
	crt, err := identity.CertificateFromPEM([]byte(u.Cert.Pem))
	if err != nil {
		return nil, nil, errors.Wrap(err, errInvalidCert.Error())
	}
	id, err := identity.NewX509Identity(org, crt)
	if err != nil {
		return nil, nil, errors.Wrap(err, errInvalidCert.Error())
	}
	priv, err := identity.PrivateKeyFromPEM([]byte(u.Key.Pem))
	if err != nil {
		return nil, nil, errors.Wrap(err, errInvalidPrivateKey.Error())
	}
	sign, err := identity.NewPrivateKeySign(priv)
	if err != nil {
		return nil, nil, errors.Wrap(err, errInvalidPrivateKey.Error())
	}
	return id, sign, nil
}
