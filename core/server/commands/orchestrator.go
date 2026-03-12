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
	"log/slog"

	"confirmate.io/core/api/orchestrator/orchestratorconnect"
	"confirmate.io/core/persistence"
	"confirmate.io/core/server"
	"confirmate.io/core/service"
	"confirmate.io/core/service/orchestrator"

	"connectrpc.com/connect"
	"github.com/urfave/cli/v3"
)

// orchestratorFlags contains the flags that are specific to configuring the orchestrator service.
var orchestratorFlags = []cli.Flag{
	&cli.StringFlag{
		Name:    "catalogs-default-path",
		Usage:   "The path to the folder containing default catalog definitions",
		Value:   orchestrator.DefaultConfig.DefaultCatalogsPath,
		Sources: envVarSources("catalogs-default-path"),
	},
	&cli.BoolFlag{
		Name:    "catalogs-load-default",
		Usage:   "Load default catalogs from the catalogs-default-path",
		Value:   orchestrator.DefaultConfig.LoadDefaultCatalogs,
		Sources: envVarSources("catalogs-load-default"),
	},
	&cli.StringFlag{
		Name:    "metrics-default-path",
		Usage:   "The path to the folder containing default metrics (e.g., security-metrics repository)",
		Value:   orchestrator.DefaultConfig.DefaultMetricsPath,
		Sources: envVarSources("metrics-default-path"),
	},
	&cli.BoolFlag{
		Name:    "metrics-load-default",
		Usage:   "Load default metrics from the metrics-default-path",
		Value:   orchestrator.DefaultConfig.LoadDefaultMetrics,
		Sources: envVarSources("metrics-load-default"),
	},
	&cli.BoolFlag{
		Name:    "create-default-target-of-evaluation",
		Usage:   "Creates a default target of evaluation if none exists",
		Value:   orchestrator.DefaultConfig.CreateDefaultTargetOfEvaluation,
		Sources: envVarSources("create-default-target-of-evaluation"),
	},
}

// OrchestratorCommand is the command to start the orchestrator server.
var OrchestratorCommand = &cli.Command{
	Name:  "orchestrator",
	Usage: "Launches the orchestrator service",
	Action: func(ctx context.Context, cmd *cli.Command) (err error) {
		var (
			interceptors []connect.Interceptor
			svcOptions   []service.Option[orchestrator.Service]
			jwksURL      string
			opts         []service.Option[orchestrator.Service]
			svc          orchestratorconnect.OrchestratorHandler
			serverOpts   []server.Option
		)

		if cmd.Bool("auth-enabled") {
			jwksURL = cmd.String("auth-jwks-url")
			if jwksURL == server.DefaultJWKSURL {
				jwksURL = fmt.Sprintf("http://localhost:%d/v1/auth/certs", cmd.Uint16("api-port"))
			}

			interceptors = append(interceptors, server.NewAuthInterceptor(
				server.WithJWKS(jwksURL),
			))
			svcOptions = append(svcOptions, orchestrator.WithAuthorizationStrategyJWT(
				service.DefaultTargetOfEvaluationsClaim,
				service.DefaultAllowAllClaim,
			))
		}

		interceptors = append(interceptors, &server.LoggingInterceptor{})

		opts = append([]service.Option[orchestrator.Service]{
			orchestrator.WithConfig(orchestrator.Config{
				DefaultCatalogsPath:             cmd.String("catalogs-default-path"),
				LoadDefaultCatalogs:             cmd.Bool("catalogs-load-default"),
				DefaultMetricsPath:              cmd.String("metrics-default-path"),
				LoadDefaultMetrics:              cmd.Bool("metrics-load-default"),
				CreateDefaultTargetOfEvaluation: cmd.Bool("create-default-target-of-evaluation"),
				PersistenceConfig: persistence.Config{
					Host:       cmd.String("db-host"),
					Port:       cmd.Int("db-port"),
					DBName:     cmd.String("db-name"),
					User:       cmd.String("db-user-name"),
					Password:   cmd.String("db-password"),
					SSLMode:    cmd.String("db-ssl-mode"),
					InMemoryDB: cmd.Bool("db-in-memory"),
					MaxConn:    cmd.Int("db-max-connections"),
				},
			}),
		}, svcOptions...)

		slog.Info("Starting Orchestrator Service",
			"config", slog.GroupValue(
				slog.Uint64("api_port", uint64(cmd.Uint16("api-port"))),
				slog.String("log_level", cmd.String("log-level")),
				slog.Group("api_cors",
					slog.Any("allowed_origins", cmd.StringSlice("api-cors-allowed-origins")),
					slog.Any("allowed_methods", cmd.StringSlice("api-cors-allowed-methods")),
					slog.Any("allowed_headers", cmd.StringSlice("api-cors-allowed-headers")),
				),
				slog.Group("database",
					slog.String("host", cmd.String("db-host")),
					slog.Int("port", cmd.Int("db-port")),
					slog.String("name", cmd.String("db-name")),
					slog.String("user_name", cmd.String("db-user-name")),
					slog.String("password", cmd.String("db-password")),
					slog.String("sslmode", cmd.String("db-ssl-mode")),
					slog.Bool("in_memory", cmd.Bool("db-in-memory")),
					slog.Int("max_connections", cmd.Int("db-max-connections")),
				),
				slog.Group("catalogs",
					slog.String("default_path", cmd.String("catalogs-default-path")),
					slog.Bool("load_default", cmd.Bool("catalogs-load-default")),
				),
				slog.Group("metrics",
					slog.String("default_path", cmd.String("metrics-default-path")),
					slog.Bool("load_default", cmd.Bool("metrics-load-default")),
				),
				slog.Bool("create_default_target_of_evaluation", cmd.Bool("create-default-target-of-evaluation")),
				slog.Group("auth",
					slog.Bool("enabled", cmd.Bool("auth-enabled")),
					slog.String("jwks_url", cmd.String("auth-jwks-url")),
				),
			),
			"format", "multiline",
		)

		svc, err = orchestrator.NewService(opts...)
		if err != nil {
			return err
		}

		serverOpts = []server.Option{
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
			server.WithHandler(orchestratorconnect.NewOrchestratorHandler(
				svc,
				connect.WithInterceptors(interceptors...),
			)),
		}

		err = server.RunConnectServer(serverOpts...)
		return err
	},
	Flags: joinFlagSlices(
		logFlags,
		apiFlags,
		authFlags,
		dbFlags,
		orchestratorFlags,
	),
}
