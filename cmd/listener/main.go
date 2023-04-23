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

package main

import (
	"context"
	"flag"

	"github.com/bestchains/bc-explorer/pkg/auth"
	"github.com/bestchains/bc-explorer/pkg/errorsq"
	bclistener "github.com/bestchains/bc-explorer/pkg/listener"
	"github.com/go-pg/pg/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"k8s.io/klog/v2"
)

var (
	injector   = flag.String("injector", "pg", "used to initialize injector")
	dsn        = flag.String("dsn", "postgres://bestchains:Passw0rd!@127.0.0.1:5432/bc-explorer?sslmode=disable", "database connection string")
	addr       = flag.String("addr", ":9999", "used to listen and serve http requests")
	authMethod = flag.String("auth", "none", "user authentication method, none, oidc or kubernetes")
)

func main() {
	flag.Parse()
	if err := run(); err != nil {
		klog.Error(err)
	}
}

func run() error {
	pctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errq := errorsq.NewErrorsq(pctx, func(err error) {
		klog.Errorln("bc-explorer", err)
	})

	klog.Infoln("Creating a blockchain listener")

	var itr bclistener.Injector
	var str bclistener.Selector
	var err error
	if *injector == "pg" {
		klog.Infoln("Using injector postgreSQL")
		opts, err := pg.ParseURL(*dsn)
		if err != nil {
			return err
		}
		db := pg.Connect(opts)
		defer db.Close()
		if err := db.Ping(pctx); err != nil {
			panic(err)
		}

		itr, err = bclistener.NewPQInjector(db)
		if err != nil {
			return err
		}
		str, err = bclistener.NewPQSelector(db)
		if err != nil {
			return err
		}
	} else {
		klog.Infoln("Using injector log")
		itr = bclistener.NewLogInjector(func(args ...interface{}) {
			klog.Infoln(args...)
		})
	}
	listener, err := bclistener.NewListener(pctx, errq, itr, str)
	if err != nil {
		return err
	}

	klog.Infoln("Creating http server")
	handler := bclistener.NewHandler(listener)
	app := fiber.New(fiber.Config{
		CaseSensitive: true,
		StrictRouting: true,
		Immutable:     true,
		AppName:       "bc-explorer-listener",
	})
	app.Use(cors.New(cors.ConfigDefault))
	app.Use(logger.New(logger.Config{
		Format: "[${ip}]:${port} ${status} - ${method} ${path}\n",
	}))
	app.Use(auth.New(pctx, auth.Config{
		AuthMethod: *authMethod,
	}))

	// handlers
	app.Get("/networks", handler.List)
	// Register and start listening blockchain network
	app.Post("/network/register", handler.Register)
	// Stop listening blockchain network and set network status to `Deregistered`
	app.Post("/network/deregister/:nid", handler.Deregister)
	// Delete this network along with all data
	app.Delete("/network/:nid", handler.Delete)

	err = app.Listen(*addr)
	if err != nil {
		errq.Send(err)
	}

	return nil
}
