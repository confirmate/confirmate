// Copyright 2016-2026 Fraunhofer AISEC
//
// SPDX-License-Identifier: Apache-2.0
//
//                                 /$$$$$$  /$$                                     /$$
//                               /$$__  $$|__/                                    | $$
//   /$$$$$$$  /$$$$$$  /$$$$$$$ | $$  \__/ /$$  /$$$$$$  /$$$$$$/$$$$   /$$$$$$  /$$$$$$    /$$$$$$
//  /$$_____/ /$$__  $$| $$__  $$| $$$$    | $$ /$$__  $$| $$_  $$_  $$ |____  $$|_  $$_/   /$$__  $$
// | $$      | $$  \ $$| $$  \ $$| $$_/    | $$| $$  \__/| $$ \ $$ \ $$  /$$$$$$$  | $$    | $$$$$$$$
// | $$      | $$  | $$| $$  | $$| $$      | $$| $$      | $$ | $$ | $$ /$$__  $$  | $$ /$$| $$_____/
// |  $$$$$$$|  $$$$$$/| $$  | $$| $$      | $$| $$      | $$ | $$ | $$|  $$$$$$$  |  $$$$/|  $$$$$$$
// \_______/ \______/ |__/  |__/|__/      |__/|__/      |__/ |__/ |__/ \_______/   \___/   \_______/
//
// This file is part of Confirmate Core.

package commands

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"confirmate.io/core/api/ontology"
	"confirmate.io/core/log"
	"confirmate.io/core/service/collection"

	"github.com/urfave/cli/v3"
)

type noOpCollector struct{}

func (noOpCollector) Collect() (list []ontology.IsResource, err error) {
	return nil, nil
}

// CollectionCommand is the command to start the collection service.
var CollectionCommand = &cli.Command{
	Name:  "collection",
	Usage: "Launches the collection service",
	Action: func(ctx context.Context, cmd *cli.Command) (err error) {
		var (
			runCtx   context.Context
			cancel   context.CancelFunc
			svc      *collection.Service
			resultCh <-chan collection.CollectionResult
		)

		runCtx, cancel = signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
		defer cancel()

		if err = log.Configure(cmd.String("log-level")); err != nil {
			return err
		}

		svc, err = collection.NewService(collection.Config{
			Interval: cmd.Duration("collection-interval"),
			Collectors: []collection.Collector{
				noOpCollector{},
			},
		})
		if err != nil {
			return err
		}

		resultCh = svc.Start(runCtx)
		for range resultCh {
			slog.Debug("Collection cycle finished")
		}

		return nil
	},
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "log-level",
			Usage: "Log level (TRACE, DEBUG, INFO, WARN, ERROR)",
			Value: "INFO",
		},
		&cli.DurationFlag{
			Name:  "collection-interval",
			Usage: "Interval between collection runs",
			Value: collection.DefaultConfig.Interval,
		},
	},
}
