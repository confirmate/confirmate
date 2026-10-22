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
	"confirmate.io/core/persistence"
	"confirmate.io/core/service"
	"confirmate.io/core/util/timestamp"

	"connectrpc.com/connect"
	"github.com/google/uuid"
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
	allowed, _, err := CheckAccess(ctx, svc.authz, svc, orchestrator.RequestType_REQUEST_TYPE_CREATED, result.GetTargetOfEvaluationId(), orchestrator.ObjectType_OBJECT_TYPE_ASSESSMENT_RESULT)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("%w: %w", service.ErrDatabaseError, err))
	}
	if !allowed {
		return nil, service.ErrPermissionDenied
	}

	// Normalize timestamps in the assessment result to microsecond precision
	timestamp.NormalizeTimestampAssessmentResult(result)

	// Set timestamp
	result.CreatedAt = timestamp.Now()

	// Persist the assessment result in the database
	err = svc.db.Create(result)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	// Notify subscribers
	go svc.publishEvent(&orchestrator.ChangeEvent{
		Timestamp:   timestamp.Now(),
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

// StoreManualAssessmentResult stores an assessment result that was supplied manually for a single metric of a
// target of evaluation, optionally for a specific resource. It is additional information only: the evaluation prefers
// an automatic result over it. Because the latest result per resource and metric counts, a manual result for a
// resource is rejected if an automatic result of the metric exists for it already. A newer automatic result replaces
// a manual one on its own.
func (svc *Service) StoreManualAssessmentResult(
	ctx context.Context,
	req *connect.Request[orchestrator.StoreManualAssessmentResultRequest],
) (res *connect.Response[assessment.AssessmentResult], err error) {
	var (
		allowed       bool
		metric        assessment.Metric
		result        *assessment.AssessmentResult
		resourceId    = assessment.ManualTargetOfEvaluationResourceId(req.Msg.GetTargetOfEvaluationId())
		resourceTypes = []string{"Manual"}
		existing      assessment.AssessmentResult
		count         int64
		now           = timestamp.Now()
	)

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	// Check access via the configured auth strategy
	allowed, _, err = CheckAccess(ctx, svc.authz, svc, orchestrator.RequestType_REQUEST_TYPE_CREATED, req.Msg.GetTargetOfEvaluationId(), orchestrator.ObjectType_OBJECT_TYPE_ASSESSMENT_RESULT)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("%w: %w", service.ErrDatabaseError, err))
	}
	if !allowed {
		return nil, service.ErrPermissionDenied
	}

	// The metric must exist
	err = svc.db.Get(&metric, "id = ?", req.Msg.GetMetricId())
	if err = service.HandleDatabaseError(err, service.ErrNotFound("metric")); err != nil {
		return nil, err
	}

	if req.Msg.ResourceId != nil {
		resourceId = req.Msg.GetResourceId()

		// An automatic result of the same resource and metric takes precedence, but it would be hidden by a newer
		// manual result, since only the latest result per resource and metric counts.
		count, err = svc.db.Count(&assessment.AssessmentResult{}, "target_of_evaluation_id = ? AND resource_id = ? AND metric_id = ? AND manual = ?",
			req.Msg.GetTargetOfEvaluationId(), resourceId, req.Msg.GetMetricId(), false)
		if err = service.HandleDatabaseError(err); err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("an automatic result of the metric exists for the resource already"))
		}

		// Take over the types of the resource from another result of it, if there is one
		err = svc.db.Get(&existing, "target_of_evaluation_id = ? AND resource_id = ?", req.Msg.GetTargetOfEvaluationId(), resourceId)
		if err == nil {
			resourceTypes = existing.GetResourceTypes()
		} else if !errors.Is(err, persistence.ErrRecordNotFound) {
			return nil, service.HandleDatabaseError(err)
		}
	}

	result = &assessment.AssessmentResult{
		Id:        uuid.NewString(),
		CreatedAt: now,
		MetricId:  req.Msg.GetMetricId(),
		MetricConfiguration: &assessment.MetricConfiguration{
			MetricId:             req.Msg.GetMetricId(),
			TargetOfEvaluationId: req.Msg.GetTargetOfEvaluationId(),
		},
		Compliant:            req.Msg.GetCompliant(),
		EvidenceId:           req.Msg.GetEvidenceId(),
		ResourceId:           resourceId,
		ResourceTypes:        resourceTypes,
		ComplianceComment:    req.Msg.GetComment(),
		TargetOfEvaluationId: req.Msg.GetTargetOfEvaluationId(),
		ToolId:               new("manual"),
		HistoryUpdatedAt:     now,
		History:              []*assessment.Record{},
		Manual:               true,
		ValidUntil:           req.Msg.ValidUntil,
	}
	if author := actorFromContext(ctx); author != "" {
		result.Author = &author
	}

	err = svc.db.Create(result)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	// Notify subscribers
	go svc.publishEvent(&orchestrator.ChangeEvent{
		Timestamp:   timestamp.Now(),
		Category:    orchestrator.EventCategory_EVENT_CATEGORY_ASSESSMENT_RESULT,
		RequestType: orchestrator.RequestType_REQUEST_TYPE_CREATED,
		EntityId:    result.Id,
		Entity: &orchestrator.ChangeEvent_AssessmentResult{
			AssessmentResult: result,
		},
	})

	res = connect.NewResponse(result)
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
	allowed, _, err = CheckAccess(ctx, svc.authz, svc, orchestrator.RequestType_REQUEST_TYPE_GET, result.GetTargetOfEvaluationId(), orchestrator.ObjectType_OBJECT_TYPE_ASSESSMENT_RESULT)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if !allowed {
		return nil, service.ErrPermissionDenied
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
		npt     string
		where   string
		args    []any
		query   []string
		all     bool
		toeIds  []string
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

	// Apply filters if provided
	if req.Msg.Filter != nil {
		if req.Msg.Filter.TargetOfEvaluationId != nil {
			query = append(query, "target_of_evaluation_id = ?")
			args = append(args, req.Msg.Filter.GetTargetOfEvaluationId())
		}
		if req.Msg.Filter.Compliant != nil {
			query = append(query, "compliant = ?")
			args = append(args, req.Msg.Filter.GetCompliant())
		}
		if req.Msg.Filter.MetricId != nil {
			query = append(query, "metric_id = ?")
			args = append(args, req.Msg.Filter.GetMetricId())
		}
		if len(req.Msg.Filter.MetricIds) > 0 {
			query, args = persistence.AppendObjectIds(req.Msg.Filter.MetricIds, query, args, "metric_id")
		}
		if req.Msg.Filter.ToolId != nil {
			query = append(query, "tool_id = ?")
			args = append(args, req.Msg.Filter.GetToolId())
		}
		if len(req.Msg.Filter.AssessmentResultIds) > 0 {
			// Build IN clause dynamically to support ramsql (doesn't support array binding)
			query, args = persistence.AppendObjectIds(req.Msg.Filter.AssessmentResultIds, query, args, "id")
		}
		if req.Msg.Filter.EvidenceId != nil {
			query = append(query, "evidence_id = ?")
			args = append(args, req.Msg.Filter.GetEvidenceId())
		}
	}

	// Retrieve list of all allowed ToE IDs for the user to filter results by access permissions.
	all, toeIds = svc.authz.AllowedTargetOfEvaluations(ctx)
	if !all && len(toeIds) == 0 {
		// User has no access to any ToE, return empty result
		return connect.NewResponse(&orchestrator.ListAssessmentResultsResponse{
			Results:       []*assessment.AssessmentResult{},
			NextPageToken: "",
		}), nil
	}

	// If access is not allowed to all objects, add a condition to filter by the allowed object IDs
	// Note: The authorization filter is added in addition to any request ToE filter.
	// Since all where clauses are combined with AND later, a requested ToE must also
	// be part of the allowed toeIds; otherwise, the query returns no results.
	if !all {
		query, args = persistence.AppendObjectIds(toeIds, query, args, "target_of_evaluation_id")
	}

	// Handle latest_by_resource_id filter
	// This returns only the most recent assessment result for each unique (resource_id, metric_id) pair
	// Uses PostgreSQL's DISTINCT ON for efficient grouping
	if req.Msg.LatestByResourceId != nil && req.Msg.GetLatestByResourceId() {
		// Combine all WHERE clauses with AND and reuse the query and args directly.
		where = strings.Join(query, " AND ")
		if where != "" {
			where = "WHERE " + where
		}

		// Only select the IDs here (see loadByIds), the full rows of the page are loaded afterwards.
		rawQuery := fmt.Sprintf(`
				SELECT DISTINCT ON (resource_id, metric_id) id
				FROM assessment_results
				%s
				ORDER BY resource_id, metric_id, created_at DESC, id DESC
				LIMIT ? OFFSET ?
			`, where)

		results, npt, err = service.PaginateRaw[*assessment.AssessmentResult](
			req.Msg,
			service.DefaultPaginationOpts,
			func(start int64, size int32) ([]*assessment.AssessmentResult, error) {
				var (
					idRows []*assessment.AssessmentResult
					ids    []string
				)

				queryArgs := append([]any(nil), args...)
				queryArgs = append(queryArgs, int(size), int(start))

				if err := svc.db.Raw(&idRows, rawQuery, queryArgs...); err != nil {
					return nil, err
				}

				ids = make([]string, 0, len(idRows))
				for _, r := range idRows {
					ids = append(ids, r.GetId())
				}

				return loadByIds[*assessment.AssessmentResult](svc.db, ids)
			},
		)
		if err = service.HandleDatabaseError(err); err != nil {
			return nil, err
		}

		return connect.NewResponse(&orchestrator.ListAssessmentResultsResponse{
			Results:       results,
			NextPageToken: npt,
		}), nil
	}

	results, npt, err = service.PaginateStorage[*assessment.AssessmentResult](req.Msg, svc.db, service.DefaultPaginationOpts, persistence.BuildConds(query, args)...)
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
