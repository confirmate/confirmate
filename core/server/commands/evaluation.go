package commands

import (
	"context"
	"net/http"

	"confirmate.io/core/api/evaluation/evaluationconnect"
	"confirmate.io/core/persistence"
	"confirmate.io/core/server"
	"confirmate.io/core/service/evaluation"

	"connectrpc.com/connect"
	"github.com/urfave/cli/v3"
)

// EvaluationCommand is the command to start the evaluation server.
var EvaluationCommand = &cli.Command{
	Name:  "evaluation",
	Usage: "Launches the evaluation service",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		svc, err := evaluation.NewService(
			evaluation.WithConfig(evaluation.Config{
				OrchestratorAddress: cmd.String("evaluation-orchestrator-address"),
				OrchestratorClient:  http.DefaultClient,
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
			Name:  "evaluation-orchestrator-address",
			Usage: "Address of the orchestrator service the evaluation service connects to",
			Value: evaluation.DefaultOrchestratorURL,
		},
	},
}
