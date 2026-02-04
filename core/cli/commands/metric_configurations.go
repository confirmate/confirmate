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

func MetricConfigurationsListCommand() *cli.Command {
	return &cli.Command{
		Name:      "list",
		Usage:     "List metric configurations for a target of evaluation",
		ArgsUsage: "<target-id>",
		Flags:     PaginationFlags(),
		Action: func(ctx context.Context, c *cli.Command) error {
			if c.Args().Len() < 1 {
				return fmt.Errorf("target ID required")
			}
			targetID := c.Args().Get(0)

			client := OrchestratorClient(ctx, c)
			resp, err := client.ListMetricConfigurations(ctx, connect.NewRequest(&orchestrator.ListMetricConfigurationRequest{
				TargetOfEvaluationId: targetID,
				PageSize:             int32(c.Int("page-size")),
				PageToken:            c.String("page-token"),
			}))
			if err != nil {
				return err
			}
			return PrettyPrint(resp.Msg)
		},
	}
}

func MetricConfigurationsGetCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Get a metric configuration by target and metric ID",
		ArgsUsage: "<target-id> <metric-id>",
		Action: func(ctx context.Context, c *cli.Command) error {
			if c.Args().Len() < 2 {
				return fmt.Errorf("target ID and metric ID required")
			}
			targetID := c.Args().Get(0)
			metricID := c.Args().Get(1)

			client := OrchestratorClient(ctx, c)
			resp, err := client.GetMetricConfiguration(ctx, connect.NewRequest(&orchestrator.GetMetricConfigurationRequest{
				TargetOfEvaluationId: targetID,
				MetricId:             metricID,
			}))
			if err != nil {
				return err
			}
			return PrettyPrint(resp.Msg)
		},
	}
}