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
	"fmt"

	"confirmate.io/core/api/orchestrator"

	"connectrpc.com/connect"
	"github.com/urfave/cli/v3"
)

func CatalogsListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List all catalogs",
		Flags: PaginationFlags(),
		Action: func(ctx context.Context, c *cli.Command) error {
			client := OrchestratorClient(ctx, c)
			resp, err := client.ListCatalogs(ctx, connect.NewRequest(&orchestrator.ListCatalogsRequest{
				PageSize:  int32(c.Int("page-size")),
				PageToken: c.String("page-token"),
			}))
			if err != nil {
				return err
			}
			return PrettyPrint(resp.Msg)
		},
	}
}

func CatalogsGetCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Get a specific catalog by ID",
		ArgsUsage: "<catalog-id>",
		Action: func(ctx context.Context, c *cli.Command) error {
			if c.Args().Len() < 1 {
				return fmt.Errorf("catalog ID required")
			}
			catalogID := c.Args().Get(0)

			client := OrchestratorClient(ctx, c)
			resp, err := client.GetCatalog(ctx, connect.NewRequest(&orchestrator.GetCatalogRequest{
				CatalogId: catalogID,
			}))
			if err != nil {
				return err
			}
			return PrettyPrint(resp.Msg)
		},
	}
}

func CatalogsRemoveCommand() *cli.Command {
	return &cli.Command{
		Name:      "remove",
		Aliases:   []string{"rm"},
		Usage:     "Delete a catalog by ID",
		ArgsUsage: "<catalog-id>",
		Action: func(ctx context.Context, c *cli.Command) error {
			if c.Args().Len() < 1 {
				return fmt.Errorf("catalog ID required")
			}
			catalogID := c.Args().Get(0)

			client := OrchestratorClient(ctx, c)
			_, err := client.RemoveCatalog(ctx, connect.NewRequest(&orchestrator.RemoveCatalogRequest{
				CatalogId: catalogID,
			}))
			if err != nil {
				return err
			}
			fmt.Printf("Catalog %s deleted successfully\n", catalogID)
			return nil
		},
	}
}
