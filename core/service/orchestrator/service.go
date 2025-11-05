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

	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/api/orchestrator/orchestratorconnect"
	"confirmate.io/core/persistence"

	"connectrpc.com/connect"
)

// service implements the Orchestrator service handler (see
// [orchestratorconnect.OrchestratorHandler]).
type service struct {
	orchestratorconnect.UnimplementedOrchestratorHandler
	db *persistence.DB
}

// NewService creates a new Orchestrator service.
//
// It initializes the database with auto-migration for the required types
// and sets up the necessary join tables.
func NewService() (orchestratorconnect.OrchestratorHandler, error) {
	var (
		svc = &service{}
		err error
	)

	// Initialize the database with the defined auto-migration types and join tables
	svc.db, err = persistence.NewDB(
		persistence.WithAutoMigration(types),
		persistence.WithSetupJoinTable(joinTable))
	if err != nil {
		return nil, fmt.Errorf("could not create db: %w", err)
	}

	// Create a sample TargetOfEvaluation entry. This will be removed later.
	err = svc.db.Create(&orchestrator.TargetOfEvaluation{
		Id:   "1",
		Name: "TOE1",
	})
	if err != nil {
		return nil, fmt.Errorf("could not create TOE: %w", err)
	}

	return svc, nil
}

// ListTargetsOfEvaluation lists all targets of evaluation objects in the database.
func (svc *service) ListTargetsOfEvaluation(context.Context, *connect.Request[orchestrator.ListTargetsOfEvaluationRequest]) (*connect.Response[orchestrator.ListTargetsOfEvaluationResponse], error) {
	var (
		toes = []*orchestrator.TargetOfEvaluation{}
		err  error
	)

	err = svc.db.List(&toes, "name", true, 0, -1, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query targets of evaluation: %w", err)
	}

	return connect.NewResponse(&orchestrator.ListTargetsOfEvaluationResponse{
		TargetsOfEvaluation: toes,
	}), nil
}
