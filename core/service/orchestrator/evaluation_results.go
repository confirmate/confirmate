package orchestrator

import (
	"context"
	"fmt"
	"strings"
	"time"

	"confirmate.io/core/api/evaluation"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/service"
	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// StoreEvaluationResult is a method implementation of the evaluation interface
func (svc *Service) StoreEvaluationResult(_ context.Context, req *connect.Request[orchestrator.StoreEvaluationResultRequest]) (res *connect.Response[evaluation.EvaluationResult], err error) {
	var (
		eval *evaluation.EvaluationResult
	)

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	eval = &evaluation.EvaluationResult{
		Id:                   req.Msg.GetTargetOfEvaluationId(),
		TargetOfEvaluationId: req.Msg.Result.GetTargetOfEvaluationId(),
		AuditScopeId:         req.Msg.Result.GetAuditScopeId(),
		ControlId:            req.Msg.Result.GetControlId(),
		ControlCategoryName:  req.Msg.Result.GetControlCategoryName(),
		ControlCatalogId:     req.Msg.Result.GetControlCatalogId(),
		ParentControlId:      req.Msg.Result.ParentControlId,
		Status:               req.Msg.Result.GetStatus(),
		Timestamp:            timestamppb.Now(),
		AssessmentResultIds:  req.Msg.Result.GetAssessmentResultIds(),
		Comment:              req.Msg.Result.Comment,
		ValidUntil:           req.Msg.Result.GetValidUntil(),
		Data:                 req.Msg.Result.GetData(),
	}

	err = svc.db.Create(eval)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	res = connect.NewResponse(eval)

	return res, nil
}

// ListEvaluationResults is a method implementation of the evaluation interface
func (svc *Service) ListEvaluationResults(_ context.Context,
	req *connect.Request[orchestrator.ListEvaluationResultsRequest],
) (res *connect.Response[orchestrator.ListEvaluationResultsResponse], err error) {
	var (
		query     []string
		partition []string
		args      []any
	)

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	// Filtering evaluation results by
	// * target of evaluation ID
	// * control ID
	// * sub-controls
	if req.Msg.Filter != nil {
		if req.Msg.Filter.TargetOfEvaluationId != nil {
			query = append(query, "target_of_evaluation_id = ?")
			args = append(args, req.Msg.Filter.GetTargetOfEvaluationId())
		}

		if req.Msg.Filter.CatalogId != nil {
			query = append(query, "control_catalog_id = ?")
			args = append(args, req.Msg.Filter.GetCatalogId())
		}

		if req.Msg.Filter.ControlId != nil {
			query = append(query, "control_id = ?")
			args = append(args, req.Msg.Filter.GetControlId())
		}

		// TODO(anatheka): change that, in other catalogs maybe it's not that easy to get the sub-control by name
		if req.Msg.Filter.SubControls != nil {
			partition = append(partition, "control_id")
			query = append(query, "control_id LIKE ?")
			args = append(args, fmt.Sprintf("%s%%", req.Msg.Filter.GetSubControls()))
		}

		if req.Msg.Filter.GetParentsOnly() {
			query = append(query, "parent_control_id IS NULL")
		}

		if req.Msg.Filter.GetValidManualOnly() {
			query = append(query, "status IN ?")
			args = append(args, []any{
				evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT_MANUALLY,
				evaluation.EvaluationStatus_EVALUATION_STATUS_NOT_COMPLIANT_MANUALLY,
			})

			// Use parameterized query instead of CURRENT_TIMESTAMP SQL function for compatibility with in-memory test
			// database (ramsql)
			query = append(query, "valid_until IS NULL OR valid_until >= ?")
			args = append(args, time.Now())
		}
	}

	res = &connect.Response[orchestrator.ListEvaluationResultsResponse]{Msg: &orchestrator.ListEvaluationResultsResponse{Results: make([]*evaluation.EvaluationResult, 0)}}

	// If we want to have it grouped by resource ID, we need to do a raw query
	if req.Msg.GetLatestByControlId() {
		// In the raw SQL, we need to build the whole WHERE statement
		var where string
		if len(query) > 0 {
			where = "WHERE " + strings.Join(query, " AND ")
		}

		// TODO(all): Is there a better solution? Ramsql does not support our SQL statement, so we have to do it that way for now.
		// Simple query, then reduce to "latest per control_id" in Go, because doing it in SQL is to complex for ramsql. We need to order by timestamp desc, so that the first entry per control_id is the latest one.
		sql := fmt.Sprintf(`
			SELECT *
			FROM evaluation_results
			%s
			ORDER BY control_catalog_id, control_id, timestamp DESC;
		`, where)

		err = svc.db.Raw(&res.Msg.Results, sql, args...)
		if err = service.HandleDatabaseError(err); err != nil {
			return nil, err
		}

		// Reduce results to the latest entry per control_id
		deduped := make([]*evaluation.EvaluationResult, 0, len(res.Msg.Results))
		seen := make(map[string]bool)

		for _, r := range res.Msg.Results {
			key := r.GetControlId()
			if seen[key] {
				continue
			}
			seen[key] = true
			deduped = append(deduped, r)
		}

		res.Msg.Results = deduped
	} else {
		// join query with AND and prepend the query
		args = append([]any{strings.Join(query, " AND ")}, args...)

		// Paginate the results according to the request
		res.Msg.Results, res.Msg.NextPageToken, err = service.PaginateStorage[*evaluation.EvaluationResult](req.Msg, svc.db, service.DefaultPaginationOpts, args...)
		if err = service.HandleDatabaseError(err); err != nil {
			return nil, err
		}
	}

	return
}
