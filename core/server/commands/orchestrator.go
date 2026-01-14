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

	"github.com/urfave/cli/v3"

	"confirmate.io/core/api/orchestrator/orchestratorconnect"
	"confirmate.io/core/server"
	"confirmate.io/core/service/orchestrator"
)

// OrchestratorCommand is the command to start the orchestrator server.
var OrchestratorCommand = &cli.Command{
	Name:  "orchestrator",
	Usage: "Launches the orchestrator service",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		svc, err := orchestrator.NewService(
			orchestrator.WithConfig(orchestrator.Config{
				CatalogsFolder:                  cmd.String("catalogs-folder"),
				DefaultMetricsPath:              cmd.String("default-metrics-path"),
				AdditionalMetricsPath:           cmd.String("additional-metrics-path"),
				CreateDefaultTargetOfEvaluation: cmd.Bool("create-default-target-of-evaluation"),
				IgnoreDefaultMetrics:            cmd.Bool("ignore-default-metrics"),
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
			Name:  "catalogs-folder",
			Usage: "The folder containing catalog definitions",
			Value: orchestrator.DefaultConfig.CatalogsFolder,
		},
		&cli.StringFlag{
			Name:  "default-metrics-path",
			Usage: "The path to the default metric definitions (security-metrics submodule)",
			Value: orchestrator.DefaultConfig.DefaultMetricsPath,
		},
		&cli.StringFlag{
			Name:  "additional-metrics-path",
			Usage: "The path to a folder containing additional custom metric definitions",
		},
		&cli.BoolFlag{
			Name:  "create-default-target-of-evaluation",
			Usage: "Creates a default target of evaluation if none exists",
			Value: orchestrator.DefaultConfig.CreateDefaultTargetOfEvaluation,
		},
		&cli.BoolFlag{
			Name:  "ignore-default-metrics",
			Usage: "Skips loading default metrics from the security-metrics submodule",
			Value: orchestrator.DefaultConfig.IgnoreDefaultMetrics,
		},
	},
}
