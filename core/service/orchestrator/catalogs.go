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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/service"

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

	// Validate the request
	if err = service.Validate(req.Msg); err != nil {
		return nil, err
	}

	// Persist the new catalog in the database
	err = svc.db.Create(catalog)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
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

	// Validate the request
	if err = service.Validate(req.Msg); err != nil {
		return nil, err
	}

	err = svc.db.Get(&catalog, "id = ?", req.Msg.CatalogId)
	if err = service.HandleDatabaseError(err, service.ErrNotFound("catalog")); err != nil {
		return nil, err
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

	// Validate the request
	if err = service.Validate(req.Msg); err != nil {
		return nil, err
	}

	err = svc.db.List(&catalogs, "id", true, 0, -1, nil)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
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

	// Validate the request
	if err = service.Validate(req.Msg); err != nil {
		return nil, err
	}

	// Check if the catalog exists
	count, err = svc.db.Count(catalog, "id = ?", catalog.Id)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	if count == 0 {
		return nil, service.ErrNotFound("catalog")
	}

	// Save the updated catalog
	err = svc.db.Save(catalog, "id = ?", catalog.Id)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
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

	// Validate the request
	if err = service.Validate(req.Msg); err != nil {
		return nil, err
	}

	// Delete the catalog
	err = svc.db.Delete(&catalog, "id = ?", req.Msg.CatalogId)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
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

	// Validate the request
	if err = service.Validate(req.Msg); err != nil {
		return nil, err
	}

	err = svc.db.Get(&category, "name = ? AND catalog_id = ?", req.Msg.CategoryName, req.Msg.CatalogId)
	if err = service.HandleDatabaseError(err, service.ErrNotFound("category")); err != nil {
		return nil, err
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

	// Validate the request
	if err = service.Validate(req.Msg); err != nil {
		return nil, err
	}

	// Filter by catalog_id if provided
	if req.Msg.CatalogId != "" {
		conds = append(conds, "category_catalog_id = ?", req.Msg.CatalogId)
	}

	// Filter by category_name if provided
	if req.Msg.CategoryName != "" {
		conds = append(conds, "category_name = ?", req.Msg.CategoryName)
	}

	err = svc.db.List(&controls, "id", true, 0, -1, conds...)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
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

	// Validate the request
	if err = service.Validate(req.Msg); err != nil {
		return nil, err
	}

	err = svc.db.Get(&control, "id = ? AND category_name = ? AND category_catalog_id = ?",
		req.Msg.ControlId, req.Msg.CategoryName, req.Msg.CatalogId)
	if err = service.HandleDatabaseError(err, service.ErrNotFound("control")); err != nil {
		return nil, err
	}

	res = connect.NewResponse(&control)
	return
}

// loadCatalogs loads catalog definitions from a JSON file.
func (svc *Service) loadCatalogs() (err error) {
	var catalogs []*orchestrator.Catalog

	if svc.catalogsFolder == "" {
		return nil
	}

	// Get all filenames
	files, err := os.ReadDir(svc.catalogsFolder)
	if err != nil {
		return fmt.Errorf("could not read catalogs folder: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		var catalogsFromFile []*orchestrator.Catalog
		b, err := os.ReadFile(filepath.Join(svc.catalogsFolder, file.Name()))
		if err != nil {
			// log error?
			continue
		}

		err = json.Unmarshal(b, &catalogsFromFile)
		if err != nil {
			// log error?
			continue
		}

		catalogs = append(catalogs, catalogsFromFile...)
	}

	// Post-processing (setting parent IDs)
	for _, catalog := range catalogs {
		for _, category := range catalog.Categories {
			for _, control := range category.Controls {
				for _, sub := range control.Controls {
					sub.CategoryName = category.Name
					sub.CategoryCatalogId = catalog.Id

					// Set parent info
					sub.ParentControlCategoryCatalogId = &control.CategoryCatalogId
					sub.ParentControlCategoryName = &control.CategoryName
					sub.ParentControlId = &control.Id
				}
			}
		}
	}

	// Save to DB
	return svc.db.Save(catalogs)
}
