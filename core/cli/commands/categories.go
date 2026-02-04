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

func CategoriesGetCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Get a category by catalog ID and name",
		ArgsUsage: "<catalog-id> <category-name>",
		Action: func(ctx context.Context, c *cli.Command) error {
			if c.Args().Len() < 2 {
				return fmt.Errorf("catalog ID and category name required")
			}
			catalogID := c.Args().Get(0)
			categoryName := c.Args().Get(1)

			client := OrchestratorClient(ctx, c)
			resp, err := client.GetCategory(ctx, connect.NewRequest(&orchestrator.GetCategoryRequest{
				CatalogId:    catalogID,
				CategoryName: categoryName,
			}))
			if err != nil {
				return err
			}
			return PrettyPrint(resp.Msg)
		},
	}
}
