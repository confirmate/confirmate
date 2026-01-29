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
	"github.com/urfave/cli/v3"
)

// NewRootCommand returns the root CLI command for the Confirmate orchestrator.
func NewRootCommand() *cli.Command {
	return &cli.Command{
		Name:                  "cf",
		Usage:                 "Confirmate Orchestrator CLI",
		EnableShellCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "addr",
				Usage:   "Orchestrator server address",
				Value:   "http://localhost:8080",
				Sources: cli.EnvVars("CONFIRMATE_ADDR"),
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "tools",
				Usage: "Assessment tool operations",
				Commands: []*cli.Command{
					ToolsCreateCommand(),
					ToolsListCommand(),
					ToolsGetCommand(),
					ToolsDeleteCommand(),
				},
			},
			{
				Name:  "metrics",
				Usage: "Metric operations",
				Commands: []*cli.Command{
					MetricsListCommand(),
					MetricsGetCommand(),
					MetricsDeleteCommand(),
				},
			},
			{
				Name:    "targets",
				Aliases: []string{"toe"},
				Usage:   "Target of evaluation operations",
				Commands: []*cli.Command{
					TargetsCreateCommand(),
					TargetsListCommand(),
					TargetsGetCommand(),
					TargetsDeleteCommand(),
				},
			},
			{
				Name:  "results",
				Usage: "Assessment result operations",
				Commands: []*cli.Command{
					ResultsListCommand(),
					ResultsGetCommand(),
				},
			},
			{
				Name:  "catalogs",
				Usage: "Catalog operations",
				Commands: []*cli.Command{
					CatalogsListCommand(),
					CatalogsGetCommand(),
					CatalogsDeleteCommand(),
				},
			},
			{
				Name:    "certificates",
				Aliases: []string{"certs"},
				Usage:   "Certificate operations",
				Commands: []*cli.Command{
					CertificatesListCommand(),
					CertificatesGetCommand(),
					CertificatesDeleteCommand(),
				},
			},
		},
	}
}
