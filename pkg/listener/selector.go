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

package listener

import (
	"fmt"
	"strings"

	"github.com/bjwswang/bc-explorer/pkg/models"
	"github.com/go-pg/pg/v10"
)

type Selector interface {
	Networks(fields ...string) ([]models.Network, error)
	Network(nid string) (*models.Network, error)
	NetworkStartAt(nid string) (uint64, error)
}

// pqSelector used to select data into postgreSQL
type pqSelector struct {
	db *pg.DB
}

func NewPQSelector(db *pg.DB) (Selector, error) {
	return &pqSelector{
		db: db,
	}, nil
}

func (pqstr *pqSelector) Networks(fields ...string) ([]models.Network, error) {
	var nets []models.Network
	var fieldStr = "*"
	if len(fields) > 0 {
		fieldStr = strings.Join(fields, ",")
	}
	_, err := pqstr.db.Query(&nets, fmt.Sprintf("select %s from networks", fieldStr))
	if err != nil {
		return nil, err
	}
	return nets, err
}

func (pqstr *pqSelector) Network(nid string) (*models.Network, error) {
	var net = new(models.Network)
	_, err := pqstr.db.QueryOne(net, `select * from networks where "id" = ?;`, nid)
	if err != nil {
		return nil, err
	}
	return net, err
}

// select max block number in this network
func (pqstr *pqSelector) NetworkStartAt(nid string) (uint64, error) {
	var lastBlock models.Block
	_, err := pqstr.db.QueryOne(
		&lastBlock, `
		select * from blocks
		where 
			"network" = ? 
		AND 
			"blockNumber" = (select MAX("blockNumber") from blocks where "network" = ?);`, nid, nid)
	if err != nil {
		return 0, err
	}
	return lastBlock.BlockNumber, nil
}
