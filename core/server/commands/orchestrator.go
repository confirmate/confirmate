package commands

import (
	"context"

	"confirmate.io/core/api/orchestrator/orchestratorconnect"
	"confirmate.io/core/server"
	"confirmate.io/core/service/orchestrator"

	"github.com/mfridman/cli"
)

// OrchestratorCommand is the command to start the orchestrator server.
var OrchestratorCommand = &cli.Command{
	Name: "orchestrator",
	Exec: func(ctx context.Context, s *cli.State) error {
		svc, err := orchestrator.NewService()
		if err != nil {
			return err
		}

		return server.RunConnectServer(orchestratorconnect.NewOrchestratorHandler(svc))
	},
}
