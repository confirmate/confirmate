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

	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/service"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// StoreAssessmentResult stores a single assessment result.
func (svc *Service) StoreAssessmentResult(
	ctx context.Context,
	req *connect.Request[orchestrator.StoreAssessmentResultRequest],
) (res *connect.Response[orchestrator.StoreAssessmentResultResponse], err error) {
	var (
		result *assessment.AssessmentResult
	)

	// Validate the request
	if err = service.Validate(req.Msg); err != nil {
		return nil, err
	}

	result = req.Msg.Result

	// Set timestamp
	result.CreatedAt = timestamppb.Now()

	// Persist the assessment result in the database
	err = svc.db.Create(result)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	res = connect.NewResponse(&orchestrator.StoreAssessmentResultResponse{})
	return
}

// GetAssessmentResult retrieves an assessment result by ID.
func (svc *Service) GetAssessmentResult(
	ctx context.Context,
	req *connect.Request[orchestrator.GetAssessmentResultRequest],
) (res *connect.Response[assessment.AssessmentResult], err error) {
	var (
		result assessment.AssessmentResult
	)

	// Validate the request
	if err = service.Validate(req.Msg); err != nil {
		return nil, err
	}

	err = svc.db.Get(&result, "id = ?", req.Msg.Id)
	if err = service.HandleDatabaseError(err, service.ErrNotFound("assessment result")); err != nil {
		return nil, err
	}

	res = connect.NewResponse(&result)
	return
}

// ListAssessmentResults lists all assessment results with optional filtering.
func (svc *Service) ListAssessmentResults(
	ctx context.Context,
	req *connect.Request[orchestrator.ListAssessmentResultsRequest],
) (res *connect.Response[orchestrator.ListAssessmentResultsResponse], err error) {
	var (
		results []*assessment.AssessmentResult
		conds   []any
	)

	// Validate the request
	if err = service.Validate(req.Msg); err != nil {
		return nil, err
	}

	// Apply filters if provided
	if req.Msg.Filter != nil {
		if req.Msg.Filter.TargetOfEvaluationId != nil {
			conds = append(conds, "target_of_evaluation_id = ?", *req.Msg.Filter.TargetOfEvaluationId)
		}
		if req.Msg.Filter.Compliant != nil {
			conds = append(conds, "compliant = ?", *req.Msg.Filter.Compliant)
		}
		if req.Msg.Filter.MetricId != nil {
			conds = append(conds, "metric_id = ?", *req.Msg.Filter.MetricId)
		}
	}

	err = svc.db.List(&results, "timestamp DESC", false, 0, -1, conds...)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	res = connect.NewResponse(&orchestrator.ListAssessmentResultsResponse{
		Results: results,
	})
	return
}

// StoreAssessmentResults stores assessment results via a bidirectional stream.
func (svc *Service) StoreAssessmentResults(
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
		if err = service.HandleDatabaseError(err); err != nil {
			return err
		}

		// Send an acknowledgment response
		err = stream.Send(&orchestrator.StoreAssessmentResultsResponse{Status: true})
		if err != nil {
			return err
		}
	}
}
