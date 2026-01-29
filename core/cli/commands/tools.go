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

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/urfave/cli/v3"

	"confirmate.io/core/api/orchestrator"
)

func ToolsCreateCommand() *cli.Command {
	return &cli.Command{
		Name:  "create",
		Usage: "Register a new assessment tool",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "id",
				Usage:    "Tool ID (UUID). If omitted, a new UUID is generated",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "name",
				Usage:    "Tool name",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "description",
				Usage:    "Tool description",
				Required: false,
			},
			&cli.StringSliceFlag{
				Name:     "metric-id",
				Usage:    "Metric ID this tool can assess (repeatable or comma-separated)",
				Required: true,
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			metrics := ExpandCommaSeparated(c.StringSlice("metric-id"))
			if len(metrics) == 0 {
				return fmt.Errorf("at least one --metric-id is required")
			}

			toolID := c.String("id")
			if toolID == "" {
				toolID = uuid.NewString()
			} else if _, err := uuid.Parse(toolID); err != nil {
				return fmt.Errorf("invalid tool id: %w", err)
			}

			client := OrchestratorClient(ctx, c)
			resp, err := client.RegisterAssessmentTool(ctx, connect.NewRequest(&orchestrator.RegisterAssessmentToolRequest{
				Tool: &orchestrator.AssessmentTool{
					Id:               toolID,
					Name:             c.String("name"),
					Description:      c.String("description"),
					AvailableMetrics: metrics,
				},
			}))
			if err != nil {
				return err
			}
			return PrettyPrint(resp.Msg)
		},
	}
}

func ToolsListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List all assessment tools",
		Flags: PaginationFlags(),
		Action: func(ctx context.Context, c *cli.Command) error {
			client := OrchestratorClient(ctx, c)
			resp, err := client.ListAssessmentTools(ctx, connect.NewRequest(&orchestrator.ListAssessmentToolsRequest{
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

func ToolsGetCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Get a specific assessment tool by ID",
		ArgsUsage: "<tool-id>",
		Action: func(ctx context.Context, c *cli.Command) error {
			if c.Args().Len() < 1 {
				return fmt.Errorf("tool ID required")
			}
			toolID := c.Args().Get(0)

			client := OrchestratorClient(ctx, c)
			resp, err := client.GetAssessmentTool(ctx, connect.NewRequest(&orchestrator.GetAssessmentToolRequest{
				ToolId: toolID,
			}))
			if err != nil {
				return err
			}
			return PrettyPrint(resp.Msg)
		},
	}
}

func ToolsDeleteCommand() *cli.Command {
	return &cli.Command{
		Name:      "delete",
		Aliases:   []string{"rm"},
		Usage:     "Delete an assessment tool by ID",
		ArgsUsage: "<tool-id>",
		Action: func(ctx context.Context, c *cli.Command) error {
			if c.Args().Len() < 1 {
				return fmt.Errorf("tool ID required")
			}
			toolID := c.Args().Get(0)

			client := OrchestratorClient(ctx, c)
			_, err := client.DeregisterAssessmentTool(ctx, connect.NewRequest(&orchestrator.DeregisterAssessmentToolRequest{
				ToolId: toolID,
			}))
			if err != nil {
				return err
			}
			fmt.Printf("Tool %s deleted successfully\n", toolID)
			return nil
		},
	}
}
