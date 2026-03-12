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
	"net/http"

	"confirmate.io/core/api/assessment/assessmentconnect"
	"confirmate.io/core/server"
	"confirmate.io/core/service/assessment"

	"connectrpc.com/connect"
	"connectrpc.com/grpcreflect"
	"github.com/urfave/cli/v3"
)

// assessmentFlags contains the flags that are specific to configuring the assessment service.
var assessmentFlags = []cli.Flag{
	&cli.StringFlag{
		Name:    "assessment-orchestrator-address",
		Usage:   "Address of the orchestrator service the assessment service connects to",
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
			reflector         *grpcreflect.Reflector
			reflectionV1Path  string
			reflectionV1      http.Handler
			reflectionV1APath string
			reflectionV1A     http.Handler
		)

		svc, err := assessment.NewService(
			assessment.WithConfig(assessment.Config{
				OrchestratorAddress: cmd.String("assessment-orchestrator-address"),
				OrchestratorClient:  http.DefaultClient,
				RegoPackage:         cmd.String("assessment-rego-package"),
			}),
		)
		if err != nil {
			return err
		}

		// Add reflector for gRPC reflection, which allows clients to query the server for its supported services and methods.
		reflector = grpcreflect.NewStaticReflector(
			assessmentconnect.AssessmentName,
		)
		reflectionV1Path, reflectionV1 = grpcreflect.NewHandlerV1(reflector)
		reflectionV1APath, reflectionV1A = grpcreflect.NewHandlerV1Alpha(reflector)

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
				connect.WithInterceptors(&server.LoggingInterceptor{}),
			)),
			server.WithHTTPHandler(reflectionV1Path, reflectionV1),
			server.WithHTTPHandler(reflectionV1APath, reflectionV1A),
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
