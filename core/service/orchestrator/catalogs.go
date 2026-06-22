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
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/log"
	"confirmate.io/core/persistence"
	"confirmate.io/core/service"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

// CreateCatalog creates a new catalog.
func (svc *Service) CreateCatalog(
	ctx context.Context,
	req *connect.Request[orchestrator.CreateCatalogRequest],
) (res *connect.Response[orchestrator.Catalog], err error) {
	var (
		catalog *orchestrator.Catalog
		allowed bool
	)

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	catalog = &orchestrator.Catalog{
		Id:              req.Msg.GetCatalog().GetId(),
		Name:            req.Msg.GetCatalog().GetName(),
		Categories:      req.Msg.GetCatalog().GetCategories(),
		Description:     req.Msg.Catalog.GetDescription(),
		AllInScope:      req.Msg.Catalog.GetAllInScope(),
		AssuranceLevels: req.Msg.Catalog.GetAssuranceLevels(),
		ShortName:       req.Msg.Catalog.GetShortName(),
		Metadata:        req.Msg.Catalog.Metadata,
	}
	catalog = proto.Clone(catalog).(*orchestrator.Catalog)
	normalizeCatalogControls(catalog)

	// Only admins may grant or revoke permissions.
	allowed, _, err = CheckAccess(ctx, svc.authz, svc, orchestrator.RequestType_REQUEST_TYPE_CREATED, "", orchestrator.ObjectType_OBJECT_TYPE_CATALOG)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if !allowed {
		return nil, service.ErrPermissionDenied
	}
	// Persist the new catalog in the database
	err = svc.db.Create(catalog)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	res = connect.NewResponse(catalog)
	return
}

// GetCatalog retrieves a specific catalog by it's ID. The catalog includes a list of all
// of it categories as well as the first level of controls in each category.
func (svc *Service) GetCatalog(
	ctx context.Context,
	req *connect.Request[orchestrator.GetCatalogRequest],
) (res *connect.Response[orchestrator.Catalog], err error) {
	var (
		catalog orchestrator.Catalog
	)

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	err = svc.db.Get(&catalog,
		// Preload fills in associated entities, in this case controls. We want to only select those controls which do
		// not have a parent, e.g., the top-level
		persistence.WithPreload("Categories.Controls", "parent_control_id IS NULL"),
		"id = ?", req.Msg.CatalogId)
	if err = service.HandleDatabaseError(err, service.ErrNotFound("catalog")); err != nil {
		return nil, err
	}

	res = connect.NewResponse(&catalog)
	return
}

// ListCatalogs lists all security controls catalogs. Each catalog includes a list of its
// categories but no additional sub-resources.
func (svc *Service) ListCatalogs(
	ctx context.Context,
	req *connect.Request[orchestrator.ListCatalogsRequest],
) (res *connect.Response[orchestrator.ListCatalogsResponse], err error) {
	var (
		catalogs []*orchestrator.Catalog
		npt      string
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

	catalogs, npt, err = service.PaginateStorage[*orchestrator.Catalog](req.Msg, svc.db, service.DefaultPaginationOpts)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	res = connect.NewResponse(&orchestrator.ListCatalogsResponse{
		Catalogs:      catalogs,
		NextPageToken: npt,
	})
	return
}

// UpdateCatalog updates an existing catalog.
func (svc *Service) UpdateCatalog(
	ctx context.Context,
	req *connect.Request[orchestrator.UpdateCatalogRequest],
) (res *connect.Response[orchestrator.Catalog], err error) {
	var (
		catalog *orchestrator.Catalog
		allowed bool
	)

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	catalog = &orchestrator.Catalog{
		Id:              req.Msg.GetCatalog().GetId(),
		Name:            req.Msg.GetCatalog().GetName(),
		Categories:      req.Msg.GetCatalog().GetCategories(),
		Description:     req.Msg.Catalog.GetDescription(),
		AllInScope:      req.Msg.Catalog.GetAllInScope(),
		AssuranceLevels: req.Msg.Catalog.GetAssuranceLevels(),
		ShortName:       req.Msg.Catalog.GetShortName(),
		Metadata:        req.Msg.Catalog.Metadata,
	}
	catalog = proto.Clone(catalog).(*orchestrator.Catalog)
	normalizeCatalogControls(catalog)

	// Only admins may grant or revoke permissions.
	allowed, _, err = CheckAccess(ctx, svc.authz, svc, orchestrator.RequestType_REQUEST_TYPE_UPDATED, "", orchestrator.ObjectType_OBJECT_TYPE_CATALOG)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if !allowed {
		return nil, service.ErrPermissionDenied
	}

	// Update the catalog
	err = svc.db.Update(catalog, "id = ?", catalog.Id)
	if err = service.HandleDatabaseError(err, service.ErrNotFound("catalog")); err != nil {
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
		allowed bool
	)

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	// Only admins may grant or revoke permissions.
	allowed, _, err = CheckAccess(ctx, svc.authz, svc, orchestrator.RequestType_REQUEST_TYPE_UPDATED, "", orchestrator.ObjectType_OBJECT_TYPE_CATALOG)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if !allowed {
		return nil, service.ErrPermissionDenied
	}

	// Delete the catalog
	err = svc.db.Delete(&catalog, "id = ?", req.Msg.CatalogId)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	res = connect.NewResponse(&emptypb.Empty{})
	return
}

// GetCategory retrieves a category of a catalog specified by the catalog ID and the
// category name. It includes the first level of controls within each
// category.
func (svc *Service) GetCategory(
	ctx context.Context,
	req *connect.Request[orchestrator.GetCategoryRequest],
) (res *connect.Response[orchestrator.Category], err error) {
	var (
		category orchestrator.Category
	)

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	err = svc.db.Get(&category,
		// Preload fills in associated entities, in this case controls. We want to only select those controls which do
		// not have a parent, e.g., the top-level
		persistence.WithPreload("Controls", "parent_control_id IS NULL"),
		"name = ? AND catalog_id = ?", req.Msg.CategoryName, req.Msg.CatalogId)
	if err = service.HandleDatabaseError(err, service.ErrNotFound("category")); err != nil {
		return nil, err
	}

	res = connect.NewResponse(&category)
	return
}

// ListControls lists all controls.
func (svc *Service) ListControls(
	ctx context.Context,
	req *connect.Request[orchestrator.ListControlsRequest],
) (res *connect.Response[orchestrator.ListControlsResponse], err error) {
	var (
		controls     []*orchestrator.Control
		npt          string
		conds        []any
		whereClauses []string
		args         []any
		where        string
	)

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	// Set default ordering
	if req.Msg.OrderBy == "" {
		req.Msg.OrderBy = "short_name"
		req.Msg.Asc = true
	}

	// Apply filters if provided
	if req.Msg.Filter != nil {
		if req.Msg.Filter.CatalogId != nil {
			whereClauses = append(whereClauses, "catalog_id = ?")
			args = append(args, req.Msg.Filter.GetCatalogId())
		}
		if req.Msg.Filter.CategoryName != nil {
			whereClauses = append(whereClauses, "category_name = ?")
			args = append(args, req.Msg.Filter.GetCategoryName())
		}
		if req.Msg.Filter.AssuranceLevels != nil {
			whereClauses = append(whereClauses, "assurace_levels = ?")
			args = append(args, req.Msg.Filter.GetAssuranceLevels())
		}
	}

	// Combine all WHERE clauses with AND
	if len(whereClauses) > 0 {
		where = strings.Join(whereClauses, " AND ")
		conds = append(conds, where)
		conds = append(conds, args...)
	}

	// Prepare preloads and combine with any WHERE clauses/args.
	opts := []any{
		persistence.WithPreload("Controls.Metrics"),
		persistence.WithPreload("Metrics"),
	}
	if where != "" {
		// first the SQL WHERE clause, followed by its arguments
		opts = append(opts, where)
		opts = append(opts, args...)
	}

	// Preload Metrics and sub-Controls (with their own Metrics) so callers —
	// notably the evaluation service — can walk a control's full subtree and
	// resolve every metric ID without making per-control round trips.
	controls, npt, err = service.PaginateStorage[*orchestrator.Control](
		req.Msg, svc.db, service.DefaultPaginationOpts,
		opts...,
	)
	if err = service.HandleDatabaseError(err); err != nil {
		return nil, err
	}

	res = connect.NewResponse(&orchestrator.ListControlsResponse{
		Controls:      controls,
		NextPageToken: npt,
	})
	return
}

// GetControl retrieves a control by its unique control ID. If present, it also includes a list of
// sub-controls if present or a list of metrics if no sub-controls but metrics
// are present.
func (svc *Service) GetControl(
	ctx context.Context,
	req *connect.Request[orchestrator.GetControlRequest],
) (res *connect.Response[orchestrator.Control], err error) {
	var (
		control orchestrator.Control
	)

	// Validate the request
	if err = service.Validate(req); err != nil {
		return nil, err
	}

	err = svc.db.Get(&control, persistence.WithPreload("Controls.Metrics"), "id = ?", req.Msg.ControlId)
	if err = service.HandleDatabaseError(err, service.ErrNotFound("control")); err != nil {
		return nil, err
	}

	res = connect.NewResponse(&control)
	return
}

// loadCatalogs loads catalog definitions from configured sources.
// It loads catalogs from:
// 1. DefaultCatalogsPath (if LoadDefaultCatalogs is true)
// 2. LoadCatalogsFunc (if provided) for additional custom catalogs
func (svc *Service) loadCatalogs() (err error) {
	var catalogs []*orchestrator.Catalog

	// Load default catalogs from folder if enabled
	if svc.cfg.LoadDefaultCatalogs {
		defaultCatalogs, err := svc.loadCatalogsFromFolder(svc.cfg.DefaultCatalogsPath)
		if err != nil {
			return fmt.Errorf("could not load default catalogs: %w", err)
		}
		catalogs = append(catalogs, defaultCatalogs...)
	}

	// Load additional catalogs from custom function if provided
	if svc.cfg.LoadCatalogsFunc != nil {
		additionalCatalogs, err := svc.cfg.LoadCatalogsFunc(svc)
		if err != nil {
			return fmt.Errorf("could not load additional catalogs: %w", err)
		}
		catalogs = append(catalogs, additionalCatalogs...)
	}

	// Save all catalogs to DB (only if we have any)
	if len(catalogs) > 0 {
		return svc.db.Save(catalogs)
	}

	return nil
}

// loadCatalogsFromFolder loads catalogs from a specified folder.
func (svc *Service) loadCatalogsFromFolder(folder string) (catalogs []*orchestrator.Catalog, err error) {
	if folder == "" {
		return nil, nil
	}

	// Get all filenames
	files, err := os.ReadDir(folder)
	if err != nil {
		return nil, fmt.Errorf("could not read catalogs folder: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		var catalogsFromFile []*orchestrator.Catalog
		b, err := os.ReadFile(filepath.Join(folder, file.Name()))
		if err != nil {
			slog.Warn("Failed to read catalog file, skipping", "file", file.Name(), log.Err(err))
			continue
		}

		err = json.Unmarshal(b, &catalogsFromFile)
		if err != nil {
			slog.Warn("Failed to unmarshal catalog file, skipping", "file", file.Name(), log.Err(err))
			continue
		}

		catalogs = append(catalogs, catalogsFromFile...)
	}

	for _, catalog := range catalogs {
		normalizeCatalogControls(catalog)
	}

	return catalogs, nil
}

func normalizeCatalogControls(catalog *orchestrator.Catalog) {
	if catalog == nil {
		return
	}

	for _, category := range catalog.Categories {
		normalizeControls(category.GetControls(), nil)
		category.Controls = flattenControls(category.GetControls())
	}
}

func normalizeControls(controls []*orchestrator.Control, parent *orchestrator.Control) {
	for _, control := range controls {
		if control.GetShortName() == "" {
			control.ShortName = control.GetId()
		}
		if _, err := uuid.Parse(control.GetId()); err != nil {
			control.Id = uuid.NewString()
		}

		if parent != nil {
			control.ParentControlId = &parent.Id
		} else {
			control.ParentControlId = nil
		}

		normalizeControls(control.GetControls(), control)
	}
}

func flattenControls(controls []*orchestrator.Control) []*orchestrator.Control {
	var (
		flat    []*orchestrator.Control
		visited = make(map[string]struct{})
	)

	var walk func(items []*orchestrator.Control)
	walk = func(items []*orchestrator.Control) {
		for _, control := range items {
			if _, ok := visited[control.GetId()]; !ok {
				visited[control.GetId()] = struct{}{}
				flat = append(flat, control)
			}
			walk(control.GetControls())
		}
	}

	walk(controls)

	return flat
}
