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

func ControlsListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List controls",
		Flags: append(PaginationFlags(),
			&cli.StringFlag{
				Name:  "catalog-id",
				Usage: "Filter by catalog ID",
			},
			&cli.StringFlag{
				Name:  "category-name",
				Usage: "Filter by category name",
			},
			&cli.StringSliceFlag{
				Name:  "assurance-level",
				Usage: "Filter by assurance level (repeatable or comma-separated)",
			},
		),
		Action: func(ctx context.Context, c *cli.Command) error {
			levels := ExpandCommaSeparated(c.StringSlice("assurance-level"))

			req := &orchestrator.ListControlsRequest{
				CatalogId:    c.String("catalog-id"),
				CategoryName: c.String("category-name"),
				PageSize:     int32(c.Int("page-size")),
				PageToken:    c.String("page-token"),
			}
			if len(levels) > 0 {
				req.Filter = &orchestrator.ListControlsRequest_Filter{
					AssuranceLevels: levels,
				}
			}

			client := OrchestratorClient(ctx, c)
			resp, err := client.ListControls(ctx, connect.NewRequest(req))
			if err != nil {
				return err
			}
			return PrettyPrint(resp.Msg)
		},
	}
}

func ControlsGetCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Get a control by catalog ID, category name, and control ID",
		ArgsUsage: "<catalog-id> <category-name> <control-id>",
		Action: func(ctx context.Context, c *cli.Command) error {
			if c.Args().Len() < 3 {
				return fmt.Errorf("catalog ID, category name, and control ID required")
			}
			catalogID := c.Args().Get(0)
			categoryName := c.Args().Get(1)
			controlID := c.Args().Get(2)

			client := OrchestratorClient(ctx, c)
			resp, err := client.GetControl(ctx, connect.NewRequest(&orchestrator.GetControlRequest{
				CatalogId:    catalogID,
				CategoryName: categoryName,
				ControlId:    controlID,
			}))
			if err != nil {
				return err
			}
			return PrettyPrint(resp.Msg)
		},
	}
}
