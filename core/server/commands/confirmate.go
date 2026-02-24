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
	"net/http"

	"confirmate.io/core/api/assessment/assessmentconnect"
	"confirmate.io/core/api/orchestrator/orchestratorconnect"
	"confirmate.io/core/persistence"
	"confirmate.io/core/server"
	"confirmate.io/core/service/assessment"
	"confirmate.io/core/service/orchestrator"

	"connectrpc.com/connect"
	"github.com/urfave/cli/v3"
)

// ConfirmateCommand starts the full framework: orchestrator and assessment services on one server.
var ConfirmateCommand = &cli.Command{
	Name:  "confirmate",
	Usage: "Launches the confirmate framework (including orchestrator and assessment services)",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		var (
			svc             orchestratorconnect.OrchestratorHandler
			err             error
			apiPort         uint16
			orchestratorURL string
			assessmentSvc   assessmentconnect.AssessmentHandler
			path            string
			handler         http.Handler
			handlers        map[string]http.Handler
		)

		svc, err = orchestrator.NewService(
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
					User:       cmd.String("db-user"),
					Password:   cmd.String("db-password"),
					SSLMode:    cmd.String("db-sslmode"),
					InMemoryDB: cmd.Bool("db-in-memory"),
					MaxConn:    cmd.Int("db-max-connections"),
				},
			}),
		)
		if err != nil {
			return err
		}

		apiPort = cmd.Uint16("api-port")
		orchestratorURL = fmt.Sprintf("http://localhost:%d", apiPort)

		assessmentSvc, err = assessment.NewService(
			assessment.WithConfig(assessment.Config{
				OrchestratorAddress: orchestratorURL,
				OrchestratorClient:  http.DefaultClient,
				RegoPackage:         cmd.String("assessment-rego-package"),
			}),
		)
		if err != nil {
			return err
		}

		handlers = make(map[string]http.Handler)
		path, handler = orchestratorconnect.NewOrchestratorHandler(
			svc,
			connect.WithInterceptors(&server.LoggingInterceptor{}),
		)
		handlers[path] = handler
		path, handler = assessmentconnect.NewAssessmentHandler(assessmentSvc)
		handlers[path] = handler

		return server.RunConnectServer(
			server.WithConfig(server.Config{
				Port:     apiPort,
				Path:     "/",
				LogLevel: cmd.String("log-level"),
				CORS: server.CORS{
					AllowedOrigins: cmd.StringSlice("api-cors-allowed-origins"),
					AllowedMethods: cmd.StringSlice("api-cors-allowed-methods"),
					AllowedHeaders: cmd.StringSlice("api-cors-allowed-headers"),
				},
				Handlers: handlers,
			}),
		)
	},
	Flags: []cli.Flag{
		&cli.Uint16Flag{
			Name:  "api-port",
			Usage: "Port to run the API server (Connect, gRPC, REST) on",
			Value: server.DefaultConfig.Port,
		},
		&cli.StringFlag{
			Name:  "log-level",
			Usage: "Log level (TRACE, DEBUG, INFO, WARN, ERROR)",
			Value: server.DefaultConfig.LogLevel,
		},
		&cli.StringSliceFlag{
			Name:  "api-cors-allowed-origins",
			Usage: "Specifies the origins allowed in CORS",
			Value: server.DefaultConfig.CORS.AllowedOrigins,
		},
		&cli.StringSliceFlag{
			Name:  "api-cors-allowed-methods",
			Usage: "Specifies the methods allowed in CORS",
			Value: server.DefaultConfig.CORS.AllowedMethods,
		},
		&cli.StringSliceFlag{
			Name:  "api-cors-allowed-headers",
			Usage: "Specifies the headers allowed in CORS",
			Value: server.DefaultConfig.CORS.AllowedHeaders,
		},
		&cli.StringFlag{
			Name:  "catalogs-default-path",
			Usage: "The path to the folder containing default catalog definitions",
			Value: orchestrator.DefaultConfig.DefaultCatalogsPath,
		},
		&cli.BoolFlag{
			Name:  "catalogs-load-default",
			Usage: "Load default catalogs from the catalogs-default-path",
			Value: orchestrator.DefaultConfig.LoadDefaultCatalogs,
		},
		&cli.StringFlag{
			Name:  "metrics-default-path",
			Usage: "The path to the folder containing default metrics (e.g., security-metrics repository)",
			Value: orchestrator.DefaultConfig.DefaultMetricsPath,
		},
		&cli.BoolFlag{
			Name:  "metrics-load-default",
			Usage: "Load default metrics from the metrics-default-path",
			Value: orchestrator.DefaultConfig.LoadDefaultMetrics,
		},
		&cli.BoolFlag{
			Name:  "create-default-target-of-evaluation",
			Usage: "Creates a default target of evaluation if none exists",
			Value: orchestrator.DefaultConfig.CreateDefaultTargetOfEvaluation,
		},
		&cli.StringFlag{
			Name:  "db-host",
			Usage: "Specifies the server hostname",
			Value: persistence.DefaultConfig.Host,
		},
		&cli.IntFlag{
			Name:  "db-port",
			Usage: "Specifies the server port",
			Value: persistence.DefaultConfig.Port,
		},
		&cli.StringFlag{
			Name:  "db-name",
			Usage: "Specifies the database name",
			Value: persistence.DefaultConfig.DBName,
		},
		&cli.StringFlag{
			Name:  "db-user",
			Usage: "Specifies the database user",
			Value: persistence.DefaultConfig.User,
		},
		&cli.StringFlag{
			Name:  "db-password",
			Usage: "Specifies the database password",
			Value: persistence.DefaultConfig.Password,
		},
		&cli.StringFlag{
			Name:  "db-sslmode",
			Usage: "Specifies the database SSL mode (disable, require, verify-ca, verify-full)",
			Value: persistence.DefaultConfig.SSLMode,
		},
		&cli.BoolFlag{
			Name:  "db-in-memory",
			Usage: "Use in-memory database instead of PostgreSQL (useful for testing)",
			Value: persistence.DefaultConfig.InMemoryDB,
		},
		&cli.IntFlag{
			Name:  "db-max-connections",
			Usage: "Specifies the maximum number of database connections",
			Value: persistence.DefaultConfig.MaxConn,
		},
		&cli.StringFlag{
			Name:  "assessment-rego-package",
			Usage: "Rego package to use for assessments",
			Value: assessment.DefaultConfig.RegoPackage,
		},
	},
}
