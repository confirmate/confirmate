package main

import (
	"context"
	"log"
	"os"

	"confirmate.io/core/cli/commands"
)

func main() {
	cmd := commands.NewRootCommand()

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
