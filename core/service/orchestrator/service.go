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

package orchestrator

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"
	"sync"
	"time"

	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/common"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/api/orchestrator/orchestratorconnect"
	"confirmate.io/core/persistence"
	"confirmate.io/core/service"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Service implements the Orchestrator Service handler (see
// [orchestratorconnect.OrchestratorHandler]).
type Service struct {
	orchestratorconnect.UnimplementedOrchestratorHandler
	db  persistence.DB
	cfg Config

	// subscribers is a map of subscribers for change events
	subscribers      map[int64]*subscriber
	subscribersMutex sync.RWMutex

	nextSubscriberId int64
}

type subscriber struct {
	ch     chan *orchestrator.ChangeEvent
	filter *orchestrator.SubscribeRequest_Filter
}

// DefaultConfig is the default configuration for the orchestrator [Service].
var DefaultConfig = Config{
	DefaultCatalogsPath:             "catalogs",
	DefaultMetricsPath:              "./policies/security-metrics/metrics",
	CreateDefaultTargetOfEvaluation: true,
	LoadDefaultCatalogs:             true,
	LoadDefaultMetrics:              true,
}

// Config represents the configuration for the orchestrator [Service].
type Config struct {
	// DefaultCatalogsPath is the path where default catalogs are stored.
	DefaultCatalogsPath string
	// LoadDefaultCatalogs controls whether to load default catalogs from
	// [Config.DefaultCatalogsPath].
	LoadDefaultCatalogs bool
	// LoadCatalogsFunc is an optional function to load additional catalogs at service start. This
	// function is called in addition to loading from [Config.DefaultCatalogsPath] (if enabled).
	LoadCatalogsFunc func(*Service) ([]*orchestrator.Catalog, error)

	// DefaultMetricsPath is the path containing default metrics (e.g., security-metrics
	// repository).
	DefaultMetricsPath string
	// LoadDefaultMetrics controls whether to load default metrics from [Config.DefaultMetricsPath].
	LoadDefaultMetrics bool
	// LoadMetricsFunc is an optional function to load additional metrics at service start. This
	// function is called in addition to loading from [Config.DefaultMetricsPath] (if enabled).
	LoadMetricsFunc func(*Service) ([]*assessment.Metric, error)

	// CreateDefaultTargetOfEvaluation controls whether to create a default target of evaluation.
	CreateDefaultTargetOfEvaluation bool
}

// WithConfig sets the service configuration, overriding the default configuration.
func WithConfig(cfg Config) service.Option[Service] {
	return func(svc *Service) {
		svc.cfg = cfg
	}
}

// NewService creates a new orchestrator service and returns a
// [orchestratorconnect.OrchestratorHandler].
//
// It initializes the database with auto-migration for the required types and sets up the necessary
// join tables.
func NewService(opts ...service.Option[Service]) (handler orchestratorconnect.OrchestratorHandler, err error) {
	var (
		svc = &Service{
			cfg: DefaultConfig,
		}
	)

	for _, o := range opts {
		o(svc)
	}

	// Initialize the database with the defined auto-migration types and join tables
	svc.db, err = persistence.NewDB(
		persistence.WithAutoMigration(types...),
		persistence.WithSetupJoinTable(joinTables...))
	if err != nil {
		return nil, fmt.Errorf("could not create db: %w", err)
	}

	// Initialize subscribers map
	svc.subscribers = make(map[int64]*subscriber)

	// Load catalogs and metrics (log errors but continue - they're not critical for service startup)
	if err = svc.loadCatalogs(); err != nil {
		slog.Warn("could not load catalogs, continuing with empty catalog list", "error", err)
	}

	if err = svc.loadMetrics(); err != nil {
		slog.Warn("could not load metrics, continuing with empty metric list", "error", err)
	}

	// Create default target of evaluation if enabled and none exists
	if svc.cfg.CreateDefaultTargetOfEvaluation {
		if _, err = svc.CreateDefaultTargetOfEvaluation(); err != nil {
			return nil, fmt.Errorf("could not create default target of evaluation: %w", err)
		}
	}

	handler = svc
	return
}

// GetRuntimeInfo returns runtime information about the orchestrator service.
func (svc *Service) GetRuntimeInfo(
	ctx context.Context,
	req *connect.Request[common.GetRuntimeInfoRequest],
) (res *connect.Response[common.Runtime], err error) {
	var (
		runtime = &common.Runtime{}
		info    *debug.BuildInfo
		ok      bool
	)

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	if info, ok = debug.ReadBuildInfo(); ok {
		runtime.GolangVersion = info.GoVersion
		runtime.Vcs = "git"

		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				runtime.CommitHash = setting.Value
			case "vcs.time":
				if t, err := time.Parse(time.RFC3339, setting.Value); err == nil {
					runtime.CommitTime = timestamppb.New(t)
				}
			}
		}

		for _, dep := range info.Deps {
			runtime.Dependencies = append(runtime.Dependencies, &common.Dependency{
				Path:    dep.Path,
				Version: dep.Version,
			})
		}
	}

	res = connect.NewResponse(runtime)
	return
}
