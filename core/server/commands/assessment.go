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

	"confirmate.io/core/api/assessment/assessmentconnect"
	"confirmate.io/core/server"
	"confirmate.io/core/service"
	"confirmate.io/core/service/assessment"
	"golang.org/x/oauth2/clientcredentials"

	"connectrpc.com/connect"
	"github.com/urfave/cli/v3"
)

// assessmentFlags contains the flags that are specific to configuring the assessment service.
var assessmentFlags = []cli.Flag{
	&cli.StringFlag{
		Name:    "assessment-orchestrator-address",
		Usage:   "Address of the orchestrator service the assessment service connects to",
		Value:   assessment.DefaultOrchestratorURL,
		Sources: envVarSources("assessment-orchestrator-address"),
	},
	&cli.StringFlag{
		Name:    "assessment-rego-package",
		Usage:   "Rego package to use for assessments",
		Value:   assessment.DefaultConfig.RegoPackage,
		Sources: envVarSources("assessment-rego-package"),
	},
}

// AssessmentCommand is the command to start the assessment server.
var AssessmentCommand = &cli.Command{
	Name:  "assessment",
	Usage: "Launches the assessment service",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		var (
			interceptors []connect.Interceptor
			svcOptions   []service.Option[assessment.Service]
			cfg          assessment.Config
		)

		cfg = assessment.Config{
			OrchestratorAddress:    cmd.String("assessment-orchestrator-address"),
			OrchestratorHTTPClient: service.NewHTTPClient(),
			RegoPackage:            cmd.String("assessment-rego-package"),
		}

		if cmd.Bool("auth-enabled") {
			jwksURL := cmd.String("auth-jwks-url")
			if jwksURL == server.DefaultJWKSURL {
				jwksURL = fmt.Sprintf("http://localhost:%d/v1/auth/certs", cmd.Uint16("api-port"))
			}
			interceptors = append(interceptors, server.NewAuthInterceptor(authInterceptorOptions(cmd, jwksURL)...))
			svcOptions = append(svcOptions, assessment.WithAuthorizationStrategyPermissionStore())

			cfg.ServiceOAuth2Config = &clientcredentials.Config{
				ClientID:     cmd.String("service-oauth2-client-id"),
				ClientSecret: cmd.String("service-oauth2-client-secret"),
				TokenURL:     cmd.String("service-oauth2-token-endpoint"),
			}
		}

		interceptors = append(interceptors, &server.LoggingInterceptor{})
		svcOptions = append(svcOptions, assessment.WithConfig(cfg))

		svc, err := assessment.NewService(svcOptions...)
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
			server.WithHandler(assessmentconnect.NewAssessmentHandler(
				svc,
				connect.WithInterceptors(interceptors...),
			)),
			server.WithReflection(),
		)
	},
	Flags: joinFlagSlices(
		logFlags,
		apiFlags,
		authFlags,
		serviceAuthFlags,
		assessmentFlags,
	),
}
