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

package models

type Status string

const (
	Registered   Status = "Registered"
	Deregistered Status = "Deregistered"
)

type Network struct {
	ID       string `pg:"id,pk" json:"id"`
	Type     string `pg:"type" json:"type"`
	Platform string `pg:"platform" json:"platform"`
	Profile  []byte `pg:"profile" json:"profile,omitempty"`
	Status   Status `pg:"status" json:"status,omitempty"`
}
