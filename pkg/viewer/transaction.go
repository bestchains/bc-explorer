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
	"github.com/go-pg/pg/v10"
	"k8s.io/klog/v2"

	"github.com/bestchains/bc-explorer/pkg/models"
)

type TransArg struct {
	From, Size         int
	NetworkName        string
	StartTime, EndTime int64
	Hash               string
	BlockNum           uint64
}

func (ta *TransArg) ToCond() ([]string, []interface{}) {
	params := make([]interface{}, 0)
	cond := make([]string, 0)

	if ta.NetworkName != "" {
		cond = append(cond, ` network = ?`)
		params = append(params, ta.NetworkName)
	}
	if ta.StartTime > 0 {
		cond = append(cond, `"createdAt">=?`)
		params = append(params, ta.StartTime)
	}
	if ta.EndTime > 0 {
		cond = append(cond, `"createdAt"<=?`)
		params = append(params, ta.EndTime)
	}
	if ta.Hash != "" {
		cond = append(cond, ` id = ?`)
		params = append(params, ta.Hash)
	}
	if ta.BlockNum > 0 {
		cond = append(cond, ` "blockNumber"=?`)
		params = append(params, ta.BlockNum)
	}

	return cond, params
}

type Transaction interface {
	// List : query transactions
	List(ta TransArg) ([]models.Transaction, int64, error)

	// Get : query transaction by transaction hash
	Get(ta TransArg) (*models.Transaction, error)
}

type TxHandler struct {
	db *pg.DB
}

func NewTxHandler(db *pg.DB) Transaction {
	return &TxHandler{db: db}
}

func (t *TxHandler) List(ta TransArg) ([]models.Transaction, int64, error) {

	var hd TxHandler
	if ta.NetworkName == "" {
		return nil, 0, fmt.Errorf("network name can't be empty")
	}

	txs := make([]models.Transaction, 0)
	query, params := ta.ToCond()
	klog.V(5).Infof(" list query %s\n", query)

	q := hd.db.Model(&txs)
	for i := 0; i < len(query); i++ {
		q = q.Where(query[i], params[i])
	}

	c, err := q.Count()
	if err != nil {
		return txs, 0, err
	}
	q = q.Order(`createdAt desc`)
	if ta.Size != 0 {
		q = q.Limit(ta.Size).Offset(ta.From)
	}

	if err = q.Select(); err != nil {
		return txs, 0, err
	}
	return txs, int64(c), nil
}

func (t *TxHandler) Get(ta TransArg) (*models.Transaction, error) {
	var tx = new(models.Transaction)
	_, err := t.db.QueryOne(&tx, `select * from transactions where "id" = ?`, ta.Hash)
	if err != nil {
		return nil, err
	}
	return tx, err
}
