// Copyright 2026 Fraunhofer AISEC
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
	"slices"

	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/auth"
	"confirmate.io/core/persistence"
	"confirmate.io/core/service"

	"buf.build/go/protovalidate"
	"connectrpc.com/connect"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// validTransitions defines the allowed state machine transitions for a ControlImplementation.
// Any transition not listed here will be rejected by TransitionControlImplementationState.
var validTransitions = map[orchestrator.ControlImplementationState][]orchestrator.ControlImplementationState{
	orchestrator.ControlImplementationState_CONTROL_IMPLEMENTATION_STATE_OPEN: {
		orchestrator.ControlImplementationState_CONTROL_IMPLEMENTATION_STATE_IN_PROGRESS,
	},
	orchestrator.ControlImplementationState_CONTROL_IMPLEMENTATION_STATE_IN_PROGRESS: {
		orchestrator.ControlImplementationState_CONTROL_IMPLEMENTATION_STATE_OPEN,
		orchestrator.ControlImplementationState_CONTROL_IMPLEMENTATION_STATE_IMPLEMENTED,
	},
	orchestrator.ControlImplementationState_CONTROL_IMPLEMENTATION_STATE_IMPLEMENTED: {
		orchestrator.ControlImplementationState_CONTROL_IMPLEMENTATION_STATE_IN_PROGRESS,
		orchestrator.ControlImplementationState_CONTROL_IMPLEMENTATION_STATE_READY_FOR_REVIEW,
	},
	orchestrator.ControlImplementationState_CONTROL_IMPLEMENTATION_STATE_READY_FOR_REVIEW: {
		orchestrator.ControlImplementationState_CONTROL_IMPLEMENTATION_STATE_IMPLEMENTED,
		orchestrator.ControlImplementationState_CONTROL_IMPLEMENTATION_STATE_ACCEPTED,
	},
	orchestrator.ControlImplementationState_CONTROL_IMPLEMENTATION_STATE_ACCEPTED: {
		orchestrator.ControlImplementationState_CONTROL_IMPLEMENTATION_STATE_IN_PROGRESS,
	},
}

// isValidTransition reports whether moving from current to next is allowed by the state machine.
func isValidTransition(current, next orchestrator.ControlImplementationState) bool {
	return slices.Contains(validTransitions[current], next)
}

// CreateControlImplementation creates a new ControlImplementation for a control within an audit
// scope. The implementation starts in the STATE_OPEN state.
func (svc *Service) CreateControlImplementation(
	ctx context.Context,
	req *connect.Request[orchestrator.CreateControlImplementationRequest],
) (res *connect.Response[orchestrator.ControlImplementation], err error) {
	var (
		impl    *orchestrator.ControlImplementation
		scope   orchestrator.AuditScope
		allowed bool
	)

	// Validate the request, ignoring ID fields that will be auto-generated.
	if err = service.Validate(req, protovalidate.WithFilter(service.IgnoreIDFilter)); err != nil {
		return nil, err
	}

	// Look up the audit scope to obtain the target_of_evaluation_id for authorization.
	err = svc.db.Get(&scope, "id = ?", req.Msg.GetControlImplementation().GetAuditScopeId())
	if err = service.HandleDatabaseError(err, service.ErrNotFound("audit scope")); err != nil {
		return nil, err
	}

	// Check access — permissions are scoped to the audit scope.
	allowed, _, err = CheckAccess(ctx, svc.authz, svc,
		orchestrator.RequestType_REQUEST_TYPE_CREATED,
		req.Msg.GetControlImplementation().GetAuditScopeId(),
		orchestrator.ObjectType_OBJECT_TYPE_CONTROL_IMPLEMENTATION,
	)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if !allowed {
		return nil, service.ErrPermissionDenied
	}

	now := timestamppb.Now()
	impl = &orchestrator.ControlImplementation{
		Id:                       uuid.NewString(),
		AuditScopeId:             req.Msg.GetControlImplementation().GetAuditScopeId(),
		TargetOfEvaluationId:     scope.TargetOfEvaluationId,
		ControlId:                req.Msg.GetControlImplementation().GetControlId(),
		ControlCategoryName:      req.Msg.GetControlImplementation().GetControlCategoryName(),
		ControlCategoryCatalogId: req.Msg.GetControlImplementation().GetControlCategoryCatalogId(),
		State:                    orchestrator.ControlImplementationState_CONTROL_IMPLEMENTATION_STATE_OPEN,
		AssigneeId:               req.Msg.GetControlImplementation().AssigneeId,
		CreatedAt:                now,
		UpdatedAt:                now,
	}

	// Check that no ControlImplementation already exists for the same (audit scope, control) pair.
	// The unique index on (audit_scope_id, control_id, control_category_name, control_category_catalog_id)
	// also enforces this at the DB level, but we check early to return a clear error.
	var duplicate orchestrator.ControlImplementation
	err = svc.db.Get(&duplicate, persistence.WithoutPreload(),
		"audit_scope_id = ? AND control_id = ? AND control_category_name = ? AND control_category_catalog_id = ?",
		impl.AuditScopeId, impl.ControlId, impl.ControlCategoryName, impl.ControlCategoryCatalogId,
	)
	if err == nil {
		return nil, connect.NewError(connect.CodeAlreadyExists, service.ErrResourceAlreadyExists)
	} else if !errors.Is(err, persistence.ErrRecordNotFound) {
		if err = service.HandleDatabaseError(err); err != nil {
			return nil, err
		}
	}

	err = svc.db.Create(impl)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	res = connect.NewResponse(impl)
	return
}

// GetControlImplementation retrieves a control implementation by ID, including its full transition
// history.
func (svc *Service) GetControlImplementation(
	ctx context.Context,
	req *connect.Request[orchestrator.GetControlImplementationRequest],
) (res *connect.Response[orchestrator.ControlImplementation], err error) {
	var (
		impl    orchestrator.ControlImplementation
		allowed bool
	)

	if err = service.Validate(req); err != nil {
		return nil, err
	}

	err = svc.db.Get(&impl, "id = ?", req.Msg.GetId())
	if err = service.HandleDatabaseError(err, service.ErrNotFound("control implementation")); err != nil {
		return nil, err
	}

	allowed, _, err = CheckAccess(ctx, svc.authz, svc,
		orchestrator.RequestType_REQUEST_TYPE_GET,
		impl.AuditScopeId,
		orchestrator.ObjectType_OBJECT_TYPE_CONTROL_IMPLEMENTATION,
	)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if !allowed {
		return nil, service.ErrPermissionDenied
	}

	res = connect.NewResponse(&impl)
	return
}

// ListControlImplementations lists control implementations with optional filtering.
func (svc *Service) ListControlImplementations(
	ctx context.Context,
	req *connect.Request[orchestrator.ListControlImplementationsRequest],
) (res *connect.Response[orchestrator.ListControlImplementationsResponse], err error) {
	var (
		impls  []*orchestrator.ControlImplementation
		conds  []any
		npt    string
		all    bool
		toeIds []string
	)

	if err = service.Validate(req); err != nil {
		return nil, err
	}

	if req.Msg.OrderBy == "" {
		req.Msg.OrderBy = "created_at"
		req.Msg.Asc = true
	}

	all, toeIds = svc.authz.AllowedTargetOfEvaluations(ctx)
	if !all && len(toeIds) == 0 {
		return connect.NewResponse(&orchestrator.ListControlImplementationsResponse{
			ControlImplementations: []*orchestrator.ControlImplementation{},
		}), nil
	}
	if !all {
		conds = append(conds, "target_of_evaluation_id IN ?", toeIds)
	}

	if f := req.Msg.GetFilter(); f != nil {
		if f.AuditScopeId != nil {
			conds = append(conds, "audit_scope_id = ?", f.GetAuditScopeId())
		}
		if f.State != nil {
			conds = append(conds, "state = ?", f.GetState())
		}
		if f.AssigneeId != nil {
			conds = append(conds, "assignee_id = ?", f.GetAssigneeId())
		}
	}

	impls, npt, err = service.PaginateStorage[*orchestrator.ControlImplementation](req.Msg, svc.db, service.DefaultPaginationOpts, conds...)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	res = connect.NewResponse(&orchestrator.ListControlImplementationsResponse{
		ControlImplementations: impls,
		NextPageToken:          npt,
	})
	return
}

// UpdateControlImplementation updates mutable fields of an existing control implementation.
// Only assignee_id can be updated; structural fields (audit scope, control reference, state) must
// use their dedicated RPCs.
func (svc *Service) UpdateControlImplementation(
	ctx context.Context,
	req *connect.Request[orchestrator.UpdateControlImplementationRequest],
) (res *connect.Response[orchestrator.ControlImplementation], err error) {
	var (
		existing orchestrator.ControlImplementation
		allowed  bool
	)

	if err = service.Validate(req); err != nil {
		return nil, err
	}

	err = svc.db.Get(&existing, persistence.WithoutPreload(), "id = ?", req.Msg.GetId())
	if err = service.HandleDatabaseError(err, service.ErrNotFound("control implementation")); err != nil {
		return nil, err
	}

	allowed, _, err = CheckAccess(ctx, svc.authz, svc,
		orchestrator.RequestType_REQUEST_TYPE_UPDATED,
		existing.AuditScopeId,
		orchestrator.ObjectType_OBJECT_TYPE_CONTROL_IMPLEMENTATION,
	)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if !allowed {
		return nil, service.ErrPermissionDenied
	}

	// Only allow updating assignee_id and implementation_details; preserve all other fields.
	existing.AssigneeId = req.Msg.AssigneeId
	existing.ImplementationDetails = req.Msg.ImplementationDetails
	existing.UpdatedAt = timestamppb.Now()

	err = svc.db.Update(&existing, "id = ?", existing.Id)
	if err = service.HandleDatabaseError(err, service.ErrNotFound("control implementation")); err != nil {
		return nil, err
	}

	res = connect.NewResponse(&existing)
	return
}

// TransitionControlImplementationState moves a control implementation to a new state, enforcing
// the state machine defined in validTransitions and recording the transition in the history.
func (svc *Service) TransitionControlImplementationState(
	ctx context.Context,
	req *connect.Request[orchestrator.TransitionControlImplementationStateRequest],
) (res *connect.Response[orchestrator.ControlImplementation], err error) {
	var (
		impl    orchestrator.ControlImplementation
		allowed bool
	)

	if err = service.Validate(req); err != nil {
		return nil, err
	}

	err = svc.db.Get(&impl, "id = ?", req.Msg.GetId())
	if err = service.HandleDatabaseError(err, service.ErrNotFound("control implementation")); err != nil {
		return nil, err
	}

	allowed, _, err = CheckAccess(ctx, svc.authz, svc,
		orchestrator.RequestType_REQUEST_TYPE_UPDATED,
		impl.AuditScopeId,
		orchestrator.ObjectType_OBJECT_TYPE_CONTROL_IMPLEMENTATION,
	)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if !allowed {
		return nil, service.ErrPermissionDenied
	}

	toState := req.Msg.GetToState()
	if !isValidTransition(impl.State, toState) {
		return nil, connect.NewError(connect.CodeInvalidArgument,
			fmt.Errorf("invalid state transition: %s → %s",
				impl.State.String(), toState.String()))
	}

	// Determine the performing user from the authorization context.
	var userId string
	if claims, ok := auth.ClaimsFromContext(ctx); ok {
		userId = auth.GetConfirmateUserIDFromClaims(claims)
	}

	// Use a transaction to ensure atomicity of the transition record creation and state update.
	err = svc.db.Transaction(func(tx persistence.DB) error {
		transition := &orchestrator.ControlImplementationTransition{
			Id:                      uuid.NewString(),
			ControlImplementationId: impl.Id,
			FromState:               impl.State,
			ToState:                 toState,
			PerformedBy:             userId,
			Time:                    timestamppb.Now(),
		}

		if err = tx.Create(transition); err != nil {
			return err
		}

		impl.State = toState
		impl.UpdatedAt = timestamppb.Now()

		if err = tx.Update(&impl, "id = ?", impl.Id); err != nil {
			return err
		}

		return nil
	})
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	// Reload with full transition history.
	err = svc.db.Get(&impl, "id = ?", impl.Id)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	res = connect.NewResponse(&impl)
	return
}

// RemoveControlImplementation deletes a control implementation and its transition history.
func (svc *Service) RemoveControlImplementation(
	ctx context.Context,
	req *connect.Request[orchestrator.RemoveControlImplementationRequest],
) (res *connect.Response[emptypb.Empty], err error) {
	var (
		impl    orchestrator.ControlImplementation
		allowed bool
	)

	if err = service.Validate(req); err != nil {
		return nil, err
	}

	err = svc.db.Get(&impl, "id = ?", req.Msg.GetId())
	if err = service.HandleDatabaseError(err, service.ErrNotFound("control implementation")); err != nil {
		return nil, err
	}

	allowed, _, err = CheckAccess(ctx, svc.authz, svc,
		orchestrator.RequestType_REQUEST_TYPE_DELETED,
		impl.AuditScopeId,
		orchestrator.ObjectType_OBJECT_TYPE_CONTROL_IMPLEMENTATION,
	)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if !allowed {
		return nil, service.ErrPermissionDenied
	}

	err = svc.db.Delete(&impl, "id = ?", impl.Id)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	res = connect.NewResponse(&emptypb.Empty{})
	return
}
