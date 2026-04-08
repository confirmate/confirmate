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
	ctx context.Context,
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

	// Check access via the configured auth strategy
	allowed, _, err := CheckAccess(ctx, svc.authz, svc, orchestrator.RequestType_REQUEST_TYPE_CREATED, result.TargetOfEvaluationId, orchestrator.ObjectType_OBJECT_TYPE_ASSESSMENT_RESULT)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("%w: %w", service.ErrDatabaseError, err))
	}
	if !allowed {
		return nil, service.ErrPermissionDenied
	}

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
	ctx context.Context,
	req *connect.Request[orchestrator.GetAssessmentResultRequest],
) (res *connect.Response[assessment.AssessmentResult], err error) {
	var (
		result  assessment.AssessmentResult
		allowed bool
	)

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	err = svc.db.Get(&result, "id = ?", req.Msg.Id)
	if err = service.HandleDatabaseError(err, service.ErrNotFound("assessment result")); err != nil {
		return nil, err
	}

	// Check access via the configured auth strategy
	allowed, _, err = CheckAccess(ctx, svc.authz, svc, orchestrator.RequestType_REQUEST_TYPE_GET, req.Msg.GetId(), orchestrator.ObjectType_OBJECT_TYPE_ASSESSMENT_RESULT)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if !allowed {
		return nil, connect.NewError(connect.CodePermissionDenied, service.ErrPermissionDenied)
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
		results      []*assessment.AssessmentResult
		conds        []any
		npt          string
		where        string
		args         []any
		allowed      bool
		resourceList []string
	)

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	// Set default ordering
	if req.Msg.OrderBy == "" {
		req.Msg.OrderBy = "created_at"
		req.Msg.Asc = false
	}

	var whereClauses []string

	// Apply filters if provided
	if req.Msg.Filter != nil {
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
			var placeholders string
			placeholders = strings.Repeat("?,", len(req.Msg.Filter.AssessmentResultIds))
			placeholders = placeholders[:len(placeholders)-1] // Remove trailing comma
			whereClauses = append(whereClauses, "id IN ("+placeholders+")")
			for _, id := range req.Msg.Filter.AssessmentResultIds {
				args = append(args, id)
			}
		}
	}

	// Combine all WHERE clauses with AND
	if len(whereClauses) > 0 {
		where = strings.Join(whereClauses, " AND ")
		conds = append(conds, where)
		conds = append(conds, args...)
	}

	// Handle latest_by_resource_id filter
	// This returns only the most recent assessment result for each unique (resource_id, metric_id) pair
	// Uses PostgreSQL's DISTINCT ON for efficient grouping
	if req.Msg.LatestByResourceId != nil && util.Deref(req.Msg.LatestByResourceId) {
		// Reuse the WHERE query and args directly.
		if where != "" {
			where = "WHERE " + where
		}

		// Use PostgreSQL DISTINCT ON with ORDER BY to get latest result per (resource_id, metric_id)
		var query string
		query = fmt.Sprintf(`
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

	// Check access via the configured auth strategy
	allowed, resourceList, err = CheckAccess(ctx, svc.authz, svc, orchestrator.RequestType_REQUEST_TYPE_LIST, "", orchestrator.ObjectType_OBJECT_TYPE_ASSESSMENT_RESULT)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// If access is not allowed to all resources and the resource list is empty, return an empty response
	if len(resourceList) == 0 && !allowed {
		return connect.NewResponse(&orchestrator.ListAssessmentResultsResponse{
			Results:       []*assessment.AssessmentResult{},
			NextPageToken: "",
		}), nil
	}

	// If access is not allowed to all resources, add a condition to filter by the allowed resource IDs
	if !allowed {
		conds = append(conds, "id IN ?", resourceList)
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
