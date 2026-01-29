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
	"github.com/urfave/cli/v3"

	"confirmate.io/core/api/orchestrator"
)

func ToolsListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List all assessment tools",
		Flags: PaginationFlags(),
		Action: func(ctx context.Context, c *cli.Command) error {
			client := OrchestratorClient(c)
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
			
			client := OrchestratorClient(c)
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
			
			client := OrchestratorClient(c)
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
