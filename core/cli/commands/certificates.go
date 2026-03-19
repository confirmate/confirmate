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
	"fmt"

	"confirmate.io/core/api/orchestrator"

	"connectrpc.com/connect"
	"github.com/urfave/cli/v3"
)

func CertificatesListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List all compliance attestations",
		Flags: PaginationFlags(),
		Action: func(ctx context.Context, c *cli.Command) error {
			client := OrchestratorClient(ctx, c)
			resp, err := client.ListComplianceAttestations(ctx, connect.NewRequest(&orchestrator.ListComplianceAttestationsRequest{
				PageSize:  int32(c.Int("page-size")),
				PageToken: c.String("page-token"),
			}))
			if err != nil {
				return err
			}
			return PrettyPrint(resp.Msg)
		},
	}
}

func CertificatesListPublicCommand() *cli.Command {
	return &cli.Command{
		Name:    "list-public",
		Aliases: []string{"public"},
		Usage:   "List all public compliance attestations",
		Flags:   PaginationFlags(),
		Action: func(ctx context.Context, c *cli.Command) error {
			client := OrchestratorClient(ctx, c)
			resp, err := client.ListPublicComplianceAttestations(ctx, connect.NewRequest(&orchestrator.ListPublicComplianceAttestationsRequest{
				PageSize:  int32(c.Int("page-size")),
				PageToken: c.String("page-token"),
			}))
			if err != nil {
				return err
			}
			return PrettyPrint(resp.Msg)
		},
	}
}

func CertificatesGetCommand() *cli.Command {
	return &cli.Command{
		Name:      "get",
		Usage:     "Get a specific compliance attestation by ID",
		ArgsUsage: "<compliance-attestation-id>",
		Action: func(ctx context.Context, c *cli.Command) error {
			if c.Args().Len() < 1 {
				return fmt.Errorf("compliance attestation ID required")
			}
			certID := c.Args().Get(0)

			client := OrchestratorClient(ctx, c)
			resp, err := client.GetComplianceAttestation(ctx, connect.NewRequest(&orchestrator.GetComplianceAttestationRequest{
				ComplianceAttestationId: certID,
			}))
			if err != nil {
				return err
			}
			return PrettyPrint(resp.Msg)
		},
	}
}

func CertificatesRemoveCommand() *cli.Command {
	return &cli.Command{
		Name:      "remove",
		Aliases:   []string{"rm"},
		Usage:     "Delete a compliance attestation by ID",
		ArgsUsage: "<compliance-attestation-id>",
		Action: func(ctx context.Context, c *cli.Command) error {
			if c.Args().Len() < 1 {
				return fmt.Errorf("compliance attestation ID required")
			}
			certID := c.Args().Get(0)

			client := OrchestratorClient(ctx, c)
			_, err := client.RemoveComplianceAttestation(ctx, connect.NewRequest(&orchestrator.RemoveComplianceAttestationRequest{
				ComplianceAttestationId: certID,
			}))
			if err != nil {
				return err
			}
			fmt.Printf("Compliance attestation %s deleted successfully\n", certID)
			return nil
		},
	}
}
