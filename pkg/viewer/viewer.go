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
	"net/http"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/gofiber/fiber/v2"
	"k8s.io/klog/v2"
)

type handler struct {
	block       Block
	transaction Transaction
	overview    Overview
}

func NewViewHandler(t Transaction, b Block, o Overview) handler {
	return handler{transaction: t, block: b, overview: o}
}

func (h *handler) ListBlocks(ctx *fiber.Ctx) error {
	klog.Infof("viewer ListBlocks")

	arg := BlockArg{
		From:        ctx.QueryInt("from", 0),
		Size:        ctx.QueryInt("size", 10),
		Network:     ctx.Params("network"),
		StartTime:   int64(ctx.QueryInt("startTime", 0)),
		EndTime:     int64(ctx.QueryInt("endTime", 0)),
		BlockNumber: uint64(ctx.QueryInt("blockNumber", 0)),
		BlockHash:   ctx.Query("blockHash"),
	}
	klog.V(5).Infof(" with ctx  %+v arg: %+v\n", *ctx, arg)
	result, count, err := h.block.List(arg)

	if err != nil {
		klog.Error(fmt.Sprintf("List Blocks error %s", err))
		ctx.Status(http.StatusInternalServerError)
		return ctx.JSON(map[string]string{"msg": err.Error()})
	}

	data := map[string]interface{}{
		"data":  result,
		"count": count,
	}
	return ctx.JSON(data)
}

func (h *handler) GetBlock(ctx *fiber.Ctx) error {
	klog.Info("viewer GetBlock")
	blockHash := ctx.Params("blockHash")
	network := ctx.Params("network")
	if blockHash == "" {
		ctx.Status(http.StatusBadRequest)
		return ctx.JSON(map[string]string{"msg": "blockHash can't be empty"})
	}
	arg := BlockArg{BlockHash: blockHash, Network: network}
	klog.V(5).Infof(" with ctx %+v, arg: %+v\n", *ctx, arg)

	result, err := h.block.Get(arg)

	if err != nil {
		klog.Error(fmt.Sprintf("get block error %s", err))
		msg := err.Error()
		ctx.Status(http.StatusInternalServerError)
		if pg.ErrNoRows == err {
			ctx.Status(http.StatusNotFound)
			msg = fmt.Sprintf("not found block %s", blockHash)
		}
		return ctx.JSON(map[string]string{"msg": msg})
	}

	return ctx.JSON(result)
}

func (h *handler) ListTransactions(ctx *fiber.Ctx) error {
	klog.Infof("viewer list transactions")

	arg := TransArg{
		From:        ctx.QueryInt("from", 0),
		Size:        ctx.QueryInt("size", 10),
		NetworkName: ctx.Params("network"),
		StartTime:   int64(ctx.QueryInt("startTime", 0)),
		EndTime:     int64(ctx.QueryInt("endTime", 0)),
		Hash:        ctx.Query("id"),
		BlockNum:    uint64(ctx.QueryInt("blockNumber", 0)),
	}
	klog.V(5).Infof(" with ctx %+v arg: %=v\n", *ctx, arg)
	result, count, err := h.transaction.List(arg)

	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return ctx.JSON(err.Error())
	}

	data := map[string]interface{}{
		"data":  result,
		"count": count,
	}
	return ctx.JSON(data)
}

func (h *handler) GetTransactionByTxHash(ctx *fiber.Ctx) error {
	klog.Info("viewer GetTransactionByTxHash")
	txHash := ctx.Params("txHash")
	network := ctx.Params("network")
	if txHash == "" {
		return fiber.NewError(http.StatusBadRequest, "transaction hash can't be empty")
	}
	if network == "" {
		return fiber.NewError(http.StatusBadRequest, "network name can't be empty")
	}
	arg := TransArg{
		NetworkName: network,
		Hash:        txHash,
	}
	klog.V(5).Infof(" with ctx %+v, arg: %+v\n", *ctx, arg)

	result, err := h.transaction.Get(arg)

	if err != nil {
		klog.Error(fmt.Sprintf("get transaction error: %s", err))
		msg := err.Error()
		ctx.Status(http.StatusInternalServerError)
		if pg.ErrNoRows == err {
			ctx.Status(http.StatusNotFound)
			msg = fmt.Sprintf("transaction hash not found: %s", txHash)
		}
		return ctx.JSON(map[string]string{"msg": msg})
	}

	return ctx.JSON(result)
}

func (h *handler) CountTransactionsCreatedByOrg(ctx *fiber.Ctx) error {
	klog.Info("viewer count transactions created by certain organization")
	network := ctx.Params("network")
	if network == "" {
		return fiber.NewError(http.StatusBadRequest, "network name can't be empty")
	}
	arg := TransArg{
		NetworkName: network,
	}
	klog.V(5).Infof(" with ctx %+v, arg: %+v\n", *ctx, arg)

	result, err := h.transaction.CountByOrg(arg)
	if err != nil {
		klog.Error(fmt.Sprintf("count transaction error: %s", err))
		msg := err.Error()
		ctx.Status(http.StatusInternalServerError)
		if pg.ErrNoRows == err {
			ctx.Status(http.StatusNotFound)
			msg = "no transactions found"
		}
		return ctx.JSON(map[string]string{"msg": msg})
	}

	data := map[string]interface{}{
		"data":  result,
		"count": len(result),
	}

	return ctx.JSON(data)
}

func (h *handler) Summary(ctx *fiber.Ctx) error {
	klog.Info("viewer Summary")
	klog.V(5).Infof(" with ctx %+v\n", *ctx)
	network := ctx.Params("network")
	result, err := h.overview.Summary(network)
	if err != nil {
		klog.Error(err)
		ctx.Status(http.StatusInternalServerError)
		return ctx.JSON(map[string]interface{}{"msg": err.Error()})
	}
	return ctx.JSON(result)
}

func (h *handler) QueryBySeg(ctx *fiber.Ctx) error {
	klog.Info("viewer QueryBySeg")
	klog.V(5).Infof(" with ctx %+v\n", *ctx)

	from := int64(ctx.QueryInt("from"))
	if from == 0 {
		from = time.Now().Unix()
	}
	interval := int64(ctx.QueryInt("interval", 300))
	number := int64(ctx.QueryInt("number", 5))
	_type := ctx.Query("type", BlockAggregation)
	network := ctx.Params("network")

	result, err := h.overview.QueryBySeg(from, interval, number, _type, network)
	if err != nil {
		klog.Error(err)
		ctx.Status(http.StatusInternalServerError)
		return ctx.JSON(map[string]interface{}{"msg": err.Error()})
	}
	return ctx.JSON(result)
}
