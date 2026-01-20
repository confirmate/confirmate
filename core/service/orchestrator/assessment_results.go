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
	"io"
	"strings"

	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/service"
	"confirmate.io/core/util"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// StoreAssessmentResult stores a single assessment result.
func (svc *Service) StoreAssessmentResult(
	_ context.Context,
	req *connect.Request[orchestrator.StoreAssessmentResultRequest],
) (res *connect.Response[orchestrator.StoreAssessmentResultResponse], err error) {
	var (
		result *assessment.AssessmentResult
	)

	// Validate the request
	if err = service.Validate(req); err != nil {
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

	// Notify subscribers
	go svc.publishEvent(&orchestrator.ChangeEvent{
		Timestamp:   timestamppb.Now(),
		Category:    orchestrator.EventCategory_EVENT_CATEGORY_ASSESSMENT_RESULT,
		RequestType: orchestrator.RequestType_REQUEST_TYPE_CREATED,
		EntityId:    result.Id,
		Entity: &orchestrator.ChangeEvent_AssessmentResult{
			AssessmentResult: result,
		},
	})

	res = connect.NewResponse(&orchestrator.StoreAssessmentResultResponse{})
	return
}

// GetAssessmentResult retrieves an assessment result by ID.
func (svc *Service) GetAssessmentResult(
	_ context.Context,
	req *connect.Request[orchestrator.GetAssessmentResultRequest],
) (res *connect.Response[assessment.AssessmentResult], err error) {
	var (
		result assessment.AssessmentResult
	)

	// Validate the request
	if err = service.Validate(req); err != nil {
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
	_ context.Context,
	req *connect.Request[orchestrator.ListAssessmentResultsRequest],
) (res *connect.Response[orchestrator.ListAssessmentResultsResponse], err error) {
	var (
		results []*assessment.AssessmentResult
		conds   []any
		npt     string
	)

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	// Set default ordering
	if req.Msg.OrderBy == "" {
		req.Msg.OrderBy = "timestamp"
		req.Msg.Asc = false
	}

	// Apply filters if provided
	if req.Msg.Filter != nil {
		var whereClauses []string
		var args []any

		if req.Msg.Filter.TargetOfEvaluationId != nil {
			whereClauses = append(whereClauses, "target_of_evaluation_id = ?")
			args = append(args, util.Deref(req.Msg.Filter.TargetOfEvaluationId))
		}
		if req.Msg.Filter.Compliant != nil {
			whereClauses = append(whereClauses, "compliant = ?")
			args = append(args, util.Deref(req.Msg.Filter.Compliant))
		}
		if req.Msg.Filter.MetricId != nil {
			whereClauses = append(whereClauses, "metric_id = ?")
			args = append(args, util.Deref(req.Msg.Filter.MetricId))
		}
		if req.Msg.Filter.ToolId != nil {
			whereClauses = append(whereClauses, "tool_id = ?")
			args = append(args, util.Deref(req.Msg.Filter.ToolId))
		}
		if len(req.Msg.Filter.AssessmentResultIds) > 0 {
			// Build IN clause dynamically to support ramsql (doesn't support array binding)
			placeholders := strings.Repeat("?,", len(req.Msg.Filter.AssessmentResultIds))
			placeholders = placeholders[:len(placeholders)-1] // Remove trailing comma
			whereClauses = append(whereClauses, "id IN ("+placeholders+")")
			for _, id := range req.Msg.Filter.AssessmentResultIds {
				args = append(args, id)
			}
		}

		// Combine all WHERE clauses with AND
		if len(whereClauses) > 0 {
			whereQuery := strings.Join(whereClauses, " AND ")
			conds = append(conds, whereQuery)
			conds = append(conds, args...)
		}
	}

	// Handle latest_by_resource_id filter
	// This returns only the most recent assessment result for each unique (resource_id, metric_id) pair
	// Uses PostgreSQL's DISTINCT ON for efficient grouping
	if req.Msg.LatestByResourceId != nil && util.Deref(req.Msg.LatestByResourceId) {
		// Build WHERE clause from existing conditions
		var where string
		var args []any

		if len(conds) > 0 {
			// conds is structured as [query1, args1, query2, args2, ...]
			var whereParts []string
			for i := 0; i < len(conds); i += 2 {
				whereParts = append(whereParts, conds[i].(string))
				if i+1 < len(conds) {
					args = append(args, conds[i+1])
				}
			}
			where = "WHERE " + strings.Join(whereParts, " AND ")
		}

		// Use PostgreSQL DISTINCT ON with ORDER BY to get latest result per (resource_id, metric_id)
		query := fmt.Sprintf(`
			SELECT DISTINCT ON (resource_id, metric_id) *
			FROM assessment_results
			%s
			ORDER BY resource_id, metric_id, created_at DESC
		`, where)

		err = svc.db.Raw(&results, query, args...)
		if err = service.HandleDatabaseError(err); err != nil {
			return nil, err
		}

		// Since we used raw SQL, we need to handle pagination differently
		// For now, return all results without pagination support
		res = connect.NewResponse(&orchestrator.ListAssessmentResultsResponse{
			Results:       results,
			NextPageToken: "",
		})
		return
	}

	results, npt, err = service.PaginateStorage[*assessment.AssessmentResult](req.Msg, svc.db, service.DefaultPaginationOpts, conds...)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	res = connect.NewResponse(&orchestrator.ListAssessmentResultsResponse{
		Results:       results,
		NextPageToken: npt,
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
		res *orchestrator.StoreAssessmentResultsResponse
	)

	for {
		msg, err = stream.Receive()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}

		// Call StoreAssessmentResult() for storing a single assessment
		// This ensures validation, timestamp setting, persistence, and event publishing
		// all go through the same code path
		_, err = svc.StoreAssessmentResult(ctx, connect.NewRequest(msg))
		if err != nil {
			// Create error response
			res = &orchestrator.StoreAssessmentResultsResponse{
				Status:        false,
				StatusMessage: err.Error(),
			}
		} else {
			// Create success response
			res = &orchestrator.StoreAssessmentResultsResponse{
				Status: true,
			}
		}

		// Send response
		err = stream.Send(res)
		if err != nil {
			return err
		}
	}
}
