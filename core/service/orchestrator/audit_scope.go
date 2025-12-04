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

	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/persistence"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"
)

// CreateAuditScope creates a new audit scope.
func (svc *service) CreateAuditScope(
	ctx context.Context,
	req *connect.Request[orchestrator.CreateAuditScopeRequest],
) (*connect.Response[orchestrator.AuditScope], error) {
	// Generate a new UUID for the audit scope if not provided
	if req.Msg.AuditScope.Id == "" {
		req.Msg.AuditScope.Id = uuid.NewString()
	}

	// Persist the new audit scope in the database
	err := svc.db.Create(req.Msg.AuditScope)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("could not add audit scope to the database: %w", err))
	}

	return connect.NewResponse(req.Msg.AuditScope), nil
}

// GetAuditScope retrieves an audit scope by ID.
func (svc *service) GetAuditScope(
	ctx context.Context,
	req *connect.Request[orchestrator.GetAuditScopeRequest],
) (*connect.Response[orchestrator.AuditScope], error) {
	var res orchestrator.AuditScope

	err := svc.db.Get(&res, "id = ?", req.Msg.AuditScopeId)
	if errors.Is(err, persistence.ErrRecordNotFound) {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("audit scope not found"))
	} else if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	return connect.NewResponse(&res), nil
}

// ListAuditScopes lists all audit scopes.
func (svc *service) ListAuditScopes(
	ctx context.Context,
	req *connect.Request[orchestrator.ListAuditScopesRequest],
) (*connect.Response[orchestrator.ListAuditScopesResponse], error) {
	var scopes []*orchestrator.AuditScope
	var conds []any

	// Filter by target_of_evaluation_id if provided
	if req.Msg.Filter != nil && req.Msg.Filter.TargetOfEvaluationId != nil {
		conds = append(conds, "target_of_evaluation_id = ?", *req.Msg.Filter.TargetOfEvaluationId)
	}

	// Filter by catalog_id if provided
	if req.Msg.Filter != nil && req.Msg.Filter.CatalogId != nil {
		conds = append(conds, "catalog_id = ?", *req.Msg.Filter.CatalogId)
	}

	err := svc.db.List(&scopes, "id", true, 0, -1, conds...)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("could not list audit scopes: %w", err))
	}

	return connect.NewResponse(&orchestrator.ListAuditScopesResponse{
		AuditScopes: scopes,
	}), nil
}

// UpdateAuditScope updates an existing audit scope.
func (svc *service) UpdateAuditScope(
	ctx context.Context,
	req *connect.Request[orchestrator.UpdateAuditScopeRequest],
) (*connect.Response[orchestrator.AuditScope], error) {
	// Check if the audit scope exists
	count, err := svc.db.Count(req.Msg.AuditScope, "id = ?", req.Msg.AuditScope.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	if count == 0 {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("audit scope not found"))
	}

	// Save the updated audit scope
	err = svc.db.Save(req.Msg.AuditScope, "id = ?", req.Msg.AuditScope.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	return connect.NewResponse(req.Msg.AuditScope), nil
}

// RemoveAuditScope removes an audit scope by ID.
func (svc *service) RemoveAuditScope(
	ctx context.Context,
	req *connect.Request[orchestrator.RemoveAuditScopeRequest],
) (*connect.Response[emptypb.Empty], error) {
	var scope orchestrator.AuditScope

	// Delete the audit scope
	err := svc.db.Delete(&scope, "id = ?", req.Msg.AuditScopeId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}
