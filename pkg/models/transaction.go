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

import (
	"context"

	"github.com/go-pg/pg/v10"
	"k8s.io/klog/v2"
)

type TxType string

const TransactionTableName = "transactions"

const (
	Config              TxType = "Config"
	ConfigUpdate        TxType = "ConfigUpdate"
	EndorserTransaction TxType = "EndorserTransaction"
)

type Transaction struct {
	ID          string `pg:"id,pk" json:"id"`
	Network     string `pg:"network" json:"network"`
	BlockNumber uint64 `pg:"blockNumber" json:"blockNumber"`
	CreatedAt   int64  `pg:"createdAt" json:"createdAt"`
	Creator     string `pg:"creator" json:"creator"`

	Type    TxType `pg:"type" json:"type"`
	Payload []byte `pg:"payload" json:"payload"`

	// EndorserTransaction
	ChaincodeID string   `pg:"chaincodeId" json:"chaincodeId"`
	Method      string   `pg:"method" json:"method"`
	Args        []string `pg:"args" json:"args"`

	ValidationCode int32 `pg:"validationCode" json:"validationCode"`
}

var _ pg.QueryHook = (*Transaction)(nil)

func (*Transaction) BeforeQuery(ctx context.Context, event *pg.QueryEvent) (context.Context, error) {
	query, err := event.FormattedQuery()
	if err != nil {
		return ctx, err
	}
	klog.V(5).Infof("[format query] %s\n", string(query))
	return ctx, nil
}

func (*Transaction) AfterQuery(ctx context.Context, event *pg.QueryEvent) error {
	return nil
}
