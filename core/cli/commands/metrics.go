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

func MetricsListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List all metrics",
		Action: func(ctx context.Context, c *cli.Command) error {
			client := orchestratorconnect.NewOrchestratorClient(http.DefaultClient, "http://localhost:8080")
			resp, err := client.ListMetrics(ctx, connect.NewRequest(&orchestrator.ListMetricsRequest{}))
			if err != nil {
				return err
			}
			fmt.Printf("%+v\n", resp.Msg)
			return nil
		},
	}
}

func MetricsGetCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Get a specific metric by ID",
		ArgsUsage: "<metric-id>",
		Action: func(ctx context.Context, c *cli.Command) error {
			if c.NArg() < 1 {
				return fmt.Errorf("metric ID required")
			}
			metricID := c.Args().First()
			
			client := orchestratorconnect.NewOrchestratorClient(http.DefaultClient, "http://localhost:8080")
			resp, err := client.GetMetric(ctx, connect.NewRequest(&orchestrator.GetMetricRequest{
				MetricId: metricID,
			}))
			if err != nil {
				return err
			}
			fmt.Printf("%+v\n", resp.Msg)
			return nil
		},
	}
}

func MetricsDeleteCommand() *cli.Command {
	return &cli.Command{
		Name:      "delete",
		Aliases:   []string{"rm"},
		Usage:     "Delete a metric by ID",
		ArgsUsage: "<metric-id>",
		Action: func(ctx context.Context, c *cli.Command) error {
			if c.NArg() < 1 {
				return fmt.Errorf("metric ID required")
			}
			metricID := c.Args().First()
			
			client := orchestratorconnect.NewOrchestratorClient(http.DefaultClient, "http://localhost:8080")
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
