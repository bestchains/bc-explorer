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
	"context"
	"crypto/sha256"
	"encoding/asn1"
	"encoding/hex"
	"math/big"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-protos-go-apiv2/common"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"k8s.io/klog/v2"

	"github.com/bestchains/bc-explorer/pkg/internal/hyperledger/fabric/protoutil"
	"github.com/bestchains/bc-explorer/pkg/network"

	"github.com/bestchains/bc-explorer/pkg/errorsq"
	"github.com/bestchains/bc-explorer/pkg/models"
)

var (
	errInvalidFabTx = errors.New("invalid fabric transaction")
)

type BlockEventListener interface {
	CheckPoint() uint64
	Close()
	Events()
}

type fabEventListener struct {
	ctx    context.Context
	cancel context.CancelFunc

	errq errorsq.Errorsq

	nid string

	startBlock uint64

	events <-chan *common.Block

	injector Injector
}

func newFabEventListener(pctx context.Context, errq errorsq.Errorsq, injector Injector, net *network.Network, startBlock uint64) (BlockEventListener, error) {
	if errq == nil {
		return nil, errors.New("nil errorsq")
	}
	ctx, cancel := context.WithCancel(pctx)
	listener := &fabEventListener{
		ctx:        ctx,
		cancel:     cancel,
		errq:       errq,
		nid:        net.ID,
		injector:   injector,
		startBlock: startBlock,
	}
	fabclient, err := network.NewFabricClient(net)
	if err != nil {
		cancel()
		return nil, err
	}
	events, err := fabclient.Channel("").BlockEvents(ctx, client.WithStartBlock(startBlock))
	if err != nil {
		cancel()
		return nil, err
	}
	listener.events = events
	return listener, nil
}

func (listener *fabEventListener) CheckPoint() uint64 {
	return listener.startBlock
}

func (listener *fabEventListener) Close() {
	listener.cancel()
}

func (listener *fabEventListener) Events() {
	klog.Infof("Start block event listening on network %s", listener.nid)
	defer func() {
		klog.Infof("Stop block event listening on network %s", listener.nid)
	}()
	for {
		blk, ok := <-listener.events
		if !ok {
			return
		}
		klog.V(5).Infof("Received new block %d for network %s", blk.Header.Number+1, listener.nid)
		if err := listener.fabBlkHandler(blk); err != nil {
			listener.errq.Send(err)
		}
	}
}

func (listener *fabEventListener) fabBlkHandler(block *common.Block) error {
	var err error

	blk := &models.Block{
		Network:           listener.nid,
		BlockNumber:       block.Header.Number + 1, // postgresql treat 0 as null,so we start from 1
		BlockHash:         hex.EncodeToString(blockHash(block.Header)),
		PrevioudBlockHash: hex.EncodeToString(block.Header.PreviousHash),
		DataHash:          hex.EncodeToString(block.Header.DataHash),
		BlockSize:         proto.Size(block),
	}

	txsData := block.Data.GetData()
	var txs = make([]*models.Transaction, len(txsData))
	for index, txData := range txsData {
		tx, err := parseFabTx(listener.nid, blk.BlockNumber, txData)
		if err != nil {
			return errors.Wrap(errInvalidFabTx, err.Error())
		}
		txs[index] = tx

		if blk.CreatedAt == 0 {
			blk.CreatedAt = tx.CreatedAt
		}
	}

	blk.TxCount = len(txs)

	if listener.injector != nil {
		err = listener.injector.InjectBlocks(blk)
		if err != nil {
			return err
		}
		err = listener.injector.InjectTransactions(txs...)
		if err != nil {
			return err
		}
	}

	return nil
}

func blockHash(b *common.BlockHeader) []byte {
	sum := sha256.Sum256(blockHeaderBytes(b))
	return sum[:]
}

type asn1Header struct {
	Number       *big.Int
	PreviousHash []byte
	DataHash     []byte
}

func blockHeaderBytes(b *common.BlockHeader) []byte {
	asn1Header := asn1Header{
		PreviousHash: b.PreviousHash,
		DataHash:     b.DataHash,
		Number:       new(big.Int).SetUint64(b.Number),
	}
	result, err := asn1.Marshal(asn1Header)
	if err != nil {
		return []byte{}
	}
	return result
}

func parseFabTx(network string, blockNumber uint64, txData []byte) (*models.Transaction, error) {
	tx, err := protoutil.GetTransactionFromEnvelope(txData)
	if err != nil {
		return nil, err
	}

	tx.Network = network
	tx.BlockNumber = blockNumber

	return tx, nil
}
