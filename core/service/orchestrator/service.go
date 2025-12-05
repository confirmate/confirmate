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
func NewService() (handler orchestratorconnect.OrchestratorHandler, err error) {
	var (
		svc = &service{}
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

	handler = svc
	return
}

// CreateTargetOfEvaluation creates a new target of evaluation.
func (svc *service) CreateTargetOfEvaluation(
	ctx context.Context,
	req *connect.Request[orchestrator.CreateTargetOfEvaluationRequest],
) (res *connect.Response[orchestrator.TargetOfEvaluation], err error) {
	var (
		toe = req.Msg.TargetOfEvaluation
		now = timestamppb.Now()
	)

	// Generate a new UUID for the target of evaluation
	toe.Id = uuid.NewString()

	// Set timestamps
	toe.CreatedAt = now
	toe.UpdatedAt = now

	// Persist the target of evaluation in the database
	err = svc.db.Create(toe)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("could not add target of evaluation to the database: %w", err))
	}

	res = connect.NewResponse(toe)
	return
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

// StoreAssessmentResults stores assessment results via a bidirectional stream.
func (svc *service) StoreAssessmentResults(
	ctx context.Context,
	stream *connect.BidiStream[orchestrator.StoreAssessmentResultRequest, orchestrator.StoreAssessmentResultsResponse],
) (err error) {
	var (
		msg *orchestrator.StoreAssessmentResultRequest
	)

	for {
		msg, err = stream.Receive()
		if err != nil {
			return err
		}

		// Store the assessment result in the database
		err = svc.db.Create(msg.Result)
		if err != nil {
			return fmt.Errorf("failed to store assessment result: %w", err)
		}

		// Send an acknowledgment response
		err = stream.Send(&orchestrator.StoreAssessmentResultsResponse{Status: true})
		if err != nil {
			return err
		}
	}
}

// GetTargetOfEvaluation retrieves a target of evaluation by ID.
func (svc *service) GetTargetOfEvaluation(
	ctx context.Context,
	req *connect.Request[orchestrator.GetTargetOfEvaluationRequest],
) (res *connect.Response[orchestrator.TargetOfEvaluation], err error) {
	var (
		toe orchestrator.TargetOfEvaluation
	)

	err = svc.db.Get(&toe, "id = ?", req.Msg.TargetOfEvaluationId)
	if errors.Is(err, persistence.ErrRecordNotFound) {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("target of evaluation not found"))
	} else if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	res = connect.NewResponse(&toe)
	return
}

// UpdateTargetOfEvaluation updates an existing target of evaluation.
func (svc *service) UpdateTargetOfEvaluation(
	ctx context.Context,
	req *connect.Request[orchestrator.UpdateTargetOfEvaluationRequest],
) (res *connect.Response[orchestrator.TargetOfEvaluation], err error) {
	var (
		count int64
		toe   = req.Msg.TargetOfEvaluation
	)

	// Check if the target of evaluation exists
	count, err = svc.db.Count(toe, "id = ?", toe.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	if count == 0 {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("target of evaluation not found"))
	}

	// Update timestamp
	toe.UpdatedAt = timestamppb.Now()

	// Save the updated target of evaluation
	err = svc.db.Save(toe, "id = ?", toe.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	res = connect.NewResponse(toe)
	return
}

// RemoveTargetOfEvaluation removes a target of evaluation by ID.
func (svc *service) RemoveTargetOfEvaluation(
	ctx context.Context,
	req *connect.Request[orchestrator.RemoveTargetOfEvaluationRequest],
) (res *connect.Response[emptypb.Empty], err error) {
	var (
		toe orchestrator.TargetOfEvaluation
	)

	// Delete the target of evaluation
	err = svc.db.Delete(&toe, "id = ?", req.Msg.TargetOfEvaluationId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	res = connect.NewResponse(&emptypb.Empty{})
	return
}

// GetTargetOfEvaluationStatistics retrieves statistics for targets of evaluation.
func (svc *service) GetTargetOfEvaluationStatistics(
	ctx context.Context,
	req *connect.Request[orchestrator.GetTargetOfEvaluationStatisticsRequest],
) (res *connect.Response[orchestrator.GetTargetOfEvaluationStatisticsResponse], err error) {
	// TODO: Implement actual statistics calculation
	// For now, return zero statistics
	res = connect.NewResponse(&orchestrator.GetTargetOfEvaluationStatisticsResponse{
		NumberOfDiscoveredResources: 0,
		NumberOfAssessmentResults:   0,
		NumberOfEvidences:           0,
		NumberOfSelectedCatalogs:    0,
	})
	return
}

// GetRuntimeInfo returns runtime information about the orchestrator service.
func (svc *service) GetRuntimeInfo(
	ctx context.Context,
	req *connect.Request[common.GetRuntimeInfoRequest],
) (res *connect.Response[common.Runtime], err error) {
	// TODO: Implement actual runtime information gathering
	// For now, return basic runtime info
	res = connect.NewResponse(&common.Runtime{
		Vcs:        "git",
		CommitHash: "unknown",
	})
	return
}
