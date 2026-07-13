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

	"confirmate.io/core/api/evaluation"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/auth"
	"confirmate.io/core/persistence"
	"confirmate.io/core/service"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// validTransitions defines the allowed state machine transitions for a ControlInScope.
// Any transition not listed here will be rejected by TransitionControlInScopeState.
var validTransitions = map[orchestrator.ControlInScopeState][]orchestrator.ControlInScopeState{
	orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_OPEN: {
		orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_IN_PROGRESS,
	},
	orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_IN_PROGRESS: {
		orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_OPEN,
		orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_IMPLEMENTED,
	},
	orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_IMPLEMENTED: {
		orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_IN_PROGRESS,
		orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_READY_FOR_REVIEW,
	},
	orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_READY_FOR_REVIEW: {
		orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_IN_PROGRESS,
		orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_ACCEPTED,
	},
	orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_ACCEPTED: {
		orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_IN_PROGRESS,
	},
}

// isValidTransition reports whether moving from current to next is allowed by the state machine.
func isValidTransition(current, next orchestrator.ControlInScopeState) bool {
	return slices.Contains(validTransitions[current], next)
}

// actorFromContext returns the Confirmate user ID from the request context, or empty string if
// authentication context is not present.
func actorFromContext(ctx context.Context) string {
	if claims, ok := auth.ClaimsFromContext(ctx); ok {
		return auth.GetConfirmateUserIDFromClaims(claims)
	}
	return ""
}

// createAuditTrailEvent persists a new AuditTrailEvent with the given payload to the database.
// controlInScopeId may be empty for events where the ControlInScope record no longer exists.
func createAuditTrailEvent(db persistence.DB, actorId, auditScopeId, controlInScopeId, comment string, payload proto.Message) error {
	data, err := anypb.New(payload)
	if err != nil {
		return fmt.Errorf("failed to pack audit trail event data: %w", err)
	}
	event := &orchestrator.AuditTrailEvent{
		Id:           uuid.NewString(),
		AuditScopeId: auditScopeId,
		ActorId:      actorId,
		Comment:      comment,
		CreatedAt:    timestamppb.Now(),
		EventData:    data,
	}
	if controlInScopeId != "" {
		event.ControlInScopeId = &controlInScopeId
	}
	return db.Create(event)
}

// CreateControlInScope manually brings a control into scope within an audit scope. Controls are
// also brought in scope automatically when an audit scope is created.
func (svc *Service) CreateControlInScope(
	ctx context.Context,
	req *connect.Request[orchestrator.CreateControlInScopeRequest],
) (res *connect.Response[orchestrator.ControlInScope], err error) {
	var (
		cis     *orchestrator.ControlInScope
		allowed bool
	)

	if err = service.Validate(req); err != nil {
		return nil, err
	}

	allowed, _, err = CheckAccess(ctx, svc.authz, svc,
		orchestrator.RequestType_REQUEST_TYPE_CREATED,
		req.Msg.GetAuditScopeId(),
		orchestrator.ObjectType_OBJECT_TYPE_CONTROL_IN_SCOPE,
	)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if !allowed {
		return nil, service.ErrPermissionDenied
	}

	// Verify the control exists (control IDs are globally unique UUIDs since #271).
	var ctrl orchestrator.Control
	err = svc.db.Get(&ctrl, persistence.WithoutPreload(), "id = ?", req.Msg.GetControlId())
	if err = service.HandleDatabaseError(err, service.ErrNotFound("control")); err != nil {
		return nil, err
	}

	var duplicate orchestrator.ControlInScope
	err = svc.db.Get(&duplicate, persistence.WithoutPreload(),
		"audit_scope_id = ? AND control_id = ?",
		req.Msg.GetAuditScopeId(),
		req.Msg.GetControlId(),
	)
	if err == nil {
		return nil, connect.NewError(connect.CodeAlreadyExists, service.ErrResourceAlreadyExists)
	} else if !errors.Is(err, persistence.ErrRecordNotFound) {
		if err = service.HandleDatabaseError(err); err != nil {
			return nil, err
		}
	}

	now := timestamppb.Now()
	cis = &orchestrator.ControlInScope{
		Id:                   uuid.NewString(),
		AuditScopeId:         req.Msg.GetAuditScopeId(),
		TargetOfEvaluationId: req.Msg.GetTargetOfEvaluationId(),
		ControlId:            req.Msg.GetControlId(),
		State:                orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_OPEN,
		AssigneeId:           req.Msg.AssigneeId,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	err = svc.db.Transaction(func(tx persistence.DB) error {
		if err = tx.Create(cis); err != nil {
			return service.HandleDatabaseError(err)
		}
		return createAuditTrailEvent(tx, actorFromContext(ctx), cis.AuditScopeId, cis.Id, req.Msg.GetComment(),
			&orchestrator.ControlScopingEvent{
				ControlInScopeId: cis.Id,
				ControlId:        cis.ControlId,
				AuditScopeId:     cis.AuditScopeId,
				InScope:          true,
			})
	})
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	// Trigger an immediate re-evaluation of the audit scope so the newly
	// in-scope control gets evaluated without waiting for the next interval.
	if svc.scopeChangeCallback != nil {
		go func() {
			_ = svc.scopeChangeCallback(context.Background(), cis.AuditScopeId)
		}()
	}

	res = connect.NewResponse(cis)
	return
}

// GetControlInScope retrieves a ControlInScope record by ID.
func (svc *Service) GetControlInScope(
	ctx context.Context,
	req *connect.Request[orchestrator.GetControlInScopeRequest],
) (res *connect.Response[orchestrator.ControlInScope], err error) {
	var (
		cis     orchestrator.ControlInScope
		allowed bool
	)

	if err = service.Validate(req); err != nil {
		return nil, err
	}

	err = svc.db.Get(&cis, "id = ?", req.Msg.GetId())
	if err = service.HandleDatabaseError(err, service.ErrNotFound("control in scope")); err != nil {
		return nil, err
	}

	allowed, _, err = CheckAccess(ctx, svc.authz, svc,
		orchestrator.RequestType_REQUEST_TYPE_GET,
		cis.AuditScopeId,
		orchestrator.ObjectType_OBJECT_TYPE_CONTROL_IN_SCOPE,
	)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if !allowed {
		return nil, service.ErrPermissionDenied
	}

	res = connect.NewResponse(&cis)
	return
}

// ListControlsInScope lists controls in scope with optional filtering.
func (svc *Service) ListControlsInScope(
	ctx context.Context,
	req *connect.Request[orchestrator.ListControlsInScopeRequest],
) (res *connect.Response[orchestrator.ListControlsInScopeResponse], err error) {
	var (
		records  []*orchestrator.ControlInScope
		conds    []any
		npt      string
		all      bool
		scopeIds []string
	)

	if err = service.Validate(req); err != nil {
		return nil, err
	}

	if req.Msg.OrderBy == "" {
		req.Msg.OrderBy = "created_at"
		req.Msg.Asc = true
	}

	all, scopeIds = svc.authz.AllowedAuditScopes(ctx)
	if !all && len(scopeIds) == 0 {
		return connect.NewResponse(&orchestrator.ListControlsInScopeResponse{
			ControlsInScope: []*orchestrator.ControlInScope{},
		}), nil
	}
	// Collect all conditions as a single WHERE string; GORM only treats the
	// first string argument as a clause and everything after it as arguments,
	// so appending multiple "column = ?" strings silently drops all but the
	// first condition.
	var (
		query []string
		args  []any
	)

	if !all {
		query = append(query, "audit_scope_id IN ?")
		args = append(args, scopeIds)
	}

	if f := req.Msg.GetFilter(); f != nil {
		if f.AuditScopeId != nil {
			query = append(query, "audit_scope_id = ?")
			args = append(args, f.GetAuditScopeId())
		}
		if f.State != nil {
			query = append(query, "state = ?")
			args = append(args, f.GetState())
		}
		if f.AssigneeId != nil {
			query = append(query, "assignee_id = ?")
			args = append(args, f.GetAssigneeId())
		}
		if f.ControlId != nil {
			query = append(query, "control_id = ?")
			args = append(args, f.GetControlId())
		}
	}

	if len(query) > 0 {
		conds = persistence.BuildConds(query, args)
	}

	records, npt, err = service.PaginateStorage[*orchestrator.ControlInScope](req.Msg, svc.db, service.DefaultPaginationOpts, conds...)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	res = connect.NewResponse(&orchestrator.ListControlsInScopeResponse{
		ControlsInScope: records,
		NextPageToken:   npt,
	})
	return
}

// UpdateControlInScope updates the mutable fields of an existing ControlInScope record.
// Only assignee_id and implementation_details can be updated.
func (svc *Service) UpdateControlInScope(
	ctx context.Context,
	req *connect.Request[orchestrator.UpdateControlInScopeRequest],
) (res *connect.Response[orchestrator.ControlInScope], err error) {
	var (
		existing orchestrator.ControlInScope
		allowed  bool
	)

	if err = service.Validate(req); err != nil {
		return nil, err
	}

	err = svc.db.Get(&existing, persistence.WithoutPreload(), "id = ?", req.Msg.GetId())
	if err = service.HandleDatabaseError(err, service.ErrNotFound("control in scope")); err != nil {
		return nil, err
	}

	allowed, _, err = CheckAccess(ctx, svc.authz, svc,
		orchestrator.RequestType_REQUEST_TYPE_UPDATED,
		existing.AuditScopeId,
		orchestrator.ObjectType_OBJECT_TYPE_CONTROL_IN_SCOPE,
	)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if !allowed {
		return nil, service.ErrPermissionDenied
	}

	prevAssigneeId := existing.GetAssigneeId()
	prevDetails := existing.GetImplementationDetails()
	existing.AssigneeId = req.Msg.AssigneeId
	existing.ImplementationDetails = req.Msg.ImplementationDetails
	existing.UpdatedAt = timestamppb.Now()

	err = svc.db.Transaction(func(tx persistence.DB) error {
		if err := tx.Update(&existing, "id = ?", existing.Id); err != nil {
			return service.HandleDatabaseError(err, service.ErrNotFound("control in scope"))
		}
		actor := actorFromContext(ctx)
		if prevAssigneeId != req.Msg.GetAssigneeId() {
			if err := createAuditTrailEvent(tx, actor, existing.AuditScopeId, existing.Id, "",
				&orchestrator.ControlInScopeAssigneeChangedEvent{
					ControlInScopeId:   existing.Id,
					PreviousAssigneeId: &prevAssigneeId,
					NewAssigneeId:      req.Msg.AssigneeId,
				}); err != nil {
				return err
			}
		}
		if prevDetails != req.Msg.GetImplementationDetails() {
			return createAuditTrailEvent(tx, actor, existing.AuditScopeId, existing.Id, "",
				&orchestrator.ControlInScopeDetailsChangedEvent{
					ControlInScopeId: existing.Id,
				})
		}
		return nil
	})
	if err = service.HandleDatabaseError(err, service.ErrNotFound("control in scope")); err != nil {
		return nil, err
	}

	res = connect.NewResponse(&existing)
	return
}

// TransitionControlInScopeState moves a ControlInScope to a new implementation state, enforcing
// the state machine and recording the change as an AuditTrailEvent.
func (svc *Service) TransitionControlInScopeState(
	ctx context.Context,
	req *connect.Request[orchestrator.TransitionControlInScopeStateRequest],
) (res *connect.Response[orchestrator.ControlInScope], err error) {
	var (
		cis     orchestrator.ControlInScope
		allowed bool
	)

	if err = service.Validate(req); err != nil {
		return nil, err
	}

	err = svc.db.Get(&cis, "id = ?", req.Msg.GetId())
	if err = service.HandleDatabaseError(err, service.ErrNotFound("control in scope")); err != nil {
		return nil, err
	}

	allowed, _, err = CheckAccess(ctx, svc.authz, svc,
		orchestrator.RequestType_REQUEST_TYPE_UPDATED,
		cis.AuditScopeId,
		orchestrator.ObjectType_OBJECT_TYPE_CONTROL_IN_SCOPE,
	)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if !allowed {
		return nil, service.ErrPermissionDenied
	}

	toState := req.Msg.GetToState()
	if !isValidTransition(cis.State, toState) {
		return nil, connect.NewError(connect.CodeInvalidArgument,
			fmt.Errorf("invalid state transition: %s → %s",
				cis.State.String(), toState.String()))
	}

	actor := actorFromContext(ctx)
	fromState := cis.State

	err = svc.db.Transaction(func(tx persistence.DB) error {
		cis.State = toState
		cis.UpdatedAt = timestamppb.Now()
		if err = tx.Update(&cis, "id = ?", cis.Id); err != nil {
			return err
		}
		return createAuditTrailEvent(tx, actor, cis.AuditScopeId, cis.Id, req.Msg.Comment,
			&orchestrator.ControlInScopeTransitionEvent{
				ControlInScopeId: cis.Id,
				FromState:        fromState,
				ToState:          toState,
			})
	})
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	res = connect.NewResponse(&cis)
	return
}

// RemoveControlInScope removes a control from scope and records an AuditTrailEvent.
func (svc *Service) RemoveControlInScope(
	ctx context.Context,
	req *connect.Request[orchestrator.RemoveControlInScopeRequest],
) (res *connect.Response[emptypb.Empty], err error) {
	var (
		cis     orchestrator.ControlInScope
		allowed bool
	)

	if err = service.Validate(req); err != nil {
		return nil, err
	}

	err = svc.db.Get(&cis, "id = ?", req.Msg.GetId())
	if err = service.HandleDatabaseError(err, service.ErrNotFound("control in scope")); err != nil {
		return nil, err
	}

	allowed, _, err = CheckAccess(ctx, svc.authz, svc,
		orchestrator.RequestType_REQUEST_TYPE_DELETED,
		cis.AuditScopeId,
		orchestrator.ObjectType_OBJECT_TYPE_CONTROL_IN_SCOPE,
	)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if !allowed {
		return nil, service.ErrPermissionDenied
	}

	actor := actorFromContext(ctx)
	err = svc.db.Transaction(func(tx persistence.DB) error {
		// Cascade: removing a parent control from scope also removes every
		// descendant ControlInScope record so the subtree state stays consistent
		// — a sub-control can't be "in scope" while its parent isn't.
		victims, err := collectControlInScopeSubtree(tx, cis.AuditScopeId, cis.ControlId)
		if err != nil {
			return err
		}
		victims = append([]*orchestrator.ControlInScope{&cis}, victims...)
		for _, victim := range victims {
			// control_in_scope_id is intentionally empty: the ControlInScope record
			// is deleted in this same transaction, so linking to it would create a
			// dangling reference.
			if err := createAuditTrailEvent(tx, actor, victim.AuditScopeId, "", req.Msg.GetComment(),
				&orchestrator.ControlScopingEvent{
					ControlId:    victim.ControlId,
					AuditScopeId: victim.AuditScopeId,
					InScope:      false,
				}); err != nil {
				return err
			}
			if err := tx.Delete(&orchestrator.ControlInScope{}, "id = ?", victim.Id); err != nil {
				return err
			}
			// Remove any existing evaluation results for this control so the UI
			// shows it as "not evaluated" instead of a stale status. Ignore
			// not-found — there may be no results to delete.
			if err := tx.Delete(&evaluation.EvaluationResult{},
				"control_id = ? AND audit_scope_id = ?", victim.ControlId, victim.AuditScopeId,
			); err != nil && !errors.Is(err, persistence.ErrRecordNotFound) {
				return err
			}
		}
		return nil
	})
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	// Trigger an immediate re-evaluation of the audit scope so the remaining
	// controls get re-evaluated without the removed control.
	if svc.scopeChangeCallback != nil {
		go func() {
			_ = svc.scopeChangeCallback(context.Background(), cis.AuditScopeId)
		}()
	}

	res = connect.NewResponse(&emptypb.Empty{})
	return
}

// collectControlInScopeSubtree returns every ControlInScope record under the
// given audit scope whose Control is a descendant of rootControlId, by walking
// the Control.parent_control_id chain breadth-first. The root record itself is
// not included.
func collectControlInScopeSubtree(db persistence.DB, auditScopeId, rootControlId string) ([]*orchestrator.ControlInScope, error) {
	var (
		queue       = []string{rootControlId}
		descendants []*orchestrator.ControlInScope
		visited     = map[string]struct{}{rootControlId: {}}
	)

	for len(queue) > 0 {
		parent := queue[0]
		queue = queue[1:]

		var children []*orchestrator.Control
		if err := db.List(&children, "id", true, 0, -1, "parent_control_id = ?", parent); err != nil {
			return nil, fmt.Errorf("list child controls of %q: %w", parent, err)
		}
		for _, c := range children {
			if _, seen := visited[c.Id]; seen {
				continue
			}
			visited[c.Id] = struct{}{}
			queue = append(queue, c.Id)

			var cis []*orchestrator.ControlInScope
			if err := db.List(&cis, "id", true, 0, -1,
				"audit_scope_id = ? AND control_id = ?", auditScopeId, c.Id); err != nil {
				return nil, fmt.Errorf("list controls in scope for %q: %w", c.Id, err)
			}
			descendants = append(descendants, cis...)
		}
	}

	return descendants, nil
}

// ListAuditTrailEvents lists audit trail events, optionally filtered by audit scope.
func (svc *Service) ListAuditTrailEvents(
	ctx context.Context,
	req *connect.Request[orchestrator.ListAuditTrailEventsRequest],
) (res *connect.Response[orchestrator.ListAuditTrailEventsResponse], err error) {
	var (
		events   []*orchestrator.AuditTrailEvent
		conds    []any
		npt      string
		all      bool
		scopeIds []string
	)

	if err = service.Validate(req); err != nil {
		return nil, err
	}

	if req.Msg.OrderBy == "" {
		req.Msg.OrderBy = "created_at"
		req.Msg.Asc = false
	}

	all, scopeIds = svc.authz.AllowedAuditScopes(ctx)
	if !all && len(scopeIds) == 0 {
		return connect.NewResponse(&orchestrator.ListAuditTrailEventsResponse{
			AuditTrailEvents: []*orchestrator.AuditTrailEvent{},
		}), nil
	}

	// See ListControlsInScope: conditions must be joined into a single WHERE
	// string, otherwise GORM drops all but the first one.
	var (
		query []string
		args  []any
	)

	if !all {
		query = append(query, "audit_scope_id IN ?")
		args = append(args, scopeIds)
	}

	if f := req.Msg.GetFilter(); f != nil {
		if f.AuditScopeId != nil {
			query = append(query, "audit_scope_id = ?")
			args = append(args, f.GetAuditScopeId())
		}
		if f.ControlInScopeId != nil {
			query = append(query, "control_in_scope_id = ?")
			args = append(args, f.GetControlInScopeId())
		}
		if f.ActorId != nil {
			query = append(query, "actor_id = ?")
			args = append(args, f.GetActorId())
		}
	}

	if len(query) > 0 {
		conds = persistence.BuildConds(query, args)
	}

	events, npt, err = service.PaginateStorage[*orchestrator.AuditTrailEvent](req.Msg, svc.db, service.DefaultPaginationOpts, append(conds, persistence.WithoutPreload())...)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	res = connect.NewResponse(&orchestrator.ListAuditTrailEventsResponse{
		AuditTrailEvents: events,
		NextPageToken:    npt,
	})
	return
}
