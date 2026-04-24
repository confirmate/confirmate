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
	"net"
	"net/http"
	"time"

	"confirmate.io/core/api"
	"confirmate.io/core/api/assessment/assessmentconnect"
	"confirmate.io/core/api/evidence/evidenceconnect"
	"confirmate.io/core/api/orchestrator/orchestratorconnect"
	"confirmate.io/core/persistence"
	"confirmate.io/core/server"
	"confirmate.io/core/service"
	"confirmate.io/core/service/assessment"
	"confirmate.io/core/service/collection"
	"confirmate.io/core/service/evidence"
	"confirmate.io/core/service/orchestrator"

	"connectrpc.com/connect"
	"github.com/urfave/cli/v3"
	"golang.org/x/oauth2/clientcredentials"
)

const (
	DefaultServiceTokenURL              = "http://localhost:8080/v1/auth/token"
	DefaultServiceClientID              = "confirmate"
	DefaultServiceClientSecret          = "confirmate"
	defaultEmbeddedTargetOfEvaluationID = "00000000-0000-0000-0000-000000000000"
)

// oauthServerFlags contains the flags for configuring the embedded OAuth 2.0 server.
var oauthServerFlags = []cli.Flag{
	&cli.BoolFlag{
		Name:    "oauth2-embedded",
		Usage:   "Enable embedded OAuth 2.0 server",
		Value:   true,
		Sources: envVarSources("oauth2-embedded"),
	},
	&cli.StringFlag{
		Name:    "oauth2-public-url",
		Usage:   "Public base URL for the embedded OAuth 2.0 server",
		Value:   "",
		Sources: envVarSources("oauth2-public-url"),
	},
	&cli.StringFlag{
		Name:    "oauth2-key-path",
		Usage:   "Path to the OAuth 2.0 signing key",
		Value:   server.DefaultOAuth2KeyPath,
		Sources: envVarSources("oauth2-key-path"),
	},
	&cli.StringFlag{
		Name:    "oauth2-key-password",
		Usage:   "Password for the OAuth 2.0 signing key",
		Value:   server.DefaultOAuth2KeyPassword,
		Sources: envVarSources("oauth2-key-password"),
	},
	&cli.BoolFlag{
		Name:    "oauth2-key-save-on-create",
		Usage:   "Persist generated OAuth 2.0 signing keys",
		Value:   server.DefaultOAuth2KeySaveOnCreate,
		Sources: envVarSources("oauth2-key-save-on-create"),
	},
}

// ConfirmateCommand starts the full framework: orchestrator,  assessment, evidence store, and collector services on one server.
var ConfirmateCommand = &cli.Command{
	Name:  "confirmate",
	Usage: "Launches the confirmate framework (including orchestrator, assessment, evidence store, and collector services)",
	Action: func(ctx context.Context, cmd *cli.Command) (err error) {
		return runConfirmate(ctx, cmd, []collection.Collector{newNoOpCollector("confirmate-no-op-collector")}, "")
	},
	Flags: joinFlagSlices(
		logFlags,
		apiFlags,
		authFlags,
		serviceAuthFlags,
		newDBFlags(true),
		assessmentFlags,
		evidenceFlags,
		oauthServerFlags,
		orchestratorFlags,
		collectionFlags,
	),
}

// runConfirmate starts the embedded Confirmate stack with the provided collectors.
func runConfirmate(ctx context.Context, cmd *cli.Command, collectors []collection.Collector, collectionToolID string) (err error) {
	var (
		interceptors         []connect.Interceptor
		orchestratorOptions  []service.Option[orchestrator.Service]
		assessmentOptions    []service.Option[assessment.Service]
		evidenceOptions      []service.Option[evidence.Service]
		jwksURL              string
		orchestratorOpts     []service.Option[orchestrator.Service]
		assessmentOpts       []service.Option[assessment.Service]
		evidenceOpts         []service.Option[evidence.Service]
		orchestratorSvc      orchestratorconnect.OrchestratorHandler
		assessmentSvc        assessmentconnect.AssessmentHandler
		evidenceSvc          evidenceconnect.EvidenceStoreHandler
		orchestratorClient   *http.Client
		collectionHTTPClient *http.Client
		apiPort              uint16
		credentials          *clientcredentials.Config
		authorizer           api.Authorizer
		collectionSvc        *collection.Service
		collectionResults    <-chan collection.CollectionResult
		serverOpts           []server.Option
		srv                  *server.Server
		serverErrCh          chan error
		targetOfEvaluationID string
		evidenceStoreAddress string
	)

	if cmd.Bool("auth-enabled") {
		jwksURL = cmd.String("auth-jwks-url")
		if jwksURL == server.DefaultJWKSURL {
			jwksURL = fmt.Sprintf("http://localhost:%d/v1/auth/certs", cmd.Uint16("api-port"))
		}

		// Configure authentication interceptor for all services and authorization strategy for services based on JWT claims
		interceptors = append(interceptors, server.NewAuthInterceptor(
			server.WithJWKS(jwksURL),
		))
		orchestratorOptions = append(orchestratorOptions, orchestrator.WithAuthorizationStrategyPermissionStore())
		assessmentOptions = append(assessmentOptions, assessment.WithAuthorizationStrategyPermissionStore())
	}

	interceptors = append(interceptors, &server.LoggingInterceptor{})

	// Orchestrator service configuration
	orchestratorOpts = append([]service.Option[orchestrator.Service]{
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
	}, orchestratorOptions...)

	orchestratorSvc, err = orchestrator.NewService(orchestratorOpts...)
	if err != nil {
		return err
	}
	apiPort = cmd.Uint16("api-port")

	orchestratorClient = service.NewHTTPClient()
	collectionHTTPClient = service.NewHTTPClient()
	if cmd.Bool("auth-enabled") {
		credentials = &clientcredentials.Config{
			ClientID:     cmd.String("service-oauth2-client-id"),
			ClientSecret: cmd.String("service-oauth2-client-secret"),
			TokenURL:     cmd.String("service-oauth2-token-endpoint"),
		}
		authorizer = api.NewOAuthAuthorizerFromClientCredentials(credentials)
		orchestratorClient = api.NewOAuthHTTPClient(orchestratorClient, authorizer)
		collectionHTTPClient = api.NewOAuthHTTPClient(collectionHTTPClient, authorizer)
	}

	// Assessment service configuration
	assessmentOpts = append([]service.Option[assessment.Service]{
		assessment.WithConfig(assessment.Config{
			OrchestratorAddress: cmd.String("assessment-orchestrator-address"),
			OrchestratorClient:  orchestratorClient,
			RegoPackage:         cmd.String("assessment-rego-package"),
		}),
	}, assessmentOptions...)

	assessmentSvc, err = assessment.NewService(assessmentOpts...)
	if err != nil {
		return err
	}

	// EvidenceStore service configuration
	evidenceOpts = append([]service.Option[evidence.Service]{
		evidence.WithConfig(evidence.Config{
			AssessmentAddress: cmd.String("evidence-assessment-address"),
			EvidenceQueueSize: evidence.DefaultConfig.EvidenceQueueSize,
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
			AssessmentHTTPClient: func() *http.Client {
				c := service.NewHTTPClient()
				c.Timeout = cmd.Duration("evidence-assessment-http-timeout")
				return c
			}(),
		}),
	}, evidenceOptions...)

	evidenceSvc, err = evidence.NewService(evidenceOpts...)
	if err != nil {
		return err
	}

	targetOfEvaluationID = cmd.String("target-of-evaluation-id")
	if targetOfEvaluationID == "" {
		targetOfEvaluationID = defaultEmbeddedTargetOfEvaluationID
	}

	evidenceStoreAddress = fmt.Sprintf("http://localhost:%d", apiPort)
	collectionSvc, err = collection.NewService(
		collection.WithConfig(collection.Config{
			Interval:                cmd.Duration("collection-interval"),
			Collectors:              collectors,
			EvidenceStoreAddress:    evidenceStoreAddress,
			EvidenceStoreHTTPClient: collectionHTTPClient,
			TargetOfEvaluationID:    targetOfEvaluationID,
			ToolID:                  collectionToolID,
		}),
	)
	if err != nil {
		return err
	}
	defer func() {
		closeErr := collectionSvc.Close()
		if err == nil && closeErr != nil {
			err = closeErr
		}
	}()

	// Server options configuration including CORS, logging, handler and gRPC reflection
	serverOpts = []server.Option{
		server.WithConfig(server.Config{
			Port:     apiPort,
			Path:     "/",
			LogLevel: cmd.String("log-level"),
			CORS: server.CORS{
				AllowedOrigins: cmd.StringSlice("api-cors-allowed-origins"),
				AllowedMethods: cmd.StringSlice("api-cors-allowed-methods"),
				AllowedHeaders: cmd.StringSlice("api-cors-allowed-headers"),
			},
		}),
		server.WithHandler(orchestratorconnect.NewOrchestratorHandler(
			orchestratorSvc,
			connect.WithInterceptors(interceptors...),
		)),
		server.WithHandler(assessmentconnect.NewAssessmentHandler(
			assessmentSvc,
			connect.WithInterceptors(interceptors...),
		)),
		server.WithHandler(evidenceconnect.NewEvidenceStoreHandler(
			evidenceSvc,
			connect.WithInterceptors(interceptors...),
		)),
		server.WithReflection(),
	}

	if cmd.Bool("oauth2-embedded") {
		serverOpts = append(serverOpts, server.WithEmbeddedOAuth2Server(
			cmd.String("oauth2-key-path"),
			cmd.String("oauth2-key-password"),
			cmd.Bool("oauth2-key-save-on-create"),
			cmd.String("oauth2-public-url"),
		))
	}

	srv, err = server.NewConnectServer(serverOpts)
	if err != nil {
		return err
	}

	serverErrCh = make(chan error, 1)
	go func() {
		serverErrCh <- srv.ListenAndServe()
	}()

	err = waitForLocalServer(ctx, apiPort)
	if err != nil {
		return err
	}

	collectionResults = collectionSvc.Start(ctx)
	go func() {
		for range collectionResults {
			slog.Debug("Collection cycle finished")
		}
	}()

	err = <-serverErrCh
	return err
}

func waitForLocalServer(ctx context.Context, port uint16) (err error) {
	var (
		addr   string
		conn   net.Conn
		ticker *time.Ticker
	)

	addr = fmt.Sprintf("127.0.0.1:%d", port)
	ticker = time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		conn, err = net.DialTimeout("tcp", addr, 100*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
