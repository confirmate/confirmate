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

	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/persistence"
	"confirmate.io/core/service"

	"buf.build/go/protovalidate"
	"connectrpc.com/connect"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CreateAuditScope creates a new audit scope.
func (svc *Service) CreateAuditScope(
	ctx context.Context,
	req *connect.Request[orchestrator.CreateAuditScopeRequest],
) (res *connect.Response[orchestrator.AuditScope], err error) {
	var (
		scope   *orchestrator.AuditScope
		allowed bool
	)

	// Validate the request, ignoring ID field which will be auto-generated
	if err = service.Validate(req, protovalidate.WithFilter(service.IgnoreIDFilter)); err != nil {
		return nil, err
	}

	scope = &orchestrator.AuditScope{
		Id:                   uuid.NewString(),
		Name:                 req.Msg.GetAuditScope().GetName(),
		TargetOfEvaluationId: req.Msg.GetAuditScope().GetTargetOfEvaluationId(),
		CatalogId:            req.Msg.GetAuditScope().GetCatalogId(),
		AssuranceLevel:       req.Msg.GetAuditScope().AssuranceLevel,
		Status:               req.Msg.GetAuditScope().GetStatus(),
	}

	// Check access via the configured auth strategy
	allowed, _, err = CheckAccess(ctx, svc.authz, svc, orchestrator.RequestType_REQUEST_TYPE_CREATED, scope.TargetOfEvaluationId, orchestrator.ObjectType_OBJECT_TYPE_AUDIT_SCOPE)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if !allowed {
		return nil, service.ErrPermissionDenied
	}

	// Persist the new audit scope in the database, grant creator admin access, and auto-create
	// ControlInScope records for all controls in the catalog matching the assurance level.
	err = svc.db.Transaction(func(tx persistence.DB) error {
		if err = tx.Create(scope); err != nil {
			return service.HandleDatabaseError(err)
		}

		if err = grantCreatorAdminPermission(ctx, tx, scope.Id, orchestrator.ObjectType_OBJECT_TYPE_AUDIT_SCOPE); err != nil {
			return err
		}

		if err = autoCreateControlsInScope(ctx, tx, scope); err != nil {
			return err
		}

		return nil
	})
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	// Notify subscribers
	go svc.publishEvent(&orchestrator.ChangeEvent{
		Timestamp:   timestamppb.Now(),
		Category:    orchestrator.EventCategory_EVENT_CATEGORY_AUDIT_SCOPE,
		RequestType: orchestrator.RequestType_REQUEST_TYPE_CREATED,
		EntityId:    scope.Id,
		Entity: &orchestrator.ChangeEvent_AuditScope{
			AuditScope: scope,
		},
	})

	res = connect.NewResponse(scope)
	return
}

// GetAuditScope retrieves an audit scope by ID.
func (svc *Service) GetAuditScope(
	ctx context.Context,
	req *connect.Request[orchestrator.GetAuditScopeRequest],
) (res *connect.Response[orchestrator.AuditScope], err error) {
	var (
		scope   orchestrator.AuditScope
		allowed bool
	)

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	// Check access via the configured auth strategy
	allowed, _, err = CheckAccess(ctx, svc.authz, svc, orchestrator.RequestType_REQUEST_TYPE_GET, req.Msg.GetAuditScopeId(), orchestrator.ObjectType_OBJECT_TYPE_AUDIT_SCOPE)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if !allowed {
		return nil, service.ErrPermissionDenied
	}

	err = svc.db.Get(&scope, persistence.WithoutPreload(), "id = ?", req.Msg.AuditScopeId)
	if err = service.HandleDatabaseError(err, service.ErrNotFound("audit scope")); err != nil {
		return nil, err
	}

	res = connect.NewResponse(&scope)
	return
}

// ListAuditScopes lists all audit scopes.
func (svc *Service) ListAuditScopes(
	ctx context.Context,
	req *connect.Request[orchestrator.ListAuditScopesRequest],
) (res *connect.Response[orchestrator.ListAuditScopesResponse], err error) {
	var (
		scopes        []*orchestrator.AuditScope
		conds         []any
		npt           string
		all           bool
		auditScopeIds []string
	)

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	// Set default ordering
	if req.Msg.OrderBy == "" {
		req.Msg.OrderBy = "id"
		req.Msg.Asc = true
	}

	// Use filter from request to build query conditions
	// Filter by target_of_evaluation_id if provided
	if req.Msg.Filter != nil && req.Msg.Filter.TargetOfEvaluationId != nil {
		conds = append(conds, "target_of_evaluation_id = ?", *req.Msg.Filter.TargetOfEvaluationId)
	}
	// Filter by catalog_id if provided
	if req.Msg.Filter != nil && req.Msg.Filter.CatalogId != nil {
		conds = append(conds, "catalog_id = ?", *req.Msg.Filter.CatalogId)
	}

	// Retrieve list of all allowed Audit Scope IDs for the user to filter results by access permissions.
	all, auditScopeIds = svc.authz.AllowedAuditScopes(ctx)
	if !all && len(auditScopeIds) == 0 {
		// User has no access to any Audit Scope, return empty result
		return connect.NewResponse(&orchestrator.ListAuditScopesResponse{
			AuditScopes:   []*orchestrator.AuditScope{},
			NextPageToken: "",
		}), nil
	}

	// If access is not allowed to all objects, add a condition to filter by the allowed object IDs
	if !all {
		conds = append(conds, "id IN ?", auditScopeIds)
	}

	// Query the database with pagination and the constructed conditions
	scopes, npt, err = service.PaginateStorage[*orchestrator.AuditScope](req.Msg, svc.db, service.DefaultPaginationOpts,
		append([]any{persistence.WithoutPreload()}, conds...)...)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	res = connect.NewResponse(&orchestrator.ListAuditScopesResponse{
		AuditScopes:   scopes,
		NextPageToken: npt,
	})
	return
}

// UpdateAuditScope updates an existing audit scope.
func (svc *Service) UpdateAuditScope(
	ctx context.Context,
	req *connect.Request[orchestrator.UpdateAuditScopeRequest],
) (res *connect.Response[orchestrator.AuditScope], err error) {
	var (
		scope   *orchestrator.AuditScope
		allowed bool
	)

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	scope = &orchestrator.AuditScope{
		Id:                   req.Msg.GetAuditScope().GetId(),
		Name:                 req.Msg.GetAuditScope().GetName(),
		TargetOfEvaluationId: req.Msg.GetAuditScope().GetTargetOfEvaluationId(),
		CatalogId:            req.Msg.GetAuditScope().GetCatalogId(),
		AssuranceLevel:       req.Msg.GetAuditScope().AssuranceLevel,
		Status:               req.Msg.GetAuditScope().GetStatus(),
	}

	// Check access via the configured auth strategy
	allowed, _, err = CheckAccess(ctx, svc.authz, svc, orchestrator.RequestType_REQUEST_TYPE_UPDATED, scope.GetId(), orchestrator.ObjectType_OBJECT_TYPE_AUDIT_SCOPE)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if !allowed {
		return nil, service.ErrPermissionDenied
	}

	// Update the audit scope
	err = svc.db.Update(scope, "id = ?", scope.Id)
	if err = service.HandleDatabaseError(err, service.ErrNotFound("audit scope")); err != nil {
		return nil, err
	}

	// Notify subscribers
	go svc.publishEvent(&orchestrator.ChangeEvent{
		Timestamp:   timestamppb.Now(),
		Category:    orchestrator.EventCategory_EVENT_CATEGORY_AUDIT_SCOPE,
		RequestType: orchestrator.RequestType_REQUEST_TYPE_UPDATED,
		EntityId:    scope.Id,
		Entity: &orchestrator.ChangeEvent_AuditScope{
			AuditScope: scope,
		},
	})

	res = connect.NewResponse(scope)
	return
}

// RemoveAuditScope removes an audit scope by ID.
func (svc *Service) RemoveAuditScope(
	ctx context.Context,
	req *connect.Request[orchestrator.RemoveAuditScopeRequest],
) (res *connect.Response[emptypb.Empty], err error) {
	var (
		scope   orchestrator.AuditScope
		allowed bool
	)

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	// Check access via the configured auth strategy
	allowed, _, err = CheckAccess(ctx, svc.authz, svc, orchestrator.RequestType_REQUEST_TYPE_DELETED, req.Msg.AuditScopeId, orchestrator.ObjectType_OBJECT_TYPE_AUDIT_SCOPE)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if !allowed {
		return nil, service.ErrPermissionDenied
	}

	err = svc.db.Get(&scope, persistence.WithoutPreload(), "id = ?", req.Msg.AuditScopeId)
	if err = service.HandleDatabaseError(err, service.ErrNotFound("audit scope")); err != nil {
		return nil, err
	}

	// Delete the audit scope
	err = svc.db.Delete(&scope, "id = ?", req.Msg.AuditScopeId)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	// Notify subscribers
	go svc.publishEvent(&orchestrator.ChangeEvent{
		Timestamp:   timestamppb.Now(),
		Category:    orchestrator.EventCategory_EVENT_CATEGORY_AUDIT_SCOPE,
		RequestType: orchestrator.RequestType_REQUEST_TYPE_DELETED,
		EntityId:    req.Msg.AuditScopeId,
	})

	res = connect.NewResponse(&emptypb.Empty{})
	return
}

// autoCreateControlsInScope loads all controls for the catalog associated with scope and creates
// a ControlInScope record for each matching control. A control matches if the scope has no
// assurance level, the control has no assurance level, or both levels match exactly.
func autoCreateControlsInScope(ctx context.Context, tx persistence.DB, scope *orchestrator.AuditScope) error {
	var controls []*orchestrator.Control

	// Query all controls for the catalog, including sub-controls. Since
	// catalog_id is now set on every control during normalization, a simple
	// filter suffices — no join through category_controls needed.
	if err := tx.Raw(&controls,
		`SELECT * FROM controls WHERE catalog_id = ? ORDER BY controls.short_name`,
		scope.CatalogId); err != nil {
		return service.HandleDatabaseError(err)
	}

	now := timestamppb.Now()
	seen := make(map[string]bool, len(controls))
	for _, ctrl := range controls {
		if seen[ctrl.Id] {
			continue
		}
		seen[ctrl.Id] = true
		// Skip only when both levels are explicitly set and differ. Controls without an
		// assurance level are included in every scope regardless of the scope's level.
		if scope.AssuranceLevel != nil && ctrl.AssuranceLevel != nil &&
			*scope.AssuranceLevel != *ctrl.AssuranceLevel {
			continue
		}
		cis := &orchestrator.ControlInScope{
			Id:                   uuid.NewString(),
			AuditScopeId:         scope.Id,
			TargetOfEvaluationId: scope.TargetOfEvaluationId,
			ControlId:            ctrl.Id,
			State:                orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_OPEN,
			CreatedAt:            now,
			UpdatedAt:            now,
		}
		if err := tx.Create(cis); err != nil {
			return service.HandleDatabaseError(err)
		}
		if err := createAuditTrailEvent(tx, actorFromContext(ctx), cis.AuditScopeId, cis.Id, "",
			&orchestrator.ControlScopingEvent{
				ControlInScopeId: cis.Id,
				ControlId:        cis.ControlId,
				AuditScopeId:     cis.AuditScopeId,
				InScope:          true,
			}); err != nil {
			return err
		}
	}

	return nil
}
