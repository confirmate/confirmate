package commands

import (
	"context"
	"net/http"

	"confirmate.io/core/api/evaluation/evaluationconnect"
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
		},
	},
}
