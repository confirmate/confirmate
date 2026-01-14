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
		scope *orchestrator.AuditScope
	)

	// Validate the request, ignoring ID field which will be auto-generated
	if err = service.Validate(req, protovalidate.WithFilter(service.IgnoreIDFilter)); err != nil {
		return nil, err
	}

	scope = req.Msg.AuditScope

	// Generate a new UUID for the audit scope
	scope.Id = uuid.NewString()

	// Persist the new audit scope in the database
	err = svc.db.Create(scope)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	// Notify subscribers
	go svc.publishEvent(&orchestrator.ChangeEvent{
		Timestamp:  timestamppb.Now(),
		Category:   orchestrator.EventCategory_EVENT_CATEGORY_AUDIT_SCOPE,
		RequestType: orchestrator.RequestType_REQUEST_TYPE_CREATED,
		EntityId:   scope.Id,
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
		scope orchestrator.AuditScope
	)

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	err = svc.db.Get(&scope, "id = ?", req.Msg.AuditScopeId)
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
		scopes []*orchestrator.AuditScope
		conds  []any
		npt    string
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

	// Filter by target_of_evaluation_id if provided
	if req.Msg.Filter != nil && req.Msg.Filter.TargetOfEvaluationId != nil {
		conds = append(conds, "target_of_evaluation_id = ?", *req.Msg.Filter.TargetOfEvaluationId)
	}

	// Filter by catalog_id if provided
	if req.Msg.Filter != nil && req.Msg.Filter.CatalogId != nil {
		conds = append(conds, "catalog_id = ?", *req.Msg.Filter.CatalogId)
	}

	scopes, npt, err = service.PaginateStorage[*orchestrator.AuditScope](req.Msg, svc.db, service.DefaultPaginationOpts, conds...)
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
	var scope *orchestrator.AuditScope

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	scope = req.Msg.AuditScope

	// Update the audit scope
	err = svc.db.Update(scope, "id = ?", scope.Id)
	if err = service.HandleDatabaseError(err, service.ErrNotFound("audit scope")); err != nil {
		return nil, err
	}

	// Notify subscribers
	go svc.publishEvent(&orchestrator.ChangeEvent{
		Timestamp:  timestamppb.Now(),
		Category:   orchestrator.EventCategory_EVENT_CATEGORY_AUDIT_SCOPE,
		RequestType: orchestrator.RequestType_REQUEST_TYPE_UPDATED,
		EntityId:   scope.Id,
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
		scope orchestrator.AuditScope
	)

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	// Delete the audit scope
	err = svc.db.Delete(&scope, "id = ?", req.Msg.AuditScopeId)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	// Notify subscribers
	go svc.publishEvent(&orchestrator.ChangeEvent{
		Timestamp:  timestamppb.Now(),
		Category:   orchestrator.EventCategory_EVENT_CATEGORY_AUDIT_SCOPE,
		RequestType: orchestrator.RequestType_REQUEST_TYPE_DELETED,
		EntityId:   req.Msg.AuditScopeId,
	})

	res = connect.NewResponse(&emptypb.Empty{})
	return
}
