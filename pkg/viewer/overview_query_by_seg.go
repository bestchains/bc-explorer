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
	"sync"

	"github.com/bestchains/bc-explorer/pkg/models"
	"github.com/go-pg/pg/v10"
	"k8s.io/klog/v2"
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
	result := make([]BySegResp, number+1)
	ch := make(chan error, number+1)
	var wg sync.WaitGroup
	s, e := start, from
	for i := 0; i <= int(number); i++ {
		result[i] = BySegResp{Start: s, End: e, Count: 0}
		wg.Add(1)
		go func(i int, s, e int64) {
			defer wg.Done()
			if err := db.Model((*models.Block)(nil)).Where(`"network"=?`, network).Where(`"createdAt">=?`, s).Where(`"createdAt"<=?`, e).
				ColumnExpr(`count(*) as count`).Select(&result[i].Count); err != nil {
				ch <- err
				klog.Error(err)
				return
			}
		}(i, s, e)
		s, e = e, e+interval
	}

	wg.Wait()
	if len(ch) > 0 {
		return nil, <-ch
	}
	return result, nil
}

func QueryTrnasactions(db *pg.DB, network string, from, interval, number int64) ([]BySegResp, error) {
	start := from - interval
	result := make([]BySegResp, number+1)
	ch := make(chan error, number+1)
	var wg sync.WaitGroup
	s, e := start, from
	for i := 0; i <= int(number); i++ {
		result[i] = BySegResp{Start: s, End: e, Count: 0}
		wg.Add(1)
		go func(i int, s, e int64) {
			defer wg.Done()
			if err := db.Model((*models.Transaction)(nil)).Where(`"network"=?`, network).Where(`"createdAt">=?`, s).Where(`"createdAt"<=?`, e).
				ColumnExpr(`count(*) as count`).Select(&result[i].Count); err != nil {
				ch <- err
				klog.Error(err)
				return
			}
		}(i, s, e)
		s, e = e, e+interval
	}

	wg.Wait()
	if len(ch) > 0 {
		return nil, <-ch
	}
	return result, nil
}
