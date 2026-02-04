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

func MetricsListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List all metrics",
		Flags: PaginationFlags(),
		Action: func(ctx context.Context, c *cli.Command) error {
			client := OrchestratorClient(ctx, c)
			resp, err := client.ListMetrics(ctx, connect.NewRequest(&orchestrator.ListMetricsRequest{
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

func MetricsGetCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Get a specific metric by ID",
		ArgsUsage: "<metric-id>",
		Action: func(ctx context.Context, c *cli.Command) error {
			if c.Args().Len() < 1 {
				return fmt.Errorf("metric ID required")
			}
			metricID := c.Args().Get(0)

			client := OrchestratorClient(ctx, c)
			resp, err := client.GetMetric(ctx, connect.NewRequest(&orchestrator.GetMetricRequest{
				MetricId: metricID,
			}))
			if err != nil {
				return err
			}

			return PrettyPrint(resp.Msg)
		},
	}
}

func MetricsRemoveCommand() *cli.Command {
	return &cli.Command{
		Name:      "remove",
		Aliases:   []string{"rm"},
		Usage:     "Delete a metric by ID",
		ArgsUsage: "<metric-id>",
		Action: func(ctx context.Context, c *cli.Command) error {
			if c.Args().Len() < 1 {
				return fmt.Errorf("metric ID required")
			}
			metricID := c.Args().Get(0)

			client := OrchestratorClient(ctx, c)
			_, err := client.RemoveMetric(ctx, connect.NewRequest(&orchestrator.RemoveMetricRequest{
				MetricId: metricID,
			}))
			if err != nil {
				return err
			}
			fmt.Printf("Metric %s deleted successfully\n", metricID)
			return nil
		},
	}
}
