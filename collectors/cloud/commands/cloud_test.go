// Copyright 2016-2026 Fraunhofer AISEC
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
	"testing"

	cloud "confirmate.io/collectors/cloud/service"
	"confirmate.io/core/util/assert"
	"github.com/urfave/cli/v3"
)

// runCloudCommand parses args against the cloud collector's flags and returns the resulting
// *cli.Command, without invoking the actual collector service.
func runCloudCommand(t *testing.T, args []string) *cli.Command {
	var captured *cli.Command

	cmd := &cli.Command{
		Name:  "cloud-collector",
		Flags: append(append([]cli.Flag{}, cloudCollectorFlags...), cloudStandaloneFlags...),
		Action: func(_ context.Context, cmd *cli.Command) error {
			captured = cmd
			return nil
		},
	}

	err := cmd.Run(context.Background(), append([]string{"cloud-collector"}, args...))
	assert.NoError(t, err)

	return captured
}

func TestCloudServiceOptionsFromCommand(t *testing.T) {
	targetOfEvaluationID := "11111111-1111-1111-1111-111111111111"

	cmd := runCloudCommand(t, []string{
		"--collector-provider", "azure",
		"--target-of-evaluation-id", targetOfEvaluationID,
		"--collector-tool-id", "test-tool-id",
		"--collector-interval", "5",
		"--evidence-store-address", "localhost:9092",
	})

	opts := cloudServiceOptionsFromCommand(cmd, cmd.String("target-of-evaluation-id"))

	svc := cloud.NewService(opts...)

	assert.Equal(t, targetOfEvaluationID, svc.GetTargetOfEvaluationId())
}

func TestCloudServiceOptionsFromCommand_OAuth2Gating(t *testing.T) {
	base := []string{
		"--collector-provider", "azure",
		"--target-of-evaluation-id", "11111111-1111-1111-1111-111111111111",
		"--evidence-store-address", "localhost:9092",
	}

	cmdWithoutOAuth2 := runCloudCommand(t, base)
	optsWithoutOAuth2 := cloudServiceOptionsFromCommand(cmdWithoutOAuth2, cmdWithoutOAuth2.String("target-of-evaluation-id"))

	cmdWithOAuth2 := runCloudCommand(t, append(append([]string{}, base...),
		"--evidence-store-oauth2-enabled",
		"--service-oauth2-client-id", "test-client",
		"--service-oauth2-client-secret", "test-secret",
		"--service-oauth2-token-endpoint", "http://localhost:8080/v1/auth/token",
	))
	optsWithOAuth2 := cloudServiceOptionsFromCommand(cmdWithOAuth2, cmdWithOAuth2.String("target-of-evaluation-id"))

	// Enabling OAuth2 must add exactly one additional option (WithServiceOAuth2Config), and must
	// not be applied at all when the flag is unset.
	assert.Equal(t, len(optsWithoutOAuth2)+1, len(optsWithOAuth2))

	// The resulting service must actually build without error, whether or not OAuth2 is enabled.
	assert.NotNil(t, cloud.NewService(optsWithoutOAuth2...))
	assert.NotNil(t, cloud.NewService(optsWithOAuth2...))
}
