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

const BlockTableName = "blocks"

type Block struct {
	BlockHash         string `pg:"blockHash,pk" json:"blockHash"`
	Network           string `pg:"network" json:"network"`
	BlockNumber       uint64 `pg:"blockNumber" json:"blockNumber"`
	PrevioudBlockHash string `pg:"preBlockHash" json:"preBlockHash"`
	DataHash          string `pg:"dataHash" json:"dataHash"`
	CreatedAt         int64  `pg:"createdAt" json:"createdAt"`
	BlockSize         int    `pg:"blockSize" json:"blockSize"`
	TxCount           int    `pg:"txCount" json:"txCount"`
}

var _ pg.QueryHook = (*Block)(nil)

func (*Block) BeforeQuery(ctx context.Context, event *pg.QueryEvent) (context.Context, error) {
	query, err := event.FormattedQuery()
	if err != nil {
		return ctx, nil
	}
	klog.V(5).Infof("[format query] %s\n", string(query))
	return ctx, nil
}

func (*Block) AfterQuery(ctx context.Context, event *pg.QueryEvent) error {
	return nil
}
