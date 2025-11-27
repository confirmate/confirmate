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
	"confirmate.io/core/persistence"

	"connectrpc.com/connect"
)

// service implements the Orchestrator service handler (see
// [orchestratorconnect.OrchestratorHandler]).
type service struct {
	orchestratorconnect.UnimplementedOrchestratorHandler
	db *persistence.DB
}

// NewService creates a new orchestrator service and returns a
// [orchestratorconnect.OrchestratorHandler].
//
// It initializes the database with auto-migration for the required types and sets up the necessary
// join tables.
func NewService() (orchestratorconnect.OrchestratorHandler, error) {
	var (
		svc = &service{}
		err error
	)

	// Initialize the database with the defined auto-migration types and join tables
	svc.db, err = persistence.NewDB(
		persistence.WithAutoMigration(types...),
		persistence.WithSetupJoinTable(joinTables...))
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

// ListTargetsOfEvaluation lists all targets of evaluation.
func (svc *service) ListTargetsOfEvaluation(
	ctx context.Context,
	req *connect.Request[orchestrator.ListTargetsOfEvaluationRequest],
) (res *connect.Response[orchestrator.ListTargetsOfEvaluationResponse], err error) {
	var (
		toes []*orchestrator.TargetOfEvaluation
	)

	err = svc.db.List(&toes, "name", true, 0, -1, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query targets of evaluation: %w", err)
	}

	res = connect.NewResponse(&orchestrator.ListTargetsOfEvaluationResponse{
		TargetsOfEvaluation: toes,
	})
	return
}

func (svc *service) ListAssessmentResults(
	ctx context.Context,
	req *connect.Request[orchestrator.ListAssessmentResultsRequest],
) (res *connect.Response[orchestrator.ListAssessmentResultsResponse], err error) {
	var (
		results []*assessment.AssessmentResult
	)

	err = svc.db.List(&results, "createdAt", true, 0, -1, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query targets of evaluation: %w", err)
	}

	res = connect.NewResponse(&orchestrator.ListAssessmentResultsResponse{
		Results: results,
	})
	return
}

func (svc *service) StoreAssessmentResults(
	ctx context.Context,
	req *connect.BidiStream[orchestrator.StoreAssessmentResultRequest, orchestrator.StoreAssessmentResultsResponse],
) error {
	for {
		msg, err := req.Receive()
		if err != nil {
			return err
		}

		// Store the assessment result in the database
		err = svc.db.Create(msg.Result)
		if err != nil {
			return fmt.Errorf("failed to store assessment result: %w", err)
		}

		// Send an acknowledgment response
		err = req.Send(&orchestrator.StoreAssessmentResultsResponse{Status: true})
		if err != nil {
			return err
		}
	}
}
