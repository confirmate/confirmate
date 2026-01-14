package commands

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/hokaccha/go-prettyjson"
	"github.com/urfave/cli/v3"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// PrettyPrint prints a proto message as pretty-printed JSON to stdout.
func PrettyPrint(msg proto.Message) error {
	// Force colors for now to see if they show up
	color.NoColor = false

	m := protojson.MarshalOptions{
		EmitUnpopulated: false,
	}

	b, err := m.Marshal(msg)
	if err != nil {
		return err
	}

	out, err := prettyjson.Format(b)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(os.Stdout, string(out))
	return err
}

// PaginationFlags returns a slice of common pagination flags.
func PaginationFlags() []cli.Flag {
	return []cli.Flag{
		&cli.IntFlag{
			Name:    "page-size",
			Aliases: []string{"n"},
			Usage:   "Number of items to return",
			Value:   10,
		},
		&cli.StringFlag{
			Name:    "page-token",
			Aliases: []string{"p"},
			Usage:   "Page token for the next page",
		},
	}
}

// PrintJSON prints any object as pretty-printed JSON to stdout.
func PrintJSON(v interface{}) error {
	out, err := prettyjson.Marshal(v)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(os.Stdout, string(out))
	return err
}
