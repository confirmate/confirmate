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
	"os/signal"
	"syscall"
	"time"

	cloud "confirmate.io/collectors/cloud/service"
	"confirmate.io/core/service"
	"github.com/urfave/cli/v3"
	"golang.org/x/oauth2/clientcredentials"
)

const (
	// DefaultServiceOAuth2TokenEndpoint is the default OAuth 2.0 token URL for service-to-service
	// auth with the evidence store, matching the default confirmate auth server.
	DefaultServiceOAuth2TokenEndpoint = "http://localhost:8080/v1/auth/token"
	// DefaultServiceOAuth2ClientID is the default OAuth 2.0 client ID for service-to-service auth.
	DefaultServiceOAuth2ClientID = "confirmate"
	// DefaultServiceOAuth2ClientSecret is the default OAuth 2.0 client secret for service-to-service auth.
	DefaultServiceOAuth2ClientSecret = "confirmate"
)

var cloudCollectorFlags = []cli.Flag{
	&cli.StringFlag{
		Name:     "collector-provider",
		Aliases:  []string{"p"},
		Usage:    "Cloud provider (aws, azure, openstack, k8s, csaf)",
		Required: true,
	},
	&cli.StringFlag{
		Name:     "collector-tool-id",
		Aliases:  []string{"t"},
		Usage:    "Collector Tool ID to identify the collector instance",
		Required: false,
	},
	&cli.StringFlag{
		Name:     "collector-resource-group",
		Aliases:  []string{"r"},
		Usage:    "Limit the scope of the collector to a specific resource group.",
		Required: false,
	},
	&cli.StringFlag{
		Name:     "collector-csaf-domain",
		Aliases:  []string{"d"},
		Usage:    "CSAF domain to fetch the CSAF documents from.",
		Required: false,
	},
}

var cloudStandaloneFlags = []cli.Flag{
	&cli.StringFlag{
		Name:     "target-of-evaluation-id",
		Aliases:  []string{"e"},
		Usage:    "Target of evaluation ID for which to collect the cloud evidence",
		Required: false,
	},
	&cli.StringFlag{
		Name:     "collector-collector",
		Aliases:  []string{"c"},
		Usage:    "Additional collector to use.",
		Required: false,
	},
	&cli.IntFlag{
		Name:     "collector-interval",
		Aliases:  []string{"i"},
		Usage:    "Interval in minutes for periodic collection. (Default: 5 minutes)",
		Required: false,
	},
	&cli.BoolFlag{
		Name:     "collector-auto-start",
		Aliases:  []string{"a"},
		Usage:    "Collector starts automatically after launching the service. (Default: false)",
		Required: false,
	},
	&cli.StringFlag{
		Name:     "collector-evidence-store-address",
		Aliases:  []string{"s"},
		Usage:    "Address of the evidence store to send collected evidence to. (default: localhost:9092)",
		Required: false,
	},
	&cli.BoolFlag{
		Name:     "collector-service-oauth2-enabled",
		Usage:    "Authenticate the connection to the evidence store using OAuth 2.0 client credentials. (Default: false)",
		Required: false,
	},
	&cli.StringFlag{
		Name:     "collector-service-oauth2-token-endpoint",
		Usage:    "OAuth 2.0 token URL for service-to-service auth with the evidence store",
		Value:    DefaultServiceOAuth2TokenEndpoint,
		Required: false,
	},
	&cli.StringFlag{
		Name:     "collector-service-oauth2-client-id",
		Usage:    "OAuth 2.0 client ID for service-to-service auth with the evidence store",
		Value:    DefaultServiceOAuth2ClientID,
		Required: false,
	},
	&cli.StringFlag{
		Name:     "collector-service-oauth2-client-secret",
		Usage:    "OAuth 2.0 client secret for service-to-service auth with the evidence store",
		Value:    DefaultServiceOAuth2ClientSecret,
		Required: false,
	},
}

func cloudServiceOptionsFromCommand(cmd *cli.Command, targetOfEvaluationID string) (opts []service.Option[cloud.Service]) {
	if cmd.String("collector-provider") != "" {
		opts = append(opts, cloud.WithProvider(cmd.String("collector-provider")))
	}
	if targetOfEvaluationID != "" {
		opts = append(opts, cloud.WithTargetOfEvaluationID(targetOfEvaluationID))
	}
	if cmd.String("collector-tool-id") != "" {
		opts = append(opts, cloud.WithCollectorToolID(cmd.String("collector-tool-id")))
	}
	if cmd.Int("collector-interval") != 0 {
		opts = append(opts, cloud.WithCollectorInterval(time.Duration(cmd.Int("collector-interval"))*time.Minute))
	}
	if cmd.String("collector-evidence-store-address") != "" {
		opts = append(opts, cloud.WithEvidenceStoreAddress(cmd.String("collector-evidence-store-address"), service.DefaultHTTPClient))
	}
	if cmd.Bool("collector-service-oauth2-enabled") {
		opts = append(opts, cloud.WithServiceOAuth2Config(&clientcredentials.Config{
			ClientID:     cmd.String("collector-service-oauth2-client-id"),
			ClientSecret: cmd.String("collector-service-oauth2-client-secret"),
			TokenURL:     cmd.String("collector-service-oauth2-token-endpoint"),
		}))
	}

	return opts
}

var CloudCollectorCommand = &cli.Command{
	Name:  "cloud-collector",
	Usage: "Launches one cloud collector service independently",
	Flags: append(append([]cli.Flag{}, cloudCollectorFlags...), cloudStandaloneFlags...),
	Action: func(ctx context.Context, cmd *cli.Command) error {
		var (
			svc  *cloud.Service
			opts []service.Option[cloud.Service]
		)

		opts = cloudServiceOptionsFromCommand(cmd, cmd.String("target-of-evaluation-id"))

		svc = cloud.NewService(opts...)
		svc.Init(ctx, cmd)

		// Signal-Context (blocks bis SIGINT/SIGTERM)
		sigCtx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		<-sigCtx.Done() // Wait until signal

		return nil
	},
}
