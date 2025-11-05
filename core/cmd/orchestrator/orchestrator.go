package main

import (
	"confirmate.io/core/server/commands"
)

func main() {
	commands.ParseAndRun(commands.OrchestratorCommand)
}
