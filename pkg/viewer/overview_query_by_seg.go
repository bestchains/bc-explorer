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
	"github.com/bestchains/bc-explorer/pkg/models"
	"github.com/go-pg/pg/v10"
)

type BySegFunc func(*pg.DB, string, int64, int64, int64) ([]BySegResp, error)

var bySegFuncs = map[string]BySegFunc{
	BlockAggregation:       QueryBlocks,
	TransactionAggregation: QueryTrnasactions,
}

// QueryByBlocks
// from=0,interval=5,number=2
// [-5,0),[0-5), [5-10)
func QueryBlocks(db *pg.DB, network string, from, interval, number int64) ([]BySegResp, error) {
	start := from - interval
	end := from + number*interval
	result := make([]BySegResp, number+1)
	blocks := make([]models.Block, 0)
	if err := db.Model(&blocks).Where(`"network"=?`, network).
		Where(`"createdAt">=?`, start).Where(`"createdAt"<=?`, end).Order("createdAt asc").Select(); err != nil {
		return result, err
	}
	s, e := start, from
	for i := 0; i <= int(number); i++ {
		result[i] = BySegResp{Start: s, End: e, Count: 0}
		s, e = e, e+interval
	}

	end = from
	blockCount := int64(0)
	index := 0
	for _, block := range blocks {
		if block.CreatedAt < end {
			blockCount++
			continue
		}
		result[index].Count = blockCount
		blockCount = 1
		end, index = end+interval, index+1
	}
	result[index].Count = blockCount
	return result, nil
}

func QueryTrnasactions(db *pg.DB, network string, from, interval, number int64) ([]BySegResp, error) {
	start := from - interval
	end := from + number*interval
	result := make([]BySegResp, number+1)
	transactions := make([]models.Transaction, 0)
	if err := db.Model(&transactions).Where(`"network"=?`, network).
		Where(`"createdAt">=?`, start).Where(`"createdAt"<=?`, end).Order("createdAt asc").Select(); err != nil {
		return result, err
	}
	s, e := start, from
	for i := 0; i <= int(number); i++ {
		result[i] = BySegResp{Start: s, End: e, Count: 0}
		s, e = e, e+interval
	}

	end = from
	txCount, index := int64(0), 0
	for _, trans := range transactions {
		if trans.CreatedAt < end {
			txCount++
			continue
		}
		result[index].Count = txCount
		txCount = 1
		end, index = end+interval, index+1
	}
	result[index].Count = txCount
	return result, nil
}
