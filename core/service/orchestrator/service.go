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
	"runtime/debug"
	"time"

	"sync"

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
	db *persistence.DB

	catalogsFolder string
	metricsFolder  string

	// subscribers is a map of subscribers for change events
	subscribers      map[int64]*subscriber
	subscribersMutex sync.RWMutex

	nextSubscriberId int64
}

type subscriber struct {
	ch     chan *orchestrator.ChangeEvent
	filter *orchestrator.SubscribeRequest_Filter
}

// WithCatalogsFolder sets the folder where catalogs are stored.
func WithCatalogsFolder(folder string) service.Option[Service] {
	return func(svc *Service) {
		svc.catalogsFolder = folder
	}
}

// WithMetricsFolder sets the folder where metrics are stored.
func WithMetricsFolder(folder string) service.Option[Service] {
	return func(svc *Service) {
		svc.metricsFolder = folder
	}
}

// NewService creates a new orchestrator service and returns a
// [orchestratorconnect.OrchestratorHandler].
//
// It initializes the database with auto-migration for the required types and sets up the necessary
// join tables.
func NewService(opts ...service.Option[Service]) (handler orchestratorconnect.OrchestratorHandler, err error) {
	var (
		svc = &Service{}
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

	// Load catalogs and metrics
	if err = svc.loadCatalogs(); err != nil {
		return nil, fmt.Errorf("could not load catalogs: %w", err)
	}

	if err = svc.loadMetrics(); err != nil {
		return nil, fmt.Errorf("could not load metrics: %w", err)
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
	if err = service.Validate(req.Msg); err != nil {
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

