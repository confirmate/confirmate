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
		svc, err := orchestrator.NewService()
		if err != nil {
			return err
		}

		return server.RunConnectServer(
			server.WithConfig(server.Config{
				Port: cmd.Uint16("port"),
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
	},
}
