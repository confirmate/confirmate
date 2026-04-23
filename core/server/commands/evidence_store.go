// Copyright 2016-2025 Fraunhofer AISEC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package commands

import (
	"context"
	"log/slog"

	"time"

	"confirmate.io/core/api/evidence/evidenceconnect"
	"confirmate.io/core/persistence"
	"confirmate.io/core/server"
	"confirmate.io/core/service"
	"confirmate.io/core/service/evidence"

	"connectrpc.com/connect"
	"github.com/urfave/cli/v3"
)

// evidenceFlags contains the flags that are specific to configuring the evidence store service.
var evidenceFlags = []cli.Flag{
	&cli.StringFlag{
		Name:    "evidence-assessment-address",
		Usage:   "Address of the assessment service the evidence store connects to",
		Value:   evidence.DefaultAssessmentURL,
		Sources: envVarSources("evidence-assessment-address"),
	},
	&cli.DurationFlag{
		Name:    "evidence-assessment-http-timeout",
		Usage:   "Assessment HTTP client timeout",
		Value:   30 * time.Second,
		Sources: envVarSources("evidence-assessment-http-timeout"),
	},
}

// EvidenceCommand is the command to start the evidence store server.
var EvidenceCommand = &cli.Command{
	Name:  "evidence",
	Usage: "Launches the evidence store service",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		slog.Info("Starting Evidence Store",
			slog.Uint64("api_port", uint64(cmd.Uint16("api-port"))),
			slog.String("log_level", cmd.String("log-level")),
			slog.Any("api_cors_allowed_origins", cmd.StringSlice("api-cors-allowed-origins")),
			slog.Any("api_cors_allowed_methods", cmd.StringSlice("api-cors-allowed-methods")),
			slog.Any("api_cors_allowed_headers", cmd.StringSlice("api-cors-allowed-headers")),
			slog.String("db_host", cmd.String("db-host")),
			slog.Int("db_port", cmd.Int("db-port")),
			slog.String("db_name", cmd.String("db-name")),
			slog.String("db_user_name", cmd.String("db-user-name")),
			slog.String("db_password", cmd.String("db-password")),
			slog.String("db_sslmode", cmd.String("db-ssl-mode")),
			slog.Bool("db_in_memory", cmd.Bool("db-in-memory")),
			slog.Int("db_max_connections", cmd.Int("db-max-connections")),
			slog.String("assessment_address", cmd.String("evidence-assessment-address")),
			slog.Duration("assessment_timeout", cmd.Duration("evidence-assessment-http-timeout")))

		assessmentClient := service.NewHTTPClient()
		assessmentClient.Timeout = cmd.Duration("evidence-assessment-http-timeout")

		svc, err := evidence.NewService(
			evidence.WithConfig(evidence.Config{
				AssessmentAddress:    cmd.String("evidence-assessment-address"),
				AssessmentHTTPClient: assessmentClient,
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
				EvidenceQueueSize: evidence.DefaultConfig.EvidenceQueueSize,
			}),
		)
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
			server.WithHandler(evidenceconnect.NewEvidenceStoreHandler(
				svc,
				connect.WithInterceptors(&server.LoggingInterceptor{}),
			)),
			server.WithHandler(evidenceconnect.NewResourcesHandler(
				svc,
				connect.WithInterceptors(&server.LoggingInterceptor{}),
			)),
			server.WithReflection(),
		)
	},
	Flags: joinFlagSlices(
		logFlags,
		apiFlags,
		authFlags,
		serviceAuthFlags,
		dbFlags,
		evidenceFlags,
	),
}
