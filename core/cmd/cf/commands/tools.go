package commands
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
