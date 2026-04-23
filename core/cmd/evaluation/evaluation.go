package main

import (
	"log/slog"
	"os"

	"confirmate.io/core/log"
	"confirmate.io/core/server/commands"
)

func main() {
	if err := commands.ParseAndRun(commands.EvaluationCommand); err != nil {
		slog.Error("Failed to start evaluation", log.Err(err))
		os.Exit(1)
	}
}
