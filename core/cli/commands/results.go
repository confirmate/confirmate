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

func ResultsListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List all assessment results",
		Flags: append([]cli.Flag{
			&cli.StringFlag{
				Name:    "target",
				Aliases: []string{"t"},
				Usage:   "Filter by target of evaluation ID",
			},
			&cli.StringFlag{
				Name:    "metric",
				Aliases: []string{"m"},
				Usage:   "Filter by metric ID",
			},
			&cli.BoolFlag{
				Name:    "compliant",
				Aliases: []string{"c"},
				Usage:   "Filter only compliant results",
			},
		}, PaginationFlags()...),
		Action: func(ctx context.Context, c *cli.Command) error {
			req := &orchestrator.ListAssessmentResultsRequest{
				PageSize:  int32(c.Int("page-size")),
				PageToken: c.String("page-token"),
			}

			// Apply filters if provided
			if c.String("target") != "" || c.String("metric") != "" || c.IsSet("compliant") {
				filter := &orchestrator.ListAssessmentResultsRequest_Filter{}

				if targetID := c.String("target"); targetID != "" {
					filter.TargetOfEvaluationId = &targetID
				}
				if metricID := c.String("metric"); metricID != "" {
					filter.MetricId = &metricID
				}
				if c.IsSet("compliant") {
					compliant := c.Bool("compliant")
					filter.Compliant = &compliant
				}

				req.Filter = filter
			}

			client := OrchestratorClient(ctx, c)
			resp, err := client.ListAssessmentResults(ctx, connect.NewRequest(req))
			if err != nil {
				return err
			}
			return PrettyPrint(resp.Msg)
		},
	}
}

func ResultsGetCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Get a specific assessment result by ID",
		ArgsUsage: "<result-id>",
		Action: func(ctx context.Context, c *cli.Command) error {
			if c.Args().Len() < 1 {
				return fmt.Errorf("result ID required")
			}
			resultID := c.Args().Get(0)

			client := OrchestratorClient(ctx, c)
			resp, err := client.GetAssessmentResult(ctx, connect.NewRequest(&orchestrator.GetAssessmentResultRequest{
				Id: resultID,
			}))
			if err != nil {
				return err
			}
			return PrettyPrint(resp.Msg)
		},
	}
}
