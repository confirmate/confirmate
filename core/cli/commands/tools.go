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

func ToolsListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List all assessment tools",
		Action: func(ctx context.Context, c *cli.Command) error {
			client := orchestratorconnect.NewOrchestratorClient(http.DefaultClient, "http://localhost:8080")
			resp, err := client.ListAssessmentTools(ctx, connect.NewRequest(&orchestrator.ListAssessmentToolsRequest{}))
			if err != nil {
				return err
			}
			fmt.Printf("%+v\n", resp.Msg)
			return nil
		},
	}
}

func ToolsGetCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Get a specific assessment tool by ID",
		ArgsUsage: "<tool-id>",
		Action: func(ctx context.Context, c *cli.Command) error {
			if c.NArg() < 1 {
				return fmt.Errorf("tool ID required")
			}
			toolID := c.Args().First()
			
			client := orchestratorconnect.NewOrchestratorClient(http.DefaultClient, "http://localhost:8080")
			resp, err := client.GetAssessmentTool(ctx, connect.NewRequest(&orchestrator.GetAssessmentToolRequest{
				ToolId: toolID,
			}))
			if err != nil {
				return err
			}
			fmt.Printf("%+v\n", resp.Msg)
			return nil
		},
	}
}

func ToolsDeleteCommand() *cli.Command {
	return &cli.Command{
		Name:      "delete",
		Aliases:   []string{"rm"},
		Usage:     "Delete an assessment tool by ID",
		ArgsUsage: "<tool-id>",
		Action: func(ctx context.Context, c *cli.Command) error {
			if c.NArg() < 1 {
				return fmt.Errorf("tool ID required")
			}
			toolID := c.Args().First()
			
			client := orchestratorconnect.NewOrchestratorClient(http.DefaultClient, "http://localhost:8080")
			_, err := client.DeregisterAssessmentTool(ctx, connect.NewRequest(&orchestrator.DeregisterAssessmentToolRequest{
				ToolId: toolID,
			}))
			if err != nil {
				return err
			}
			fmt.Printf("Tool %s deleted successfully\n", toolID)
			return nil
		},
	}
}
