package orchestrator

import (
	"context"

	"github.com/mfridman/cli"
)

// OrchestratorCommand is the command to start the orchestrator server.
var OrchestratorCommand = &cli.Command{
	Name: "orchestrator",
	Exec: func(ctx context.Context, s *cli.State) error {
		return nil
	},
}
