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
	"k8s.io/klog/v2"
)

type BlockArg struct {
	From, Size         int
	Network            string
	StartTime, EndTime int64
	BlockNumber        uint64
	BlockHash          string
}

func (ba *BlockArg) ToCond() ([]string, []interface{}) {
	params := make([]interface{}, 0)
	cond := make([]string, 0)

	if ba.Network != "" {
		cond = append(cond, ` network=? `)
		params = append(params, ba.Network)
	}

	if ba.StartTime > 0 {
		cond = append(cond, `"createdAt">=?`)
		params = append(params, ba.StartTime)
	}
	if ba.EndTime > 0 {
		cond = append(cond, `"createdAt"<=?`)
		params = append(params, ba.EndTime)
	}
	if ba.BlockNumber != 0 {
		cond = append(cond, `"blockNumber"=?`)
		params = append(params, ba.BlockNumber)
	}
	if ba.BlockHash != "" {
		cond = append(cond, `"blockHash"=?`)
		params = append(params, ba.BlockHash)
	}
	return cond, params
}

type Block interface {
	List(BlockArg) ([]models.Block, int64, error)
	Get(BlockArg) (models.Block, error)
}
type blockHandler struct {
	db *pg.DB
}

func NewBlockHandler(db *pg.DB) Block {
	return &blockHandler{db: db}
}

func (bh *blockHandler) List(arg BlockArg) ([]models.Block, int64, error) {
	if arg.Network == "" {
		return nil, 0, fmt.Errorf("network name can't be empty")
	}

	result := make([]models.Block, 0)
	query, params := arg.ToCond()
	klog.V(5).Infof(" list query %s\n", query)

	q := bh.db.Model(&result)
	for i := 0; i < len(query); i++ {
		q = q.Where(query[i], params[i])
	}

	c, err := q.Count()
	if err != nil {
		return result, 0, err
	}
	q = q.Order(`createdAt desc`)
	if arg.Size != 0 {
		q = q.Limit(arg.Size).Offset(arg.From)
	}
	if err = q.Select(); err != nil {
		return result, 0, err
	}
	return result, int64(c), nil
}

func (bh *blockHandler) Get(arg BlockArg) (models.Block, error) {
	if arg.BlockHash == "" {
		return models.Block{}, fmt.Errorf("blockHash can't be empty")
	}
	query, params := arg.ToCond()
	var result models.Block
	q := bh.db.Model(&result)
	for i := 0; i < len(query); i++ {
		q = q.Where(query[i], params[i])
	}
	err := q.Select()
	return result, err
}
