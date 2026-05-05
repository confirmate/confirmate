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
	"fmt"

	"confirmate.io/core/api/evaluation/evaluationconnect"
	"confirmate.io/core/server"
	"confirmate.io/core/service"
	"confirmate.io/core/service/evaluation"

	"connectrpc.com/connect"
	"github.com/urfave/cli/v3"
	"golang.org/x/oauth2/clientcredentials"
)

// evaluationFlags contains the flags that are specific to configuring the evaluation service.
var evaluationFlags = []cli.Flag{
	&cli.StringFlag{
		Name:    "evaluation-orchestrator-address",
		Usage:   "Address of the orchestrator service the evaluation service connects to",
		Value:   evaluation.DefaultOrchestratorURL,
		Sources: envVarSources("evaluation-orchestrator-address"),
	},
}

// EvaluationCommand is the command to start the evaluation server.
var EvaluationCommand = &cli.Command{
	Name:  "evaluation",
	Usage: "Launches the evaluation service",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		var (
			interceptors []connect.Interceptor
			svcOptions   []service.Option[evaluation.Service]
			cfg          evaluation.Config
		)

		cfg = evaluation.Config{
			OrchestratorAddress: cmd.String("evaluation-orchestrator-address"),
			OrchestratorClient:  service.NewHTTPClient(),
		}

		if cmd.Bool("auth-enabled") {
			jwksURL := cmd.String("auth-jwks-url")
			if jwksURL == server.DefaultJWKSURL {
				jwksURL = fmt.Sprintf("http://localhost:%d/v1/auth/certs", cmd.Uint16("api-port"))
			}

			interceptors = append(interceptors, server.NewAuthInterceptor(
				server.WithJWKS(jwksURL),
			))
			svcOptions = append(svcOptions, evaluation.WithAuthorizationStrategyPermissionStore())

			cfg.ServiceOAuth2Config = &clientcredentials.Config{
				ClientID:     cmd.String("service-oauth2-client-id"),
				ClientSecret: cmd.String("service-oauth2-client-secret"),
				TokenURL:     cmd.String("service-oauth2-token-endpoint"),
			}
		}

		interceptors = append(interceptors, &server.LoggingInterceptor{})
		svcOptions = append(svcOptions, evaluation.WithConfig(cfg))

		svc, err := evaluation.NewService(svcOptions...)
		if err != nil {
			return err
		}

		return server.RunConnectServer(
			server.WithConfig(server.Config{
				Port:     cmd.Uint16("api-port"),
				Path:     "/",
				LogLevel: cmd.String("log-level"),
				CORS: server.CORS{
					AllowedOrigins: cmd.StringSlice("api-cors-allowed-origins"),
					AllowedMethods: cmd.StringSlice("api-cors-allowed-methods"),
					AllowedHeaders: cmd.StringSlice("api-cors-allowed-headers"),
				},
			}),
			server.WithHandler(evaluationconnect.NewEvaluationHandler(
				svc,
				connect.WithInterceptors(interceptors...),
			)),
		)
	},
	Flags: joinFlagSlices(
		logFlags,
		apiFlags,
		authFlags,
		serviceAuthFlags,
		evaluationFlags,
	),
}
