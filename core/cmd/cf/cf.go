package main

import (
	"context"
	"log"
	"os"

	"github.com/urfave/cli/v3"

	"confirmate.io/core/cli/commands"
)

func main() {
	cmd := &cli.Command{
		Name:                   "cf",
		Usage:                  "Confirmate Orchestrator CLI",
		EnableShellCompletion:  true,
		Commands: []*cli.Command{
			{
				Name:  "tools",
				Usage: "Assessment tool operations",
				Commands: []*cli.Command{
					commands.ToolsListCommand(),
					commands.ToolsGetCommand(),
					commands.ToolsDeleteCommand(),
				},
			},
			{
				Name:  "metrics",
				Usage: "Metric operations",
				Commands: []*cli.Command{
					commands.MetricsListCommand(),
					commands.MetricsGetCommand(),
					commands.MetricsDeleteCommand(),
				},
			},
			{
				Name:    "targets",
				Aliases: []string{"toe"},
				Usage:   "Target of evaluation operations",
				Commands: []*cli.Command{
					commands.TargetsListCommand(),
					commands.TargetsGetCommand(),
					commands.TargetsDeleteCommand(),
				},
			},
			{
				Name:  "results",
				Usage: "Assessment result operations",
				Commands: []*cli.Command{
					commands.ResultsListCommand(),
					commands.ResultsGetCommand(),
				},
			},
			{
				Name:  "catalogs",
				Usage: "Catalog operations",
				Commands: []*cli.Command{
					commands.CatalogsListCommand(),
					commands.CatalogsGetCommand(),
					commands.CatalogsDeleteCommand(),
				},
			},
			{
				Name:    "certificates",
				Aliases: []string{"certs"},
				Usage:   "Certificate operations",
				Commands: []*cli.Command{
					commands.CertificatesListCommand(),
					commands.CertificatesGetCommand(),
					commands.CertificatesDeleteCommand(),
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
