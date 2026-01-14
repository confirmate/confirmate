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
	"confirmate.io/core/service"

	"buf.build/go/protovalidate"
	"connectrpc.com/connect"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CreateTargetOfEvaluation creates a new target of evaluation.
func (svc *Service) CreateTargetOfEvaluation(
	ctx context.Context,
	req *connect.Request[orchestrator.CreateTargetOfEvaluationRequest],
) (res *connect.Response[orchestrator.TargetOfEvaluation], err error) {
	var (
		toe *orchestrator.TargetOfEvaluation
		now = timestamppb.Now()
	)

	// Validate the request, ignoring ID field which may be auto-generated
	if err = service.Validate(req, protovalidate.WithFilter(service.IgnoreIDFilter)); err != nil {
		return nil, err
	}

	toe = req.Msg.TargetOfEvaluation

	// Generate a new UUID for the target of evaluation
	toe.Id = uuid.NewString()

	// Set timestamps
	toe.CreatedAt = now
	toe.UpdatedAt = now

	// Persist the target of evaluation in the database
	err = svc.db.Create(toe)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	// Notify subscribers
	go svc.publishEvent(&orchestrator.ChangeEvent{
		Timestamp:  timestamppb.Now(),
		Category:   orchestrator.EventCategory_EVENT_CATEGORY_TARGET_OF_EVALUATION,
		RequestType: orchestrator.RequestType_REQUEST_TYPE_CREATED,
		EntityId:   toe.Id,
		Entity: &orchestrator.ChangeEvent_TargetOfEvaluation{
			TargetOfEvaluation: toe,
		},
	})

	res = connect.NewResponse(toe)
	return
}

// GetTargetOfEvaluation retrieves a target of evaluation by ID.
func (svc *Service) GetTargetOfEvaluation(
	ctx context.Context,
	req *connect.Request[orchestrator.GetTargetOfEvaluationRequest],
) (res *connect.Response[orchestrator.TargetOfEvaluation], err error) {
	var (
		toe orchestrator.TargetOfEvaluation
	)

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	err = svc.db.Get(&toe, "id = ?", req.Msg.TargetOfEvaluationId)
	if err = service.HandleDatabaseError(err, service.ErrNotFound("target of evaluation")); err != nil {
		return nil, err
	}

	res = connect.NewResponse(&toe)
	return
}

// ListTargetsOfEvaluation lists all targets of evaluation.
func (svc *Service) ListTargetsOfEvaluation(
	ctx context.Context,
	req *connect.Request[orchestrator.ListTargetsOfEvaluationRequest],
) (res *connect.Response[orchestrator.ListTargetsOfEvaluationResponse], err error) {
	var (
		toes []*orchestrator.TargetOfEvaluation
		npt  string
	)

	// Validate request
	err = service.Validate(req)
	if err != nil {
		return nil, err
	}

	// Set default ordering
	if req.Msg.OrderBy == "" {
		req.Msg.OrderBy = "name"
		req.Msg.Asc = true
	}

	toes, npt, err = service.PaginateStorage[*orchestrator.TargetOfEvaluation](req.Msg, svc.db, service.DefaultPaginationOpts)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	res = connect.NewResponse(&orchestrator.ListTargetsOfEvaluationResponse{
		TargetsOfEvaluation: toes,
		NextPageToken:       npt,
	})
	return
}

// UpdateTargetOfEvaluation updates an existing target of evaluation.
func (svc *Service) UpdateTargetOfEvaluation(
	ctx context.Context,
	req *connect.Request[orchestrator.UpdateTargetOfEvaluationRequest],
) (res *connect.Response[orchestrator.TargetOfEvaluation], err error) {
	var toe *orchestrator.TargetOfEvaluation

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	toe = req.Msg.TargetOfEvaluation

	// Update timestamp
	toe.UpdatedAt = timestamppb.Now()

	// Update the target of evaluation
	err = svc.db.Update(toe, "id = ?", toe.Id)
	if err = service.HandleDatabaseError(err, service.ErrNotFound("target of evaluation")); err != nil {
		return nil, err
	}

	// Notify subscribers
	go svc.publishEvent(&orchestrator.ChangeEvent{
		Timestamp:  timestamppb.Now(),
		Category:   orchestrator.EventCategory_EVENT_CATEGORY_TARGET_OF_EVALUATION,
		RequestType: orchestrator.RequestType_REQUEST_TYPE_UPDATED,
		EntityId:   toe.Id,
		Entity: &orchestrator.ChangeEvent_TargetOfEvaluation{
			TargetOfEvaluation: toe,
		},
	})

	res = connect.NewResponse(toe)
	return
}

// RemoveTargetOfEvaluation removes a target of evaluation by ID.
func (svc *Service) RemoveTargetOfEvaluation(
	ctx context.Context,
	req *connect.Request[orchestrator.RemoveTargetOfEvaluationRequest],
) (res *connect.Response[emptypb.Empty], err error) {
	var (
		toe orchestrator.TargetOfEvaluation
	)

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	// Delete the target of evaluation
	err = svc.db.Delete(&toe, "id = ?", req.Msg.TargetOfEvaluationId)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	// Notify subscribers
	go svc.publishEvent(&orchestrator.ChangeEvent{
		Timestamp:  timestamppb.Now(),
		Category:   orchestrator.EventCategory_EVENT_CATEGORY_TARGET_OF_EVALUATION,
		RequestType: orchestrator.RequestType_REQUEST_TYPE_DELETED,
		EntityId:   req.Msg.TargetOfEvaluationId,
	})

	res = connect.NewResponse(&emptypb.Empty{})
	return
}

// GetTargetOfEvaluationStatistics retrieves statistics for targets of evaluation.
func (svc *Service) GetTargetOfEvaluationStatistics(
	ctx context.Context,
	req *connect.Request[orchestrator.GetTargetOfEvaluationStatisticsRequest],
) (res *connect.Response[orchestrator.GetTargetOfEvaluationStatisticsResponse], err error) {
	var (
		count int64
	)

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	res = connect.NewResponse(&orchestrator.GetTargetOfEvaluationStatisticsResponse{})

	// Get number of selected catalogs (Audit Scopes)
	count, err = svc.db.Count(&orchestrator.AuditScope{}, "target_of_evaluation_id = ?", req.Msg.TargetOfEvaluationId)
	if err != nil {
		return nil, service.HandleDatabaseError(err)
	}
	res.Msg.NumberOfSelectedCatalogs = count

	// Get number of assessment results
	count, err = svc.db.Count(&assessment.AssessmentResult{}, "target_of_evaluation_id = ?", req.Msg.TargetOfEvaluationId)
	if err != nil {
		return nil, service.HandleDatabaseError(err)
	}
	res.Msg.NumberOfAssessmentResults = count

	// TODO: Get number of discovered resources
	res.Msg.NumberOfDiscoveredResources = 0

	// TODO: Get number of evidences
	res.Msg.NumberOfEvidences = 0

	return
}

// CreateDefaultTargetOfEvaluation creates a new "default" target of evaluation,
// if no target of evaluation exists in the database.
//
// If a new target of evaluation was created, it will be returned.
func (svc *Service) CreateDefaultTargetOfEvaluation() (target *orchestrator.TargetOfEvaluation, err error) {
	var (
		count int64
	)

	count, err = svc.db.Count(&orchestrator.TargetOfEvaluation{})
	if err != nil {
		return nil, fmt.Errorf("storage error: %w", err)
	}

	if count == 0 {
		now := timestamppb.Now()

		// Create a default target of evaluation
		target = &orchestrator.TargetOfEvaluation{
			Id:          "00000000-0000-0000-0000-000000000000",
			Name:        "Default Target of Evaluation",
			Description: "This is the default target of evaluation",
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		// Save it in the database
		err = svc.db.Create(target)
		if err != nil {
			return nil, fmt.Errorf("storage error: %w", err)
		}
	}

	return
}
