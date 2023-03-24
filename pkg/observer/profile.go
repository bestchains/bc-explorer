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

// copy from https://github.com/bestchains/fabric-operator/blob/d717b9e2df3319aaeaa5d804afec515ad8b948d3/pkg/connector/profile.go#L38-L138

// Profile contasins all we need to connect with a blockchain network. Currently we use embedded pem by default
// +k8s:deepcopy-gen=true
type Profile struct {
	Version       string `yaml:"version,omitempty" json:"version,omitempty"`
	Client        `yaml:"client,omitempty" json:"client,omitempty"`
	Channels      map[string]ChannelInfo      `yaml:"channels" json:"channels"`
	Organizations map[string]OrganizationInfo `yaml:"organizations,omitempty" json:"organizations,omitempty"`
	// Orderers defines all orderer endpoints which can be used
	Orderers map[string]NodeEndpoint `yaml:"orderers,omitempty" json:"orderers,omitempty"`
	// Peers defines all peer endpoints which can be used
	Peers map[string]NodeEndpoint `yaml:"peers,omitempty" json:"peers,omitempty"`
}

// Client defines who is trying to connect with networks
type Client struct {
	Organization string `yaml:"organization,omitempty" json:"organization,omitempty"`
	Logging      `yaml:"logging,omitempty" json:"logging,omitempty"`
	// For blockchain explorer
	AdminCredential `yaml:"adminCredential,omitempty" json:"adminCredential,omitempty"`
	CredentialStore `yaml:"credentialStore,omitempty" json:"credentialStore,omitempty"`
	TLSEnable       bool `yaml:"tlsEnable,omitempty" json:"tlsEnable,omitempty"`
}

// +k8s:deepcopy-gen=true
type Logging struct {
	Level string `yaml:"level,omitempty" json:"level,omitempty"`
}

// +k8s:deepcopy-gen=true
type CryptoConfig struct {
	Path string `yaml:"path,omitempty" json:"path,omitempty"`
}

// +k8s:deepcopy-gen=true
type AdminCredential struct {
	ID       string `yaml:"id,omitempty" json:"id,omitempty"`
	Password string `yaml:"password,omitempty" json:"password" default:"passw0rd"`
}

// +k8s:deepcopy-gen=true
type CredentialStore struct {
	Path        string `yaml:"path,omitempty" json:"path,omitempty"`
	CryptoStore `yaml:"cryptoStore,omitempty" json:"cryptoStore,omitempty"`
}

// +k8s:deepcopy-gen=true
type CryptoStore struct {
	Path string `yaml:"path,omitempty" json:"path,omitempty"`
}

// ChannelInfo defines configurations when connect to this channel
// +k8s:deepcopy-gen=true
type ChannelInfo struct {
	// Peers which can be used to connect to this channel
	Peers map[string]PeerInfo `yaml:"peers" json:"peers"`
}

// +k8s:deepcopy-gen=true
type PeerInfo struct {
	EndorsingPeer  *bool `yaml:"endorsingPeer,omitempty" json:"endorsingPeer,omitempty"`
	ChaincodeQuery *bool `yaml:"chaincodeQuery,omitempty" json:"chaincodeQuery,omitempty"`
	LedgerQuery    *bool `yaml:"ledgerQuery,omitempty" json:"ledgerQuery,omitempty"`
	EventSource    *bool `yaml:"eventSource,omitempty" json:"eventSource,omitempty"`
}

// OrganizationInfo defines a organization along with its users and peers
// +k8s:deepcopy-gen=true
type OrganizationInfo struct {
	MSPID string          `yaml:"mspid,omitempty" json:"mspid,omitempty"`
	Users map[string]User `yaml:"users,omitempty" json:"users,omitempty"`
	Peers []string        `yaml:"peers,omitempty" json:"peers,omitempty"`

	// For blockchain explorer
	AdminPrivateKey Pem `yaml:"adminPrivateKey,omitempty" json:"adminPrivateKey,omitempty"`
	SignedCert      Pem `yaml:"signedCert,omitempty" json:"signedCert,omitempty"`
}

// User is the ca identity which has a private key(embedded pem) and signed certificate(embedded pem)
// +k8s:deepcopy-gen=true
type User struct {
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	Key  Pem    `yaml:"key,omitempty" json:"key,omitempty"`
	Cert Pem    `yaml:"cert,omitempty" json:"cert,omitempty"`
}

// +k8s:deepcopy-gen=true
type Pem struct {
	Pem string `yaml:"pem,omitempty" json:"pem,omitempty"`
}

// +k8s:deepcopy-gen=true
type NodeEndpoint struct {
	URL        string `yaml:"url,omitempty" json:"url,omitempty"`
	TLSCACerts `yaml:"tlsCACerts,omitempty" json:"tlsCACerts,omitempty"`
}

// +k8s:deepcopy-gen=true
type TLSCACerts struct {
	Path string `yaml:"path,omitempty" json:"path,omitempty"`
	Pem  string `yaml:"pem,omitempty" json:"pem,omitempty"`
}
