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

	"github.com/google/uuid"
	"github.com/urfave/cli/v3"
)

type noOpCollector struct {
	id   string
	name string
}

func newNoOpCollector(name string) noOpCollector {
	if name == "" {
		name = "no-op-collector"
	}

	return noOpCollector{
		id:   uuid.NewString(),
		name: name,
	}
}

func (c noOpCollector) ID() string {
	return c.id
}

func (c noOpCollector) Name() string {
	return c.name
}

func (noOpCollector) Collect() (list []ontology.IsResource, err error) {
	return nil, nil
}

// collectionFlags contains the flags that are specific to configuring the collection service.
var collectionFlags = []cli.Flag{
	&cli.DurationFlag{
		Name:    "collection-interval",
		Usage:   "Interval between collection runs",
		Value:   collection.DefaultConfig.Interval,
		Sources: envVarSources("collection-interval"),
	},
	&cli.StringFlag{
		Name:    "evidence-store-address",
		Usage:   "Evidence store base URL for forwarding collected resources (empty disables forwarding)",
		Value:   collection.DefaultConfig.EvidenceStoreAddress,
		Sources: envVarSources("evidence-store-address"),
	},
	&cli.StringFlag{
		Name:    "target-of-evaluation-id",
		Usage:   "Target of evaluation UUID used when creating evidence records",
		Value:   "",
		Sources: envVarSources("target-of-evaluation-id"),
	},
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
			Interval:             cmd.Duration("collection-interval"),
			EvidenceStoreAddress: cmd.String("evidence-store-address"),
			TargetOfEvaluationID: cmd.String("target-of-evaluation-id"),
			Collectors: []collection.Collector{
				newNoOpCollector("cli-no-op-collector"),
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
	Flags: joinFlagSlices(
		logFlags,
		collectionFlags,
	),
}
