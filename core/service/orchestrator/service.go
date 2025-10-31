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

	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/api/orchestrator/orchestratorconnect"
	"confirmate.io/core/db"
	"connectrpc.com/connect"
)

type service struct {
	orchestratorconnect.UnimplementedOrchestratorHandler
	storage *db.Storage
}

func NewService() (orchestratorconnect.OrchestratorHandler, error) {
	var (
		svc = &service{}
		err error
	)

	svc.storage, err = db.NewStorage(
		db.WithAutoMigration(types),
		db.WithSetupJoinTable(jointTable))
	if err != nil {
		return nil, fmt.Errorf("could not create storage: %w", err)
	}

	// Setup Join Table
	if err = svc.storage.DB.SetupJoinTable(orchestrator.TargetOfEvaluation{}, "ConfiguredMetrics", assessment.MetricConfiguration{}); err != nil {
		return nil, fmt.Errorf("error during join-table: %w", err)
	}
	// Create table
	err = svc.storage.DB.AutoMigrate(
		orchestrator.TargetOfEvaluation{})
	if err != nil {
		return nil, fmt.Errorf("could not migrate TargetOfEvaluation: %w", err)
	}

	err = svc.storage.Create(&orchestrator.TargetOfEvaluation{
		Id:   "1",
		Name: "TOE1",
	})
	if err != nil {
		return nil, fmt.Errorf("could not create TOE: %w", err)
	}

	return svc, nil
}

func (svc *service) ListTargetsOfEvaluation(context.Context, *connect.Request[orchestrator.ListTargetsOfEvaluationRequest]) (*connect.Response[orchestrator.ListTargetsOfEvaluationResponse], error) {
	var (
		toes = []*orchestrator.TargetOfEvaluation{}
		err  error
	)

	err = svc.storage.List(&toes, "name", true, 0, -1, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query targets of evaluation: %w", err)
	}

	return connect.NewResponse(&orchestrator.ListTargetsOfEvaluationResponse{
		TargetsOfEvaluation: toes,
	}), nil
}
