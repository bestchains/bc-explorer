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

package viewer

import (
	"fmt"

	"github.com/bestchains/bc-explorer/pkg/models"
	"github.com/go-pg/pg/v10"
)

const (
	BlockAggregation       = "blocks"
	TransactionAggregation = "transactions"
)

type SummaryResp struct {
	BlockNumber uint64 `pg:"blockNumber" json:"blockNumber"`
	TxCount     uint64 `pg:"txCount" json:"txCount"`
}

type BySegResp struct {
	Start int64 `json:"start"`
	End   int64 `json:"end"`
	Count int64 `json:"count"`
}

type Overview interface {
	// Summary returns block height, number of transactions, number of nodes, total number of contracts.
	Summary(string) (SummaryResp, error)

	// QueryBySeg query the total number of transactions or blocks for a number of time periods
	// from, interval,number of time periods
	QueryBySeg(int64, int64, int64, string, string) ([]BySegResp, error)
}

type overview struct {
	db *pg.DB
}

func NewOverview(db *pg.DB) Overview {
	return &overview{db: db}
}

func (o *overview) Summary(network string) (SummaryResp, error) {
	var resp SummaryResp
	if err := o.db.Model((*models.Transaction)(nil)).Where(`"network"=?`, network).
		ColumnExpr(`count(*) as "txCount"`).Select(&resp.TxCount); err != nil {
		return resp, err
	}
	if err := o.db.Model((*models.Block)(nil)).Where(`"network"=?`, network).
		ColumnExpr(`max("blockNumber") as "blockNumber"`).Select(&resp.BlockNumber); err != nil {
		return resp, err
	}
	return resp, nil
}

func (o *overview) QueryBySeg(from, interval, number int64, which, network string) ([]BySegResp, error) {
	// If there are more types, cast them into interface implementations
	f, ok := bySegFuncs[which]
	if !ok {
		return nil, fmt.Errorf("not support type %s", which)
	}
	return f(o.db, network, from, interval, number)
}
