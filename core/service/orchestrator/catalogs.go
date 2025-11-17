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

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

// CreateCatalog creates a new catalog.
func (svc *service) CreateCatalog(
	ctx context.Context,
	req *connect.Request[orchestrator.CreateCatalogRequest],
) (*connect.Response[orchestrator.Catalog], error) {
	// Persist the new catalog in the database
	err := svc.db.Create(req.Msg.Catalog)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("could not add catalog to the database: %w", err))
	}

	return connect.NewResponse(req.Msg.Catalog), nil
}

// GetCatalog retrieves a catalog by ID.
func (svc *service) GetCatalog(
	ctx context.Context,
	req *connect.Request[orchestrator.GetCatalogRequest],
) (*connect.Response[orchestrator.Catalog], error) {
	var res orchestrator.Catalog

	err := svc.db.Get(&res, "id = ?", req.Msg.CatalogId)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("catalog not found"))
	} else if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	return connect.NewResponse(&res), nil
}

// ListCatalogs lists all catalogs.
func (svc *service) ListCatalogs(
	ctx context.Context,
	req *connect.Request[orchestrator.ListCatalogsRequest],
) (*connect.Response[orchestrator.ListCatalogsResponse], error) {
	var catalogs []*orchestrator.Catalog

	err := svc.db.List(&catalogs, "id", true, 0, -1, nil)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("could not list catalogs: %w", err))
	}

	return connect.NewResponse(&orchestrator.ListCatalogsResponse{
		Catalogs: catalogs,
	}), nil
}

// UpdateCatalog updates an existing catalog.
func (svc *service) UpdateCatalog(
	ctx context.Context,
	req *connect.Request[orchestrator.UpdateCatalogRequest],
) (*connect.Response[orchestrator.Catalog], error) {
	// Check if the catalog exists
	count, err := svc.db.Count(req.Msg.Catalog, "id = ?", req.Msg.Catalog.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	if count == 0 {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("catalog not found"))
	}

	// Save the updated catalog
	err = svc.db.Save(req.Msg.Catalog, "id = ?", req.Msg.Catalog.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	return connect.NewResponse(req.Msg.Catalog), nil
}

// RemoveCatalog removes a catalog by ID.
func (svc *service) RemoveCatalog(
	ctx context.Context,
	req *connect.Request[orchestrator.RemoveCatalogRequest],
) (*connect.Response[emptypb.Empty], error) {
	var catalog orchestrator.Catalog

	// Delete the catalog
	err := svc.db.Delete(&catalog, "id = ?", req.Msg.CatalogId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

// GetCategory retrieves a category by name and catalog ID.
func (svc *service) GetCategory(
	ctx context.Context,
	req *connect.Request[orchestrator.GetCategoryRequest],
) (*connect.Response[orchestrator.Category], error) {
	var res orchestrator.Category

	err := svc.db.Get(&res, "name = ? AND catalog_id = ?", req.Msg.CategoryName, req.Msg.CatalogId)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("category not found"))
	} else if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	return connect.NewResponse(&res), nil
}

// ListControls lists all controls, optionally filtered by catalog ID.
func (svc *service) ListControls(
	ctx context.Context,
	req *connect.Request[orchestrator.ListControlsRequest],
) (*connect.Response[orchestrator.ListControlsResponse], error) {
	var controls []*orchestrator.Control
	var conds []any

	// Filter by catalog_id if provided
	if req.Msg.CatalogId != "" {
		conds = append(conds, "category_catalog_id = ?", req.Msg.CatalogId)
	}

	// Filter by category_name if provided
	if req.Msg.CategoryName != "" {
		conds = append(conds, "category_name = ?", req.Msg.CategoryName)
	}

	err := svc.db.List(&controls, "id", true, 0, -1, conds...)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("could not list controls: %w", err))
	}

	return connect.NewResponse(&orchestrator.ListControlsResponse{
		Controls: controls,
	}), nil
}

// GetControl retrieves a control by ID, category name, and catalog ID.
func (svc *service) GetControl(
	ctx context.Context,
	req *connect.Request[orchestrator.GetControlRequest],
) (*connect.Response[orchestrator.Control], error) {
	var res orchestrator.Control

	err := svc.db.Get(&res, "id = ? AND category_name = ? AND category_catalog_id = ?", 
		req.Msg.ControlId, req.Msg.CategoryName, req.Msg.CatalogId)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("control not found"))
	} else if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	return connect.NewResponse(&res), nil
}
