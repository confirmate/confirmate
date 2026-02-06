package main

import (
	"log/slog"
	"os"

	"confirmate.io/core/log"
	"confirmate.io/core/server/commands"
)

func main() {
	if err := commands.ParseAndRun(commands.OrchestratorCommand); err != nil {
		slog.Error("Failed to start orchestrator", log.Err(err))
		os.Exit(1)
	}
}
