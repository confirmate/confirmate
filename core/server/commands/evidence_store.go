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

	"confirmate.io/core/api/evidence/evidenceconnect"
	"confirmate.io/core/persistence"
	"confirmate.io/core/server"
	"confirmate.io/core/service/evidence"

	"net/http"
	"time"

	"connectrpc.com/connect"
	"github.com/urfave/cli/v3"
)

var EvidenceCommand = &cli.Command{
	Name:  "evidence store",
	Usage: "Launches the evidence store service",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		slog.Info("Starting Evidence Store",
			slog.Uint64("api_port", uint64(cmd.Uint16("api-port"))),
			slog.String("log_level", cmd.String("log-level")),
			slog.String("db_host", cmd.String("db-host")),
			slog.Int("db_port", cmd.Int("db-port")),
			slog.String("db_name", cmd.String("db-name")),
			slog.String("db_user", cmd.String("db-user")),
			slog.Bool("db_in_memory", cmd.Bool("db-in-memory")),
			slog.Int("db_max_connections", cmd.Int("db-max-connections")),
			slog.String("assessment_address", cmd.String("assessment-address")),
			slog.Duration("assessment_timeout", cmd.Duration("assessment-timeout")))

		svc, err := evidence.NewService(
			evidence.WithConfig(evidence.Config{
				AssessmentAddress: cmd.String("assessment-address"),
				AssessmentClient: &http.Client{
					Timeout: cmd.Duration("assessment-timeout"),
				},
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
			Name:  "assessment-address",
			Usage: "Assessment service address",
			Value: evidence.DefaultAssessmentURL,
		},
		&cli.DurationFlag{
			Name:  "assessment-timeout",
			Usage: "Assessment HTTP client timeout",
			Value: 30 * time.Second,
		},
	},
}
