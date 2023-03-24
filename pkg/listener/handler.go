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

	"github.com/pkg/errors"

	"github.com/bestchains/bc-explorer/pkg/network"
	"github.com/gofiber/fiber/v2"
)

var (
	errInvalidNetwork = errors.New("invalid network")
)

type Handler struct {
	listener Listener
}

func NewHandler(listener Listener) *Handler {
	handler := &Handler{
		listener: listener,
	}
	return handler
}

func (handler *Handler) List(c *fiber.Ctx) error {
	nets, err := handler.listener.Selector().Networks("id", "type", "platform", "status")
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(nets)
}

func (handler *Handler) Register(c *fiber.Ctx) error {
	var err error
	net := new(network.Network)
	err = c.BodyParser(net)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("%s: %s", errInvalidNetwork.Error(), err.Error()))
	}

	err = handler.listener.Register(net)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.SendStatus(fiber.StatusOK)
}

func (handler *Handler) Deregister(c *fiber.Ctx) error {
	nid := c.Params("nid")

	err := handler.listener.Deregister(nid)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.SendStatus(fiber.StatusOK)
}

func (handler *Handler) Delete(c *fiber.Ctx) error {
	nid := c.Params("nid")

	err := handler.listener.Delete(nid)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.SendStatus(fiber.StatusOK)
}
