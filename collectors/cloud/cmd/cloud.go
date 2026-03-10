package main

import (
	"confirmate.io/collectors/cloud/commands"
	core_commands "confirmate.io/core/server/commands"
)

func main() {
	core_commands.ParseAndRun(commands.CloudCollectorCommand)
}
