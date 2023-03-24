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

type TxType string

const TransactionTableName = "transactions"

const (
	Config              TxType = "Config"
	ConfigUpdate        TxType = "ConfigUpdate"
	EndorserTransaction TxType = "EndorserTransaction"
)

type Transaction struct {
	ID          string `pg:"id,pk"`
	Network     string `pg:"network"`
	BlockNumber uint64 `pg:"blockNumber"`
	CreatedAt   int64  `pg:"createdAt"`
	Creator     string `pg:"creator"`

	Type    TxType `pg:"type"`
	Payload []byte `pg:"payload"`

	// EndorserTransaction
	ChaincodeID string   `pg:"chaincodeId"`
	Method      string   `pg:"method"`
	Args        []string `pg:"args"`

	ValidationCode int32 `pg:"validationCode"`
}
