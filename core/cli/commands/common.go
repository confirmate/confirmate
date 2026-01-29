// Copyright 2016-2025 Fraunhofer AISEC
//
// SPDX-License-Identifier: Apache-2.0
//
//                                 /$$$$$$  /$$                                     /$$
//                               /$$__  $$|__/                                    | $$
//   /$$$$$$$  /$$$$$$  /$$$$$$$ | $$  \__/ /$$  /$$$$$$  /$$$$$$/$$$$   /$$$$$$  /$$$$$$    /$$$$$$
//  /$$_____/ /$$__  $$| $$__  $$| $$$$    | $$ /$$__  $$| $$_  $$_  $$ |____  $$|_  $$_/   /$$__  $$
// | $$      | $$  \ $$| $$  \ $$| $$_/    | $$| $$  \__/| $$ \ $$ \ $$  /$$$$$$$  | $$    | $$$$$$$$
// | $$      | $$  | $$| $$  | $$| $$      | $$| $$      | $$ | $$ | $$ /$$__  $$  | $$ /$$| $$_____/
// |  $$$$$$$|  $$$$$$/| $$  | $$| $$      | $$| $$      | $$ | $$ | $$|  $$$$$$$  |  $$$$/|  $$$$$$$
// \_______/ \______/ |__/  |__/|__/      |__/|__/      |__/ |__/ |__/ \_______/   \___/   \_______/
//
// This file is part of Confirmate Core.

package commands

import (
	"fmt"
	"net/http"
	"os"

	"github.com/fatih/color"
	"github.com/hokaccha/go-prettyjson"
	"github.com/urfave/cli/v3"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"confirmate.io/core/api/orchestrator/orchestratorconnect"
)

// OrchestratorClient returns an orchestrator client based on the addr flag.
func OrchestratorClient(c *cli.Command) orchestratorconnect.OrchestratorClient {
	return orchestratorconnect.NewOrchestratorClient(http.DefaultClient, c.Root().String("addr"))
}

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
