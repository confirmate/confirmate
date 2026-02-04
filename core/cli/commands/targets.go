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
	"strings"

	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/orchestrator"

	"connectrpc.com/connect"
	"github.com/urfave/cli/v3"
)

func TargetsCreateCommand() *cli.Command {
	return &cli.Command{
		Name:  "create",
		Usage: "Create a target of evaluation",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "name",
				Usage:    "Target name",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "description",
				Usage:    "Target description",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "target-type",
				Usage:    "Target type (cloud|product|organization)",
				Required: true,
			},
			&cli.StringSliceFlag{
				Name:     "metric-id",
				Usage:    "Metric ID to configure (repeatable or comma-separated)",
				Required: true,
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			metricsIDs := ExpandCommaSeparated(c.StringSlice("metric-id"))
			if len(metricsIDs) == 0 {
				return fmt.Errorf("at least one --metric-id is required")
			}

			targetType, err := parseTargetType(c.String("target-type"))
			if err != nil {
				return err
			}

			client := OrchestratorClient(ctx, c)
			metrics := make([]*assessment.Metric, 0, len(metricsIDs))
			for _, id := range metricsIDs {
				resp, getErr := client.GetMetric(ctx, connect.NewRequest(&orchestrator.GetMetricRequest{
					MetricId: id,
				}))
				if getErr != nil {
					return getErr
				}
				metrics = append(metrics, resp.Msg)
			}

			resp, err := client.CreateTargetOfEvaluation(ctx, connect.NewRequest(&orchestrator.CreateTargetOfEvaluationRequest{
				TargetOfEvaluation: &orchestrator.TargetOfEvaluation{
					Name:              c.String("name"),
					Description:       c.String("description"),
					ConfiguredMetrics: metrics,
					TargetType:        targetType,
				},
			}))
			if err != nil {
				return err
			}
			return PrettyPrint(resp.Msg)
		},
	}
}

func parseTargetType(value string) (orchestrator.TargetOfEvaluation_TargetType, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case "cloud", "target_type_cloud":
		return orchestrator.TargetOfEvaluation_TARGET_TYPE_CLOUD, nil
	case "product", "target_type_product":
		return orchestrator.TargetOfEvaluation_TARGET_TYPE_PRODUCT, nil
	case "organization", "target_type_organization":
		return orchestrator.TargetOfEvaluation_TARGET_TYPE_ORGANIZATION, nil
	default:
		return orchestrator.TargetOfEvaluation_TARGET_TYPE_UNSPECIFIED, fmt.Errorf("invalid target type: %s", value)
	}
}

func TargetsListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List all targets of evaluation",
		Flags: PaginationFlags(),
		Action: func(ctx context.Context, c *cli.Command) error {
			client := OrchestratorClient(ctx, c)
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

			client := OrchestratorClient(ctx, c)
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

func TargetsStatsCommand() *cli.Command {
	return &cli.Command{
		Name:      "stats",
		Aliases:   []string{"statistics"},
		Usage:     "Get statistics for a target of evaluation",
		ArgsUsage: "<target-id>",
		Action: func(ctx context.Context, c *cli.Command) error {
			if c.Args().Len() < 1 {
				return fmt.Errorf("target ID required")
			}
			targetID := c.Args().Get(0)

			client := OrchestratorClient(ctx, c)
			resp, err := client.GetTargetOfEvaluationStatistics(ctx, connect.NewRequest(&orchestrator.GetTargetOfEvaluationStatisticsRequest{
				TargetOfEvaluationId: targetID,
			}))
			if err != nil {
				return err
			}
			return PrettyPrint(resp.Msg)
		},
	}
}

func TargetsRemoveCommand() *cli.Command {
	return &cli.Command{
		Name:      "remove",
		Aliases:   []string{"rm"},
		Usage:     "Delete a target of evaluation by ID",
		ArgsUsage: "<target-id>",
		Action: func(ctx context.Context, c *cli.Command) error {
			if c.Args().Len() < 1 {
				return fmt.Errorf("target ID required")
			}
			targetID := c.Args().Get(0)

			client := OrchestratorClient(ctx, c)
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
