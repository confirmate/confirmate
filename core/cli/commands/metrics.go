package commands

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/urfave/cli/v3"

	"confirmate.io/core/api/orchestrator"
)

func MetricsListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List all metrics",
		Flags: PaginationFlags(),
		Action: func(ctx context.Context, c *cli.Command) error {
			client := OrchestratorClient(c)
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
			
			client := OrchestratorClient(c)
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

func MetricsDeleteCommand() *cli.Command {
	return &cli.Command{
		Name:      "delete",
		Aliases:   []string{"rm"},
		Usage:     "Delete a metric by ID",
		ArgsUsage: "<metric-id>",
		Action: func(ctx context.Context, c *cli.Command) error {
			if c.Args().Len() < 1 {
				return fmt.Errorf("metric ID required")
			}
			metricID := c.Args().Get(0)
			
			client := OrchestratorClient(c)
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
