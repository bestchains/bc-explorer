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
	"github.com/bestchains/bc-explorer/pkg/models"
	"github.com/go-pg/pg/v10"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
)

type Injector interface {
	InjectNetworks(...*models.Network) error
	InjectBlocks(...*models.Block) error
	InjectTransactions(...*models.Transaction) error
	DeleteNetwork(string) error
}

var _ Injector = new(logInjector)

type logInjector struct {
	logger func(args ...interface{})
}

func NewLogInjector(logger func(args ...interface{})) Injector {
	return &logInjector{
		logger: logger,
	}
}

func (litr *logInjector) InjectNetworks(nets ...*models.Network) error {
	for _, net := range nets {
		litr.logger("Inject network:%s platform:%s type:%s", net.ID, net.Platform, net.Type)
	}
	return nil
}
func (litr *logInjector) DeleteNetwork(nid string) error {
	litr.logger("Delete network:%s", nid)
	return nil
}

func (litr *logInjector) InjectBlocks(blks ...*models.Block) error {
	for _, blk := range blks {
		litr.logger("Inject block:%d network:%s", blk.BlockNumber, blk.Network)
	}
	return nil
}

func (litr *logInjector) InjectTransactions(txs ...*models.Transaction) error {
	for _, tx := range txs {
		litr.logger("Inject tx:%s network:%s block:%d", tx.ID, tx.Network, tx.BlockNumber)
	}
	return nil
}

func NewPQInjector(db *pg.DB) (Injector, error) {
	if err := models.Init(db); err != nil {
		return nil, err
	}
	return &pqInjector{
		db: db,
	}, nil
}

var _ Injector = new(pqInjector)

// pqInjector used to inject data into postgreSQL
type pqInjector struct {
	db *pg.DB
}

func (pqitr *pqInjector) InjectNetworks(nets ...*models.Network) error {
	for _, net := range nets {
		klog.V(5).Infof("PQInjector: inject network %s", net.ID)
		_, err := pqitr.db.Model(net).OnConflict("(id) DO UPDATE").Set("status = EXCLUDED.status").Insert()
		if err != nil {
			return err
		}
	}
	return nil
}

func (pqitr *pqInjector) DeleteNetwork(nid string) error {
	klog.Infof("PQInjector: delete network %s", nid)
	net := &models.Network{
		ID: nid,
	}
	// delete network
	_, err := pqitr.db.Model(net).WherePK().ForceDelete()
	if err != nil {
		return errors.Wrap(err, "delete network")
	}
	// delete all blocks
	_, err = pqitr.db.Model(&models.Block{}).Where(`"network" = ?`, nid).Delete()
	if err != nil {
		return errors.Wrap(err, "delete network's blocks")
	}
	// delete all txs
	_, err = pqitr.db.Model(&models.Transaction{}).Where(`"network" = ?`, nid).Delete()
	if err != nil {
		return errors.Wrap(err, "delete network's transactions")
	}
	return nil
}

func (pqitr *pqInjector) InjectBlocks(blks ...*models.Block) error {
	for _, blk := range blks {
		klog.V(5).Infof("PQInjector: inject block %d %s", blk.BlockNumber, blk.BlockHash)
		_, err := pqitr.db.Model(blk).Insert()
		if err != nil {
			return err
		}
	}
	return nil
}

func (pqitr *pqInjector) InjectTransactions(txs ...*models.Transaction) error {
	for _, tx := range txs {
		klog.V(5).Infof("PQInjector: inject transaction %s", tx.ID)
		_, err := pqitr.db.Model(tx).Insert()
		if err != nil {
			return err
		}
	}
	return nil
}
