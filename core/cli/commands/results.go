package commands

import (
	"context"
	"fmt"
	"net/http"

	"github.com/urfave/cli/v3"
	"connectrpc.com/connect"
	"confirmate.io/core/api/orchestrator/orchestratorconnect"
	"confirmate.io/core/api/orchestrator"
)

func ResultsListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List all assessment results",
		Flags: []cli.Flag{
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
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			req := &orchestrator.ListAssessmentResultsRequest{}
			
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
			
			client := orchestratorconnect.NewOrchestratorClient(http.DefaultClient, "http://localhost:8080")
			resp, err := client.ListAssessmentResults(ctx, connect.NewRequest(req))
			if err != nil {
				return err
			}
			fmt.Printf("%+v\n", resp.Msg)
			return nil
		},
	}
}

func ResultsGetCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Get a specific assessment result by ID",
		ArgsUsage: "<result-id>",
		Action: func(ctx context.Context, c *cli.Command) error {
			if c.NArg() < 1 {
				return fmt.Errorf("result ID required")
			}
			resultID := c.Args().First()
			
			client := orchestratorconnect.NewOrchestratorClient(http.DefaultClient, "http://localhost:8080")
			resp, err := client.GetAssessmentResult(ctx, connect.NewRequest(&orchestrator.GetAssessmentResultRequest{
				Id: resultID,
			}))
			if err != nil {
				return err
			}
			fmt.Printf("%+v\n", resp.Msg)
			return nil
		},
	}
}
