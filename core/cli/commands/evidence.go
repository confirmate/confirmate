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
	"context"

	"confirmate.io/core/api/evidence"

	"connectrpc.com/connect"
	"github.com/urfave/cli/v3"
)

// EvidenceListToolsCommand returns a CLI command that lists all evidence collecting tools that have provided
// evidence to the evidence store, together with the timestamp of their latest evidence.
func EvidenceListToolsCommand() *cli.Command {
	return &cli.Command{
		Name:  "list-tools",
		Usage: "List all evidence collecting tools seen by the evidence store and the time of their latest evidence",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "target",
				Aliases: []string{"t"},
				Usage:   "Filter by target of evaluation ID",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			req := &evidence.ListToolsRequest{}
			if targetID := c.String("target"); targetID != "" {
				req.Filter = &evidence.ListToolsRequest_Filter{TargetOfEvaluationId: &targetID}
			}

			client := EvidenceStoreClient(ctx, c)
			resp, err := client.ListTools(ctx, connect.NewRequest(req))
			if err != nil {
				return err
			}
			return PrettyPrint(resp.Msg)
		},
	}
}
