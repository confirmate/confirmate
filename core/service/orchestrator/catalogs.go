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
	"google.golang.org/protobuf/types/known/emptypb"
)

// CreateCatalog creates a new catalog.
func (svc *Service) CreateCatalog(
	ctx context.Context,
	req *connect.Request[orchestrator.CreateCatalogRequest],
) (res *connect.Response[orchestrator.Catalog], err error) {
	var (
		catalog = req.Msg.Catalog
	)

	// Persist the new catalog in the database
	err = svc.db.Create(catalog)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("could not add catalog to the database: %w", err))
	}

	res = connect.NewResponse(catalog)
	return
}

// GetCatalog retrieves a catalog by ID.
func (svc *Service) GetCatalog(
	ctx context.Context,
	req *connect.Request[orchestrator.GetCatalogRequest],
) (res *connect.Response[orchestrator.Catalog], err error) {
	var (
		catalog orchestrator.Catalog
	)

	err = svc.db.Get(&catalog, "id = ?", req.Msg.CatalogId)
	if errors.Is(err, persistence.ErrRecordNotFound) {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("catalog not found"))
	} else if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	res = connect.NewResponse(&catalog)
	return
}

// ListCatalogs lists all catalogs.
func (svc *Service) ListCatalogs(
	ctx context.Context,
	req *connect.Request[orchestrator.ListCatalogsRequest],
) (res *connect.Response[orchestrator.ListCatalogsResponse], err error) {
	var (
		catalogs []*orchestrator.Catalog
	)

	err = svc.db.List(&catalogs, "id", true, 0, -1, nil)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("could not list catalogs: %w", err))
	}

	res = connect.NewResponse(&orchestrator.ListCatalogsResponse{
		Catalogs: catalogs,
	})
	return
}

// UpdateCatalog updates an existing catalog.
func (svc *Service) UpdateCatalog(
	ctx context.Context,
	req *connect.Request[orchestrator.UpdateCatalogRequest],
) (res *connect.Response[orchestrator.Catalog], err error) {
	var (
		count   int64
		catalog = req.Msg.Catalog
	)

	// Check if the catalog exists
	count, err = svc.db.Count(catalog, "id = ?", catalog.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	if count == 0 {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("catalog not found"))
	}

	// Save the updated catalog
	err = svc.db.Save(catalog, "id = ?", catalog.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	res = connect.NewResponse(catalog)
	return
}

// RemoveCatalog removes a catalog by ID.
func (svc *Service) RemoveCatalog(
	ctx context.Context,
	req *connect.Request[orchestrator.RemoveCatalogRequest],
) (res *connect.Response[emptypb.Empty], err error) {
	var (
		catalog orchestrator.Catalog
	)

	// Delete the catalog
	err = svc.db.Delete(&catalog, "id = ?", req.Msg.CatalogId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	res = connect.NewResponse(&emptypb.Empty{})
	return
}

// GetCategory retrieves a category by name and catalog ID.
func (svc *Service) GetCategory(
	ctx context.Context,
	req *connect.Request[orchestrator.GetCategoryRequest],
) (res *connect.Response[orchestrator.Category], err error) {
	var (
		category orchestrator.Category
	)

	err = svc.db.Get(&category, "name = ? AND catalog_id = ?", req.Msg.CategoryName, req.Msg.CatalogId)
	if errors.Is(err, persistence.ErrRecordNotFound) {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("category not found"))
	} else if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	res = connect.NewResponse(&category)
	return
}

// ListControls lists all controls, optionally filtered by catalog ID.
func (svc *Service) ListControls(
	ctx context.Context,
	req *connect.Request[orchestrator.ListControlsRequest],
) (res *connect.Response[orchestrator.ListControlsResponse], err error) {
	var (
		controls []*orchestrator.Control
		conds    []any
	)

	// Filter by catalog_id if provided
	if req.Msg.CatalogId != "" {
		conds = append(conds, "category_catalog_id = ?", req.Msg.CatalogId)
	}

	// Filter by category_name if provided
	if req.Msg.CategoryName != "" {
		conds = append(conds, "category_name = ?", req.Msg.CategoryName)
	}

	err = svc.db.List(&controls, "id", true, 0, -1, conds...)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("could not list controls: %w", err))
	}

	res = connect.NewResponse(&orchestrator.ListControlsResponse{
		Controls: controls,
	})
	return
}

// GetControl retrieves a control by ID, category name, and catalog ID.
func (svc *Service) GetControl(
	ctx context.Context,
	req *connect.Request[orchestrator.GetControlRequest],
) (res *connect.Response[orchestrator.Control], err error) {
	var (
		control orchestrator.Control
	)

	err = svc.db.Get(&control, "id = ? AND category_name = ? AND category_catalog_id = ?",
		req.Msg.ControlId, req.Msg.CategoryName, req.Msg.CatalogId)
	if errors.Is(err, persistence.ErrRecordNotFound) {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("control not found"))
	} else if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
	}

	res = connect.NewResponse(&control)
	return
}
