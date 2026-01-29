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

func TargetsListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List all targets of evaluation",
		Flags: PaginationFlags(),
		Action: func(ctx context.Context, c *cli.Command) error {
			client := OrchestratorClient(c)
			resp, err := client.ListTargetsOfEvaluation(ctx, connect.NewRequest(&orchestrator.ListTargetsOfEvaluationRequest{
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

func TargetsGetCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Get a specific target of evaluation by ID",
		ArgsUsage: "<target-id>",
		Action: func(ctx context.Context, c *cli.Command) error {
			if c.Args().Len() < 1 {
				return fmt.Errorf("target ID required")
			}
			targetID := c.Args().Get(0)
			
			client := OrchestratorClient(c)
			resp, err := client.GetTargetOfEvaluation(ctx, connect.NewRequest(&orchestrator.GetTargetOfEvaluationRequest{
				TargetOfEvaluationId: targetID,
			}))
			if err != nil {
				return err
			}
			return PrettyPrint(resp.Msg)
		},
	}
}

func TargetsDeleteCommand() *cli.Command {
	return &cli.Command{
		Name:      "delete",
		Aliases:   []string{"rm"},
		Usage:     "Delete a target of evaluation by ID",
		ArgsUsage: "<target-id>",
		Action: func(ctx context.Context, c *cli.Command) error {
			if c.Args().Len() < 1 {
				return fmt.Errorf("target ID required")
			}
			targetID := c.Args().Get(0)
			
			client := OrchestratorClient(c)
			_, err := client.RemoveTargetOfEvaluation(ctx, connect.NewRequest(&orchestrator.RemoveTargetOfEvaluationRequest{
				TargetOfEvaluationId: targetID,
			}))
			if err != nil {
				return err
			}
			fmt.Printf("Target %s deleted successfully\n", targetID)
			return nil
		},
	}
}
