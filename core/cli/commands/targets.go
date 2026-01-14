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

func TargetsListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List all targets of evaluation",
		Action: func(ctx context.Context, c *cli.Command) error {
			client := orchestratorconnect.NewOrchestratorClient(http.DefaultClient, "http://localhost:8080")
			resp, err := client.ListTargetsOfEvaluation(ctx, connect.NewRequest(&orchestrator.ListTargetsOfEvaluationRequest{}))
			if err != nil {
				return err
			}
			fmt.Printf("%+v\n", resp.Msg)
			return nil
		},
	}
}

func TargetsGetCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Get a specific target of evaluation by ID",
		ArgsUsage: "<target-id>",
		Action: func(ctx context.Context, c *cli.Command) error {
			if c.NArg() < 1 {
				return fmt.Errorf("target ID required")
			}
			targetID := c.Args().First()
			
			client := orchestratorconnect.NewOrchestratorClient(http.DefaultClient, "http://localhost:8080")
			resp, err := client.GetTargetOfEvaluation(ctx, connect.NewRequest(&orchestrator.GetTargetOfEvaluationRequest{
				TargetOfEvaluationId: targetID,
			}))
			if err != nil {
				return err
			}
			fmt.Printf("%+v\n", resp.Msg)
			return nil
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
			if c.NArg() < 1 {
				return fmt.Errorf("target ID required")
			}
			targetID := c.Args().First()
			
			client := orchestratorconnect.NewOrchestratorClient(http.DefaultClient, "http://localhost:8080")
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
