// Copyright 2016-2026 Fraunhofer AISEC
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
	"time"

	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/persistence"
	"confirmate.io/core/service"

	"buf.build/go/protovalidate"
	"connectrpc.com/connect"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// reviewInterval is the time after the approval of a risk assessment by which it has to be reviewed, if the assessment
// does not set a time itself.
const reviewInterval = 365 * 24 * time.Hour

// CreateRiskAssessment stores a risk assessment of an audit scope as a draft. The scope decisions are applied when it
// is approved.
func (svc *Service) CreateRiskAssessment(
	ctx context.Context,
	req *connect.Request[orchestrator.CreateRiskAssessmentRequest],
) (res *connect.Response[orchestrator.RiskAssessment], err error) {
	var (
		input      *orchestrator.RiskAssessment
		scope      orchestrator.AuditScope
		assessment *orchestrator.RiskAssessment
		allowed    bool
		now        = timestamppb.Now()
	)

	// Validate the request, ignoring the ID, which is generated
	if err = service.Validate(req, protovalidate.WithFilter(service.IgnoreIDFilter)); err != nil {
		return nil, err
	}

	input = req.Msg.GetRiskAssessment()

	err = svc.db.Get(&scope, persistence.WithoutPreload(), "id = ?", input.GetAuditScopeId())
	if err = service.HandleDatabaseError(err, service.ErrNotFound("audit scope")); err != nil {
		return nil, err
	}

	if scope.TargetOfEvaluationId != input.GetTargetOfEvaluationId() {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("the audit scope does not belong to the target of evaluation"))
	}

	// A risk assessment changes the scope of the audit scope, so it needs the permission to change it.
	allowed, _, err = CheckAccess(ctx, svc.authz, svc, orchestrator.RequestType_REQUEST_TYPE_UPDATED, scope.Id, orchestrator.ObjectType_OBJECT_TYPE_AUDIT_SCOPE)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if !allowed {
		return nil, service.ErrPermissionDenied
	}

	err = svc.validateScopeDecisions(&scope, input.GetDecisions())
	if err != nil {
		return nil, err
	}

	assessment = &orchestrator.RiskAssessment{
		Id:                   uuid.NewString(),
		TargetOfEvaluationId: scope.TargetOfEvaluationId,
		AuditScopeId:         scope.Id,
		Status:               orchestrator.RiskAssessmentStatus_RISK_ASSESSMENT_STATUS_DRAFT,
		Method:               input.GetMethod(),
		Tool:                 input.Tool,
		ToolVersion:          input.ToolVersion,
		Summary:              input.GetSummary(),
		DocumentEvidenceId:   input.DocumentEvidenceId,
		Answers:              input.GetAnswers(),
		Decisions:            input.GetDecisions(),
		CreatedAt:            now,
		ReviewDueAt:          input.GetReviewDueAt(),
	}
	if actor := actorFromContext(ctx); actor != "" {
		assessment.CreatedBy = &actor
	}

	err = svc.db.Transaction(func(tx persistence.DB) error {
		// The version counts the assessments of the audit scope
		count, err := tx.Count(&orchestrator.RiskAssessment{}, "audit_scope_id = ?", scope.Id)
		if err != nil {
			return service.HandleDatabaseError(err)
		}
		assessment.Version = int32(count) + 1

		return tx.Create(assessment)
	})
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	res = connect.NewResponse(assessment)
	return
}

// GetRiskAssessment retrieves a risk assessment by ID.
func (svc *Service) GetRiskAssessment(
	ctx context.Context,
	req *connect.Request[orchestrator.GetRiskAssessmentRequest],
) (res *connect.Response[orchestrator.RiskAssessment], err error) {
	var (
		assessment orchestrator.RiskAssessment
		allowed    bool
	)

	if err = service.Validate(req); err != nil {
		return nil, err
	}

	err = svc.db.Get(&assessment, "id = ?", req.Msg.GetRiskAssessmentId())
	if err = service.HandleDatabaseError(err, service.ErrNotFound("risk assessment")); err != nil {
		return nil, err
	}

	allowed, _, err = CheckAccess(ctx, svc.authz, svc, orchestrator.RequestType_REQUEST_TYPE_GET, assessment.AuditScopeId, orchestrator.ObjectType_OBJECT_TYPE_AUDIT_SCOPE)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if !allowed {
		return nil, service.ErrPermissionDenied
	}

	res = connect.NewResponse(&assessment)
	return
}

// ListRiskAssessments lists the risk assessments of the audit scopes that the caller can access.
func (svc *Service) ListRiskAssessments(
	ctx context.Context,
	req *connect.Request[orchestrator.ListRiskAssessmentsRequest],
) (res *connect.Response[orchestrator.ListRiskAssessmentsResponse], err error) {
	var (
		assessments   []*orchestrator.RiskAssessment
		conds         []any
		npt           string
		all           bool
		auditScopeIds []string
	)

	if err = service.Validate(req); err != nil {
		return nil, err
	}

	// The newest assessment comes first by default
	if req.Msg.OrderBy == "" {
		req.Msg.OrderBy = "created_at"
		req.Msg.Asc = false
	}

	if req.Msg.Filter != nil && req.Msg.Filter.AuditScopeId != nil {
		conds = append(conds, "audit_scope_id = ?", req.Msg.Filter.GetAuditScopeId())
	}
	if req.Msg.Filter != nil && req.Msg.Filter.TargetOfEvaluationId != nil {
		conds = append(conds, "target_of_evaluation_id = ?", req.Msg.Filter.GetTargetOfEvaluationId())
	}

	// Only the assessments of the audit scopes that the caller can access are listed
	all, auditScopeIds = svc.authz.AllowedAuditScopes(ctx)
	if !all && len(auditScopeIds) == 0 {
		return connect.NewResponse(&orchestrator.ListRiskAssessmentsResponse{
			RiskAssessments: []*orchestrator.RiskAssessment{},
		}), nil
	}
	if !all {
		conds = append(conds, "audit_scope_id IN ?", auditScopeIds)
	}

	assessments, npt, err = service.PaginateStorage[*orchestrator.RiskAssessment](req.Msg, svc.db, service.DefaultPaginationOpts, conds...)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	res = connect.NewResponse(&orchestrator.ListRiskAssessmentsResponse{
		RiskAssessments: assessments,
		NextPageToken:   npt,
	})
	return
}

// ApproveRiskAssessment approves a draft risk assessment and applies its scope decisions to the audit scope. The
// assessment that was approved before is superseded: controls that it took out of scope and that the new assessment
// does not take out come back into scope.
func (svc *Service) ApproveRiskAssessment(
	ctx context.Context,
	req *connect.Request[orchestrator.ApproveRiskAssessmentRequest],
) (res *connect.Response[orchestrator.RiskAssessment], err error) {
	var (
		assessment orchestrator.RiskAssessment
		scope      orchestrator.AuditScope
		allowed    bool
		now        = timestamppb.Now()
	)

	if err = service.Validate(req); err != nil {
		return nil, err
	}

	err = svc.db.Get(&assessment, "id = ?", req.Msg.GetRiskAssessmentId())
	if err = service.HandleDatabaseError(err, service.ErrNotFound("risk assessment")); err != nil {
		return nil, err
	}

	err = svc.db.Get(&scope, persistence.WithoutPreload(), "id = ?", assessment.AuditScopeId)
	if err = service.HandleDatabaseError(err, service.ErrNotFound("audit scope")); err != nil {
		return nil, err
	}

	allowed, _, err = CheckAccess(ctx, svc.authz, svc, orchestrator.RequestType_REQUEST_TYPE_UPDATED, scope.Id, orchestrator.ObjectType_OBJECT_TYPE_AUDIT_SCOPE)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if !allowed {
		return nil, service.ErrPermissionDenied
	}

	if assessment.Status != orchestrator.RiskAssessmentStatus_RISK_ASSESSMENT_STATUS_DRAFT {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("only a draft risk assessment can be approved"))
	}

	actor := actorFromContext(ctx)
	err = svc.db.Transaction(func(tx persistence.DB) error {
		var previous []*orchestrator.RiskAssessment

		err := tx.List(&previous, "id", true, 0, -1, "audit_scope_id = ? AND status = ?",
			scope.Id, orchestrator.RiskAssessmentStatus_RISK_ASSESSMENT_STATUS_APPROVED)
		if err != nil {
			return err
		}

		// The new assessment is the complete statement about the scope: what the previous ones took out and the new
		// one does not take out comes back.
		outBefore := make(map[string]struct{})
		for _, p := range previous {
			for _, d := range p.GetDecisions() {
				if !d.GetInScope() {
					outBefore[d.GetControlId()] = struct{}{}
				}
			}

			p.Status = orchestrator.RiskAssessmentStatus_RISK_ASSESSMENT_STATUS_SUPERSEDED
			if err = tx.Save(p); err != nil {
				return err
			}
		}

		err = applyScopeDecisions(ctx, tx, &scope, &assessment, outBefore, actor, now)
		if err != nil {
			return err
		}

		assessment.Status = orchestrator.RiskAssessmentStatus_RISK_ASSESSMENT_STATUS_APPROVED
		assessment.ApprovedAt = now
		if actor != "" {
			assessment.ApprovedBy = &actor
		}
		if assessment.ReviewDueAt == nil {
			assessment.ReviewDueAt = timestamppb.New(now.AsTime().Add(reviewInterval))
		}

		return tx.Save(&assessment)
	})
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	// Evaluate the audit scope again, so that the controls that changed are evaluated or not evaluated anymore.
	svc.triggerScopeChangeCallback(scope.Id)

	res = connect.NewResponse(&assessment)
	return
}

// validateScopeDecisions checks that the decisions are about controls of the catalog of the audit scope and that
// there is at most one decision per control. A whole control cannot be taken out of scope: its requirements (the
// sub-controls) have to be decided one by one, and at least one of them has to stay in scope. Otherwise a point of the
// legal text could silently drop out because of one decision.
func (svc *Service) validateScopeDecisions(scope *orchestrator.AuditScope, decisions []*orchestrator.ScopeDecision) error {
	var (
		seen = make(map[string]struct{}, len(decisions))
		// The number of decisions that take a sub-control out of scope, by the ID of its parent control
		outByParent = make(map[string]int)
	)

	for _, d := range decisions {
		if _, ok := seen[d.GetControlId()]; ok {
			return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("more than one decision for control %s", d.GetControlId()))
		}
		seen[d.GetControlId()] = struct{}{}

		var ctrl orchestrator.Control
		err := svc.db.Get(&ctrl, persistence.WithoutPreload(), "id = ? AND catalog_id = ?", d.GetControlId(), scope.CatalogId)
		if errors.Is(err, persistence.ErrRecordNotFound) {
			return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("control %s is not part of the catalog of the audit scope", d.GetControlId()))
		}
		if err != nil {
			return service.HandleDatabaseError(err)
		}

		if d.GetInScope() {
			continue
		}

		children, err := svc.db.Count(&orchestrator.Control{}, "parent_control_id = ?", ctrl.Id)
		if err != nil {
			return service.HandleDatabaseError(err)
		}
		if children > 0 {
			return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("control %s cannot be taken out of scope as a whole, decide on its requirements", ctrl.GetShortName()))
		}

		if ctrl.ParentControlId != nil {
			outByParent[*ctrl.ParentControlId]++
		}
	}

	for parent, out := range outByParent {
		children, err := svc.db.Count(&orchestrator.Control{}, "parent_control_id = ?", parent)
		if err != nil {
			return service.HandleDatabaseError(err)
		}
		if int64(out) >= children {
			return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("at least one requirement of control %s has to stay in scope", parent))
		}
	}

	return nil
}

// applyScopeDecisions brings the controls in scope of the audit scope in line with the decisions of an assessment.
// Controls that are out of scope are removed with the justification as the reason in the audit trail. Controls that
// were out of scope before or that are in scope according to the assessment are added again, together with their
// sub-controls that are not out of scope themselves. It has to be called within a transaction.
func applyScopeDecisions(ctx context.Context, tx persistence.DB, scope *orchestrator.AuditScope, assessment *orchestrator.RiskAssessment, outBefore map[string]struct{}, actor string, now *timestamppb.Timestamp) error {
	var (
		out   = make(map[string]struct{})
		comes []string
	)

	for _, d := range assessment.GetDecisions() {
		if d.GetInScope() {
			comes = append(comes, d.GetControlId())
			continue
		}
		out[d.GetControlId()] = struct{}{}

		var cis orchestrator.ControlInScope
		err := tx.Get(&cis, "audit_scope_id = ? AND control_id = ?", scope.Id, d.GetControlId())
		if errors.Is(err, persistence.ErrRecordNotFound) {
			// Already out of scope, for example because of its parent control
			continue
		}
		if err != nil {
			return err
		}

		comment := fmt.Sprintf("Risk assessment, version %d: %s", assessment.GetVersion(), d.GetJustification())
		if err = removeControlInScope(tx, actor, &cis, comment); err != nil {
			return err
		}
	}

	for id := range outBefore {
		if _, stillOut := out[id]; !stillOut {
			comes = append(comes, id)
		}
	}

	for _, id := range comes {
		var ctrl orchestrator.Control
		if err := tx.Get(&ctrl, persistence.WithoutPreload(), "id = ?", id); err != nil {
			if errors.Is(err, persistence.ErrRecordNotFound) {
				continue
			}
			return err
		}

		// A sub-control cannot be in scope while its parent control is not. If the parent comes back itself, it brings
		// the sub-control with it.
		if ctrl.ParentControlId != nil {
			count, err := tx.Count(&orchestrator.ControlInScope{}, "audit_scope_id = ? AND control_id = ?", scope.Id, *ctrl.ParentControlId)
			if err != nil {
				return err
			}
			if count == 0 {
				continue
			}
		}

		if err := addControlSubtreeInScope(ctx, tx, scope, &ctrl, out, now); err != nil {
			return err
		}
	}

	return nil
}

// addControlSubtreeInScope brings a control and all of its sub-controls into the scope of an audit scope, unless they
// are in scope already, do not match the assurance level of the scope or are in the skip set.
func addControlSubtreeInScope(ctx context.Context, tx persistence.DB, scope *orchestrator.AuditScope, ctrl *orchestrator.Control, skip map[string]struct{}, now *timestamppb.Timestamp) error {
	var children []*orchestrator.Control

	if _, ok := skip[ctrl.Id]; ok || !matchesAssuranceLevel(scope, ctrl) {
		return nil
	}

	count, err := tx.Count(&orchestrator.ControlInScope{}, "audit_scope_id = ? AND control_id = ?", scope.Id, ctrl.Id)
	if err != nil {
		return err
	}
	if count == 0 {
		if err = addControlInScope(ctx, tx, scope, ctrl.Id, now); err != nil {
			return err
		}
	}

	if err = tx.List(&children, "id", true, 0, -1, persistence.WithoutPreload(), "parent_control_id = ?", ctrl.Id); err != nil {
		return err
	}
	for _, child := range children {
		if err = addControlSubtreeInScope(ctx, tx, scope, child, skip, now); err != nil {
			return err
		}
	}

	return nil
}
