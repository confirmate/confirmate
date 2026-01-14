package commands

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/urfave/cli/v3"

	"confirmate.io/core/api/orchestrator"
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
			
			client := OrchestratorClient(c)
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
			
			client := OrchestratorClient(c)
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
