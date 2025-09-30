package commands

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/mfridman/cli"
)

// ParseAndRun parses the command line arguments and runs the given command.
// If an error occurs, it is printed to stderr and the program exits with a non-zero
// status code.
// If the help flag is provided, the usage information is printed to stdout
// and the function returns without error.
func ParseAndRun(cmd *cli.Command) error {
	if err := cli.Parse(cmd, os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fmt.Fprintf(os.Stdout, "%s\n", cli.DefaultUsage(cmd))
			return nil
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if err := cli.Run(context.Background(), cmd, nil); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	return nil
}
