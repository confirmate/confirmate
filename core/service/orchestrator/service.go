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
	"errors"
	"fmt"

	"confirmate.io/core/api/common"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/api/orchestrator/orchestratorconnect"
	"confirmate.io/core/persistence"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
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

// CreateTargetOfEvaluation creates a new target of evaluation.
func (svc *service) CreateTargetOfEvaluation(
	ctx context.Context,
	req *connect.Request[orchestrator.CreateTargetOfEvaluationRequest],
) (*connect.Response[orchestrator.TargetOfEvaluation], error) {
	var res *orchestrator.TargetOfEvaluation

	// Generate a new UUID for the target of evaluation
	req.Msg.TargetOfEvaluation.Id = uuid.NewString()

	res = req.Msg.TargetOfEvaluation

	// Set timestamps
	now := timestamppb.Now()
	res.CreatedAt = now
	res.UpdatedAt = now

	// Persist the target of evaluation in the database
	err := svc.db.Create(res)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("could not add target of evaluation to the database: %w", err))
	}

	return connect.NewResponse(res), nil
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

// GetTargetOfEvaluation retrieves a target of evaluation by ID.
func (svc *service) GetTargetOfEvaluation(
	ctx context.Context,
	req *connect.Request[orchestrator.GetTargetOfEvaluationRequest],
) (*connect.Response[orchestrator.TargetOfEvaluation], error) {
	var res orchestrator.TargetOfEvaluation

	err := svc.db.Get(&res, "id = ?", req.Msg.TargetOfEvaluationId)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("target of evaluation not found"))
	} else if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	return connect.NewResponse(&res), nil
}

// UpdateTargetOfEvaluation updates an existing target of evaluation.
func (svc *service) UpdateTargetOfEvaluation(
	ctx context.Context,
	req *connect.Request[orchestrator.UpdateTargetOfEvaluationRequest],
) (*connect.Response[orchestrator.TargetOfEvaluation], error) {
	// Check if the target of evaluation exists
	count, err := svc.db.Count(req.Msg.TargetOfEvaluation, "id = ?", req.Msg.TargetOfEvaluation.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	if count == 0 {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("target of evaluation not found"))
	}

	// Update timestamp
	res := req.Msg.TargetOfEvaluation
	res.UpdatedAt = timestamppb.Now()

	// Save the updated target of evaluation
	err = svc.db.Save(res, "id = ?", req.Msg.TargetOfEvaluation.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	return connect.NewResponse(res), nil
}

// RemoveTargetOfEvaluation removes a target of evaluation by ID.
func (svc *service) RemoveTargetOfEvaluation(
	ctx context.Context,
	req *connect.Request[orchestrator.RemoveTargetOfEvaluationRequest],
) (*connect.Response[emptypb.Empty], error) {
	var toe orchestrator.TargetOfEvaluation

	// Delete the target of evaluation
	err := svc.db.Delete(&toe, "id = ?", req.Msg.TargetOfEvaluationId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

// GetTargetOfEvaluationStatistics retrieves statistics for targets of evaluation.
func (svc *service) GetTargetOfEvaluationStatistics(
	ctx context.Context,
	req *connect.Request[orchestrator.GetTargetOfEvaluationStatisticsRequest],
) (*connect.Response[orchestrator.GetTargetOfEvaluationStatisticsResponse], error) {
	// TODO: Implement actual statistics calculation
	// For now, return zero statistics
	return connect.NewResponse(&orchestrator.GetTargetOfEvaluationStatisticsResponse{
		NumberOfDiscoveredResources: 0,
		NumberOfAssessmentResults:   0,
		NumberOfEvidences:           0,
		NumberOfSelectedCatalogs:    0,
	}), nil
}

// GetRuntimeInfo returns runtime information about the orchestrator service.
func (svc *service) GetRuntimeInfo(
	ctx context.Context,
	req *connect.Request[common.GetRuntimeInfoRequest],
) (*connect.Response[common.Runtime], error) {
	// TODO: Implement actual runtime information gathering
	// For now, return basic runtime info
	return connect.NewResponse(&common.Runtime{
		Vcs:        "git",
		CommitHash: "unknown",
	}), nil
}
