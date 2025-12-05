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

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"
)

// CreateAuditScope creates a new audit scope.
func (svc *Service) CreateAuditScope(
	ctx context.Context,
	req *connect.Request[orchestrator.CreateAuditScopeRequest],
) (res *connect.Response[orchestrator.AuditScope], err error) {
	var (
		scope = req.Msg.AuditScope
	)

	// Generate a new UUID for the audit scope
	scope.Id = uuid.NewString()

	// Persist the new audit scope in the database
	err = svc.db.Create(scope)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

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
	)

	// Filter by target_of_evaluation_id if provided
	if req.Msg.Filter != nil && req.Msg.Filter.TargetOfEvaluationId != nil {
		conds = append(conds, "target_of_evaluation_id = ?", *req.Msg.Filter.TargetOfEvaluationId)
	}

	// Filter by catalog_id if provided
	if req.Msg.Filter != nil && req.Msg.Filter.CatalogId != nil {
		conds = append(conds, "catalog_id = ?", *req.Msg.Filter.CatalogId)
	}

	err = svc.db.List(&scopes, "id", true, 0, -1, conds...)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	res = connect.NewResponse(&orchestrator.ListAuditScopesResponse{
		AuditScopes: scopes,
	})
	return
}

// UpdateAuditScope updates an existing audit scope.
func (svc *Service) UpdateAuditScope(
	ctx context.Context,
	req *connect.Request[orchestrator.UpdateAuditScopeRequest],
) (res *connect.Response[orchestrator.AuditScope], err error) {
	var (
		count int64
		scope = req.Msg.AuditScope
	)

	// Check if the audit scope exists
	count, err = svc.db.Count(scope, "id = ?", scope.Id)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	if count == 0 {
		return nil, service.ErrNotFound("audit scope")
	}

	// Save the updated audit scope
	err = svc.db.Save(scope, "id = ?", scope.Id)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

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

	// Delete the audit scope
	err = svc.db.Delete(&scope, "id = ?", req.Msg.AuditScopeId)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	res = connect.NewResponse(&emptypb.Empty{})
	return
}
