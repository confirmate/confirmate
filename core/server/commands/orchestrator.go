// Copyright 2016-2025 Fraunhofer AISEC
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

	"confirmate.io/core/api/orchestrator/orchestratorconnect"
	"confirmate.io/core/server"
	"confirmate.io/core/service/orchestrator"

	"github.com/urfave/cli/v3"
)

// OrchestratorCommand is the command to start the orchestrator server.
var OrchestratorCommand = &cli.Command{
	Name:  "orchestrator",
	Usage: "Launches the orchestrator service",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		svc, err := orchestrator.NewService(
			orchestrator.WithConfig(orchestrator.Config{
				DefaultCatalogsFolder:           cmd.String("catalogs-default-path"),
				LoadDefaultCatalogs:             cmd.Bool("catalogs-load-default"),
				DefaultMetricsPath:              cmd.String("metrics-default-path"),
				LoadDefaultMetrics:              cmd.Bool("metrics-load-default"),
				CreateDefaultTargetOfEvaluation: cmd.Bool("create-default-target-of-evaluation"),
			}),
		)
		if err != nil {
			return err
		}

		return server.RunConnectServer(
			server.WithConfig(server.Config{
				Port: cmd.Uint16("api-port"),
				Path: "/",
				CORS: server.CORS{
					AllowedOrigins: cmd.StringSlice("api-cors-allowed-origins"),
					AllowedMethods: cmd.StringSlice("api-cors-allowed-methods"),
					AllowedHeaders: cmd.StringSlice("api-cors-allowed-headers"),
				},
			}),
			server.WithHandler(orchestratorconnect.NewOrchestratorHandler(svc)),
		)
	},
	Flags: []cli.Flag{
		&cli.Uint16Flag{
			Name:  "api-port",
			Usage: "Port to run the API server (Connect, gRPC, REST) on",
			Value: server.DefaultConfig.Port,
		},
		&cli.StringSliceFlag{
			Name:  "api-cors-allowed-origins",
			Usage: "Specifies the origins allowed in CORS",
			Value: server.DefaultConfig.CORS.AllowedOrigins,
		},
		&cli.StringSliceFlag{
			Name:  "api-cors-allowed-methods",
			Usage: "Specifies the methods allowed in CORS",
			Value: server.DefaultConfig.CORS.AllowedMethods,
		},
		&cli.StringSliceFlag{
			Name:  "api-cors-allowed-headers",
			Usage: "Specifies the headers allowed in CORS",
			Value: server.DefaultConfig.CORS.AllowedHeaders,
		},
		&cli.StringFlag{
			Name:  "catalogs-default-path",
			Usage: "The path to the folder containing default catalog definitions",
			Value: orchestrator.DefaultConfig.DefaultCatalogsFolder,
		},
		&cli.BoolFlag{
			Name:  "catalogs-load-default",
			Usage: "Load default catalogs from the catalogs-default-path",
			Value: orchestrator.DefaultConfig.LoadDefaultCatalogs,
		},
		&cli.StringFlag{
			Name:  "metrics-default-path",
			Usage: "The path to the folder containing default metrics (e.g., security-metrics repository)",
			Value: orchestrator.DefaultConfig.DefaultMetricsPath,
		},
		&cli.BoolFlag{
			Name:  "metrics-load-default",
			Usage: "Load default metrics from the metrics-default-path",
			Value: orchestrator.DefaultConfig.LoadDefaultMetrics,
		},
		&cli.BoolFlag{
			Name:  "create-default-target-of-evaluation",
			Usage: "Creates a default target of evaluation if none exists",
			Value: orchestrator.DefaultConfig.CreateDefaultTargetOfEvaluation,
		},
	},
}
