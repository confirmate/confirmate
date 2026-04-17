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
	"net/http"
	"strings"

	cloud "confirmate.io/collectors/cloud"
	"confirmate.io/core/api"
	"confirmate.io/core/api/assessment/assessmentconnect"
	"confirmate.io/core/api/evidence/evidenceconnect"
	orchestratorapi "confirmate.io/core/api/orchestrator"
	"confirmate.io/core/api/orchestrator/orchestratorconnect"
	"confirmate.io/core/persistence"
	"confirmate.io/core/server"
	"confirmate.io/core/service"
	"confirmate.io/core/service/assessment"
	"confirmate.io/core/service/evidence"
	"confirmate.io/core/service/orchestrator"

	"connectrpc.com/connect"
	"github.com/urfave/cli/v3"
	"golang.org/x/oauth2/clientcredentials"
)

const (
	DefaultServiceTokenURL     = "http://localhost:8080/v1/auth/token"
	DefaultServiceClientID     = "confirmate"
	DefaultServiceClientSecret = "confirmate"
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

// discoveryFlags contains the flags for configuring optional collector auto-start.
var discoveryFlags = []cli.Flag{
	&cli.StringFlag{
		Name:    "discovery-auto-start",
		Usage:   "Auto-start discovery collector (none|azure)",
		Value:   "none",
		Sources: envVarSources("discovery-auto-start"),
	},
	&cli.StringFlag{
		Name:    "discovery-target-of-evaluation-id",
		Usage:   "Target of evaluation ID used by auto-started discovery collector",
		Value:   "",
		Sources: envVarSources("discovery-target-of-evaluation-id"),
	},
	&cli.StringFlag{
		Name:    "discovery-azure-subscription-id",
		Usage:   "Azure subscription ID used by auto-started azure discovery collector",
		Value:   "",
		Sources: envVarSources("discovery-azure-subscription-id"),
	},
	&cli.StringFlag{
		Name:    "discovery-azure-resource-group",
		Usage:   "Azure resource group used by auto-started azure discovery collector",
		Value:   "",
		Sources: envVarSources("discovery-azure-resource-group"),
	},
	&cli.DurationFlag{
		Name:    "discovery-interval",
		Usage:   "Collection interval for auto-started discovery collector",
		Value:   cloud.DefaultConfig().Interval,
		Sources: envVarSources("discovery-interval"),
	},
	&cli.DurationFlag{
		Name:    "discovery-cycle-timeout",
		Usage:   "Per-cycle timeout for auto-started discovery collector",
		Value:   cloud.DefaultConfig().CycleTimeout,
		Sources: envVarSources("discovery-cycle-timeout"),
	},
	&cli.DurationFlag{
		Name:    "discovery-http-timeout",
		Usage:   "HTTP timeout used by auto-started discovery collector",
		Value:   cloud.DefaultConfig().HTTPTimeout,
		Sources: envVarSources("discovery-http-timeout"),
	},
	&cli.StringFlag{
		Name:    "discovery-tool-id",
		Usage:   "Tool ID used by auto-started discovery collector",
		Value:   cloud.DefaultConfig().ToolID,
		Sources: envVarSources("discovery-tool-id"),
	},
	&cli.StringFlag{
		Name:    "discovery-evidence-store-address",
		Usage:   "Evidence store base address used by auto-started discovery collector",
		Value:   "",
		Sources: envVarSources("discovery-evidence-store-address"),
	},
}

// ConfirmateCommand starts the full framework: orchestrator,  assessment and evidence store services on one server.
var ConfirmateCommand = &cli.Command{
	Name:  "confirmate",
	Usage: "Launches the confirmate framework (including orchestrator, assessment and evidence store services)",
	Action: func(ctx context.Context, cmd *cli.Command) (err error) {
		var (
			interceptors                  []connect.Interceptor
			orchestratorOptions           []service.Option[orchestrator.Service]
			assessmentOptions             []service.Option[assessment.Service]
			evidenceOptions               []service.Option[evidence.Service]
			jwksURL                       string
			orchestratorOpts              []service.Option[orchestrator.Service]
			assessmentOpts                []service.Option[assessment.Service]
			evidenceOpts                  []service.Option[evidence.Service]
			orchestratorSvc               orchestratorconnect.OrchestratorHandler
			assessmentSvc                 assessmentconnect.AssessmentHandler
			evidenceSvc                   evidenceconnect.EvidenceStoreHandler
			orchestratorClient            *http.Client
			apiPort                       uint16
			credentials                   *clientcredentials.Config
			authorizer                    api.Authorizer
			serverOpts                    []server.Option
			discoveryMode                 string
			discoveryCfg                  cloud.Config
			toes                          *connect.Response[orchestratorapi.ListTargetsOfEvaluationResponse]
			assessmentOrchestratorAddress string
			evidenceAssessmentAddress     string
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

		assessmentOrchestratorAddress = cmd.String("assessment-orchestrator-address")
		if assessmentOrchestratorAddress == assessment.DefaultOrchestratorURL {
			assessmentOrchestratorAddress = fmt.Sprintf("http://localhost:%d", apiPort)
		}

		evidenceAssessmentAddress = cmd.String("evidence-assessment-address")
		if evidenceAssessmentAddress == evidence.DefaultAssessmentURL {
			evidenceAssessmentAddress = fmt.Sprintf("http://localhost:%d", apiPort)
		}

		orchestratorClient = http.DefaultClient
		if cmd.Bool("auth-enabled") {
			credentials = &clientcredentials.Config{
				ClientID:     cmd.String("service-oauth2-client-id"),
				ClientSecret: cmd.String("service-oauth2-client-secret"),
				TokenURL:     cmd.String("service-oauth2-token-endpoint"),
			}
			authorizer = api.NewOAuthAuthorizerFromClientCredentials(credentials)
			orchestratorClient = api.NewOAuthHTTPClient(orchestratorClient, authorizer)
		}

		// Assessment service configuration
		assessmentOpts = append([]service.Option[assessment.Service]{
			assessment.WithConfig(assessment.Config{
				OrchestratorAddress: assessmentOrchestratorAddress,
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
				AssessmentAddress: evidenceAssessmentAddress,
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
				AssessmentHTTPClient: &http.Client{
					Timeout: cmd.Duration("evidence-assessment-http-timeout"),
				},
			}),
		}, evidenceOptions...)

		evidenceSvc, err = evidence.NewService(evidenceOpts...)
		if err != nil {
			return err
		}

		discoveryMode = strings.ToLower(strings.TrimSpace(cmd.String("discovery-auto-start")))
		if discoveryMode == "azure" {
			discoveryCfg = cloud.DefaultConfig()
			discoveryCfg.EvidenceStoreAddress = strings.TrimSpace(cmd.String("discovery-evidence-store-address"))
			if discoveryCfg.EvidenceStoreAddress == "" {
				discoveryCfg.EvidenceStoreAddress = fmt.Sprintf("http://localhost:%d", apiPort)
			}

			discoveryCfg.TargetOfEvaluationID = strings.TrimSpace(cmd.String("discovery-target-of-evaluation-id"))
			if discoveryCfg.TargetOfEvaluationID == "" {
				toes, err = orchestratorSvc.ListTargetsOfEvaluation(ctx, connect.NewRequest(&orchestratorapi.ListTargetsOfEvaluationRequest{PageSize: 1}))
				if err != nil {
					return fmt.Errorf("resolving default target of evaluation for discovery: %w", err)
				}
				if len(toes.Msg.GetTargetsOfEvaluation()) == 0 {
					return fmt.Errorf("no target of evaluation available for discovery collector; set --discovery-target-of-evaluation-id")
				}
				discoveryCfg.TargetOfEvaluationID = toes.Msg.GetTargetsOfEvaluation()[0].GetId()
			}

			discoveryCfg.AzureSubscriptionID = strings.TrimSpace(cmd.String("discovery-azure-subscription-id"))
			discoveryCfg.AzureResourceGroup = strings.TrimSpace(cmd.String("discovery-azure-resource-group"))
			discoveryCfg.Interval = cmd.Duration("discovery-interval")
			discoveryCfg.CycleTimeout = cmd.Duration("discovery-cycle-timeout")
			discoveryCfg.HTTPTimeout = cmd.Duration("discovery-http-timeout")
			discoveryCfg.ToolID = strings.TrimSpace(cmd.String("discovery-tool-id"))

			if err = discoveryCfg.Validate(); err != nil {
				return fmt.Errorf("invalid discovery collector configuration: %w", err)
			}

			go func() {
				err := cloud.Start(ctx, discoveryCfg, slog.Default())
				if err != nil && !strings.Contains(err.Error(), "context canceled") {
					slog.Error("discovery collector stopped with error", slog.String("error", err.Error()))
				}
			}()
			slog.Info("started discovery collector", slog.String("mode", discoveryMode), slog.String("target_of_evaluation_id", discoveryCfg.TargetOfEvaluationID), slog.String("resource_group", discoveryCfg.AzureResourceGroup))
		} else if discoveryMode != "" && discoveryMode != "none" {
			return fmt.Errorf("unsupported discovery-auto-start mode %q", discoveryMode)
		}

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

		err = server.RunConnectServer(serverOpts...)
		return err
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
		discoveryFlags,
	),
}
