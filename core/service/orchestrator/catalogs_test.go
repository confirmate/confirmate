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
	"errors"
	"os"
	"path/filepath"
	"testing"

	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/persistence"
	"confirmate.io/core/persistence/persistencetest"
	"confirmate.io/core/service/orchestrator/orchestratortest"
	"confirmate.io/core/util/assert"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestService_CreateCatalog(t *testing.T) {
	type args struct {
		req *orchestrator.CreateCatalogRequest
	}
	type fields struct {
		db *persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[orchestrator.Catalog]]
		wantErr assert.WantErr
	}{
		{
			name: "happy path",
			args: args{
				req: &orchestrator.CreateCatalogRequest{
					Catalog: orchestratortest.MockCatalog1,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.Catalog], args ...any) bool {
				assert.NotNil(t, got.Msg)
				return assert.Equal(t, orchestratortest.MockCatalog1.Id, got.Msg.Id)
			},
			wantErr: assert.NoError,
		},
		{
			name: "validation error - empty request",
			args: args{
				req: &orchestrator.CreateCatalogRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.Catalog]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument)
			},
		},
		{
			name: "validation error - missing catalog",
			args: args{
				req: &orchestrator.CreateCatalogRequest{
					Catalog: &orchestrator.Catalog{},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.Catalog]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.IsValidationError(t, err, "catalog.name")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: tt.fields.db,
			}
			res, err := svc.CreateCatalog(context.Background(), connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}

func TestService_GetCatalog(t *testing.T) {
	type args struct {
		req *orchestrator.GetCatalogRequest
	}
	type fields struct {
		db *persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[orchestrator.Catalog]]
		wantErr assert.WantErr
	}{
		{
			name: "happy path",
			args: args{
				req: &orchestrator.GetCatalogRequest{
					CatalogId: orchestratortest.MockCatalog1.Id,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d *persistence.DB) {
					err := d.Create(orchestratortest.MockCatalog1)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.Catalog], args ...any) bool {
				assert.NotNil(t, got.Msg)
				return assert.Equal(t, orchestratortest.MockCatalog1.Id, got.Msg.Id)
			},
			wantErr: assert.NoError,
		},
		{
			name: "validation error - empty request",
			args: args{
				req: &orchestrator.GetCatalogRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.Catalog]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument)
			},
		},
		{
			name: "not found",
			args: args{
				req: &orchestrator.GetCatalogRequest{
					CatalogId: "non-existent",
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.Catalog]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: tt.fields.db,
			}
			res, err := svc.GetCatalog(context.Background(), connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}

func TestService_ListCatalogs(t *testing.T) {
	type args struct {
		req *orchestrator.ListCatalogsRequest
	}
	type fields struct {
		db *persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[orchestrator.ListCatalogsResponse]]
		wantErr assert.WantErr
	}{
		{
			name: "list all",
			args: args{
				req: &orchestrator.ListCatalogsRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d *persistence.DB) {
					err := d.Create(orchestratortest.MockCatalog1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockCatalog2)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ListCatalogsResponse], args ...any) bool {
				assert.NotNil(t, got.Msg)
				return assert.Equal(t, 2, len(got.Msg.Catalogs))
			},
			wantErr: assert.NoError,
		},
		{
			name: "empty list",
			args: args{
				req: &orchestrator.ListCatalogsRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ListCatalogsResponse], args ...any) bool {
				assert.NotNil(t, got.Msg)
				return assert.Equal(t, 0, len(got.Msg.Catalogs))
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: tt.fields.db,
			}
			res, err := svc.ListCatalogs(context.Background(), connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}

func TestService_UpdateCatalog(t *testing.T) {
	type args struct {
		req *orchestrator.UpdateCatalogRequest
	}
	type fields struct {
		db *persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[orchestrator.Catalog]]
		wantErr assert.WantErr
	}{
		{
			name: "happy path",
			args: args{
				req: &orchestrator.UpdateCatalogRequest{
					Catalog: &orchestrator.Catalog{
						Id:          orchestratortest.MockCatalog1.Id,
						Name:        "Updated Catalog",
						Description: "Updated description",
					},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d *persistence.DB) {
					err := d.Create(orchestratortest.MockCatalog1)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.Catalog], args ...any) bool {
				assert.NotNil(t, got.Msg)
				return assert.Equal(t, "Updated Catalog", got.Msg.Name)
			},
			wantErr: assert.NoError,
		},
		{
			name: "validation error - empty request",
			args: args{
				req: &orchestrator.UpdateCatalogRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.Catalog]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument)
			},
		},
		{
			name: "validation error - missing id",
			args: args{
				req: &orchestrator.UpdateCatalogRequest{
					Catalog: &orchestrator.Catalog{
						Name: "Updated Catalog",
					},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.Catalog]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.IsValidationError(t, err, "catalog.id")
			},
		},
		{
			name: "not found",
			args: args{
				req: &orchestrator.UpdateCatalogRequest{
					Catalog: &orchestrator.Catalog{
						Id:          "non-existent",
						Name:        "Updated Catalog",
						Description: "Updated description",
					},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.Catalog]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: tt.fields.db,
			}
			res, err := svc.UpdateCatalog(context.Background(), connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}

func TestService_RemoveCatalog(t *testing.T) {
	type args struct {
		req *orchestrator.RemoveCatalogRequest
	}
	type fields struct {
		db *persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[emptypb.Empty]]
		wantErr assert.WantErr
	}{
		{
			name: "happy path",
			args: args{
				req: &orchestrator.RemoveCatalogRequest{
					CatalogId: orchestratortest.MockCatalog1.Id,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d *persistence.DB) {
					err := d.Create(orchestratortest.MockCatalog1)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *connect.Response[emptypb.Empty], args ...any) bool {
				return assert.NotNil(t, got.Msg)
			},
			wantErr: assert.NoError,
		},
		{
			name: "validation error - empty request",
			args: args{
				req: &orchestrator.RemoveCatalogRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[emptypb.Empty]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: tt.fields.db,
			}
			res, err := svc.RemoveCatalog(context.Background(), connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}

func TestService_GetCategory(t *testing.T) {
	type args struct {
		req *orchestrator.GetCategoryRequest
	}
	type fields struct {
		db *persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[orchestrator.Category]]
		wantErr assert.WantErr
	}{
		{
			name: "happy path",
			args: args{
				req: &orchestrator.GetCategoryRequest{
					CatalogId:    orchestratortest.MockCategory1.CatalogId,
					CategoryName: orchestratortest.MockCategory1.Name,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d *persistence.DB) {
					err := d.Create(orchestratortest.MockCatalog1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockCategory1)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.Category], args ...any) bool {
				assert.NotNil(t, got.Msg)
				return assert.Equal(t, orchestratortest.MockCategory1.Name, got.Msg.Name)
			},
			wantErr: assert.NoError,
		},
		{
			name: "not found",
			args: args{
				req: &orchestrator.GetCategoryRequest{
					CatalogId:    orchestratortest.MockCategory1.CatalogId,
					CategoryName: "non-existent",
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.Category]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: tt.fields.db,
			}
			res, err := svc.GetCategory(context.Background(), connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}

func TestService_ListControls(t *testing.T) {
	type args struct {
		req *orchestrator.ListControlsRequest
	}
	type fields struct {
		db *persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[orchestrator.ListControlsResponse]]
		wantErr assert.WantErr
	}{
		{
			name: "list all",
			args: args{
				req: &orchestrator.ListControlsRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d *persistence.DB) {
					err := d.Create(orchestratortest.MockCatalog1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockCatalog2)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockCategory1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockCategory2)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockControl1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockControl2)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ListControlsResponse], args ...any) bool {
				assert.NotNil(t, got.Msg)
				return assert.Equal(t, 2, len(got.Msg.Controls))
			},
			wantErr: assert.NoError,
		},
		{
			name: "filter by catalog",
			args: args{
				req: &orchestrator.ListControlsRequest{
					CatalogId: orchestratortest.MockControl1.CategoryCatalogId,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d *persistence.DB) {
					err := d.Create(orchestratortest.MockCatalog1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockCatalog2)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockCategory1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockCategory2)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockControl1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockControl2)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ListControlsResponse], args ...any) bool {
				assert.NotNil(t, got.Msg)
				return assert.Equal(t, 1, len(got.Msg.Controls))
			},
			wantErr: assert.NoError,
		},
		{
			name: "filter by category",
			args: args{
				req: &orchestrator.ListControlsRequest{
					CategoryName: orchestratortest.MockControl1.CategoryName,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d *persistence.DB) {
					err := d.Create(orchestratortest.MockCatalog1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockCategory1)
					assert.NoError(t, err)
					err = d.Create(&orchestrator.Category{
						Name:      "category-2",
						CatalogId: orchestratortest.MockCatalog1.Id,
					})
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockControl1)
					assert.NoError(t, err)
					err = d.Create(&orchestrator.Control{
						Id:                "control-3",
						CategoryName:      "category-2",
						CategoryCatalogId: orchestratortest.MockControl1.CategoryCatalogId,
					})
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ListControlsResponse], args ...any) bool {
				assert.NotNil(t, got.Msg)
				return assert.Equal(t, 1, len(got.Msg.Controls))
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: tt.fields.db,
			}
			res, err := svc.ListControls(context.Background(), connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}

func TestService_GetControl(t *testing.T) {
	type args struct {
		req *orchestrator.GetControlRequest
	}
	type fields struct {
		db *persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[orchestrator.Control]]
		wantErr assert.WantErr
	}{
		{
			name: "happy path",
			args: args{
				req: &orchestrator.GetControlRequest{
					ControlId:    orchestratortest.MockControl1.Id,
					CategoryName: orchestratortest.MockControl1.CategoryName,
					CatalogId:    orchestratortest.MockControl1.CategoryCatalogId,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d *persistence.DB) {
					err := d.Create(orchestratortest.MockCatalog1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockCategory1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockControl1)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.Control], args ...any) bool {
				assert.NotNil(t, got.Msg)
				return assert.Equal(t, orchestratortest.MockControl1.Id, got.Msg.Id)
			},
			wantErr: assert.NoError,
		},
		{
			name: "not found",
			args: args{
				req: &orchestrator.GetControlRequest{
					ControlId:    "non-existent",
					CategoryName: orchestratortest.MockControl1.CategoryName,
					CatalogId:    orchestratortest.MockControl1.CategoryCatalogId,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.Control]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: tt.fields.db,
			}
			res, err := svc.GetControl(context.Background(), connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}

func TestService_loadCatalogs(t *testing.T) {
	tests := []struct {
		name             string
		loadDefaultCats  bool
		catalogsPath     string
		loadCatalogsFunc func(*Service) ([]*orchestrator.Catalog, error)
		setupFiles       func(t *testing.T, dir string)
		wantErr          assert.WantErr
		wantDB           assert.Want[*persistence.DB]
	}{
		{
			name:            "load from default folder with valid catalogs",
			loadDefaultCats: true,
			catalogsPath:    "",
			setupFiles: func(t *testing.T, dir string) {
				catalog := []*orchestrator.Catalog{
					{
						Id:          "test-catalog-1",
						Name:        "Test Catalog 1",
						Description: "Test description",
					},
				}
				data, err := json.Marshal(catalog)
				assert.NoError(t, err)
				err = os.WriteFile(filepath.Join(dir, "catalog1.json"), data, 0644)
				assert.NoError(t, err)
			},
			wantErr: assert.NoError,
			wantDB: func(t *testing.T, db *persistence.DB, args ...any) bool {
				catalog := assert.InDB[orchestrator.Catalog](t, db, "test-catalog-1")
				return assert.Equal(t, "Test Catalog 1", catalog.Name)
			},
		},
		{
			name:            "load from custom function",
			loadDefaultCats: false,
			loadCatalogsFunc: func(svc *Service) ([]*orchestrator.Catalog, error) {
				return []*orchestrator.Catalog{
					orchestratortest.MockCatalog1,
					orchestratortest.MockCatalog2,
				}, nil
			},
			wantErr: assert.NoError,
			wantDB: func(t *testing.T, db *persistence.DB, args ...any) bool {
				catalog1 := assert.InDB[orchestrator.Catalog](t, db, orchestratortest.MockCatalog1.Id)
				catalog2 := assert.InDB[orchestrator.Catalog](t, db, orchestratortest.MockCatalog2.Id)
				return assert.NotNil(t, catalog1) &&
					assert.NotNil(t, catalog2)
			},
		},
		{
			name:            "load from both default folder and custom function",
			loadDefaultCats: true,
			setupFiles: func(t *testing.T, dir string) {
				catalog := []*orchestrator.Catalog{
					{
						Id:          "folder-catalog",
						Name:        "Folder Catalog",
						Description: "From folder",
					},
				}
				data, err := json.Marshal(catalog)
				assert.NoError(t, err)
				err = os.WriteFile(filepath.Join(dir, "catalog.json"), data, 0644)
				assert.NoError(t, err)
			},
			loadCatalogsFunc: func(svc *Service) ([]*orchestrator.Catalog, error) {
				return []*orchestrator.Catalog{
					orchestratortest.MockCatalog1,
				}, nil
			},
			wantErr: assert.NoError,
			wantDB: func(t *testing.T, db *persistence.DB, args ...any) bool {
				folderCatalog := assert.InDB[orchestrator.Catalog](t, db, "folder-catalog")
				customCatalog := assert.InDB[orchestrator.Catalog](t, db, orchestratortest.MockCatalog1.Id)
				return assert.NotNil(t, folderCatalog) &&
					assert.NotNil(t, customCatalog)
			},
		},
		{
			name:            "empty folder and no custom function",
			loadDefaultCats: true,
			wantErr:         assert.NoError,
			wantDB:          assert.NotNil[*persistence.DB],
		},
		{
			name:            "custom function returns error",
			loadDefaultCats: false,
			loadCatalogsFunc: func(svc *Service) ([]*orchestrator.Catalog, error) {
				return nil, errors.New("custom error")
			},
			wantErr: func(t *testing.T, err error, args ...any) bool {
				return assert.ErrorContains(t, err, "custom error")
			},
			wantDB: assert.NotNil[*persistence.DB],
		},
		{
			name:            "invalid catalogs path",
			loadDefaultCats: true,
			catalogsPath:    "/nonexistent/path",
			wantErr: func(t *testing.T, err error, args ...any) bool {
				return assert.ErrorContains(t, err, "could not load default catalogs")
			},
			wantDB: assert.NotNil[*persistence.DB],
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			if tt.setupFiles != nil {
				tt.setupFiles(t, tempDir)
			}

			catalogsPath := tt.catalogsPath
			if catalogsPath == "" && tt.loadDefaultCats {
				catalogsPath = tempDir
			}

			svc := &Service{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
				cfg: Config{
					LoadDefaultCatalogs: tt.loadDefaultCats,
					DefaultCatalogsPath: catalogsPath,
					LoadCatalogsFunc:    tt.loadCatalogsFunc,
				},
			}

			err := svc.loadCatalogs()
			tt.wantErr(t, err)
			tt.wantDB(t, svc.db)
		})
	}
}

func TestService_loadCatalogsFromFolder(t *testing.T) {
	tests := []struct {
		name       string
		setupFiles func(t *testing.T, dir string)
		folder     string
		wantCount  int
		wantErr    assert.WantErr
	}{
		{
			name: "load valid catalog file",
			setupFiles: func(t *testing.T, dir string) {
				catalog := []*orchestrator.Catalog{
					{
						Id:          "catalog-1",
						Name:        "Catalog 1",
						Description: "Test catalog",
					},
				}
				data, err := json.Marshal(catalog)
				assert.NoError(t, err)
				err = os.WriteFile(filepath.Join(dir, "catalog1.json"), data, 0644)
				assert.NoError(t, err)
			},
			wantCount: 1,
			wantErr:   assert.NoError,
		},
		{
			name: "load multiple catalog files",
			setupFiles: func(t *testing.T, dir string) {
				catalog1 := []*orchestrator.Catalog{
					{
						Id:          "catalog-1",
						Name:        "Catalog 1",
						Description: "First catalog",
					},
				}
				catalog2 := []*orchestrator.Catalog{
					{
						Id:          "catalog-2",
						Name:        "Catalog 2",
						Description: "Second catalog",
					},
				}
				data1, _ := json.Marshal(catalog1)
				data2, _ := json.Marshal(catalog2)
				os.WriteFile(filepath.Join(dir, "catalog1.json"), data1, 0644)
				os.WriteFile(filepath.Join(dir, "catalog2.json"), data2, 0644)
			},
			wantCount: 2,
			wantErr:   assert.NoError,
		},
		{
			name: "skip non-json files",
			setupFiles: func(t *testing.T, dir string) {
				catalog := []*orchestrator.Catalog{
					{
						Id:   "catalog-1",
						Name: "Catalog 1",
					},
				}
				data, _ := json.Marshal(catalog)
				os.WriteFile(filepath.Join(dir, "catalog.json"), data, 0644)
				os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("test"), 0644)
				os.WriteFile(filepath.Join(dir, "config.yaml"), []byte("test"), 0644)
			},
			wantCount: 1,
			wantErr:   assert.NoError,
		},
		{
			name: "skip directories",
			setupFiles: func(t *testing.T, dir string) {
				catalog := []*orchestrator.Catalog{
					{
						Id:   "catalog-1",
						Name: "Catalog 1",
					},
				}
				data, _ := json.Marshal(catalog)
				os.WriteFile(filepath.Join(dir, "catalog.json"), data, 0644)
				os.Mkdir(filepath.Join(dir, "subdir"), 0755)
			},
			wantCount: 1,
			wantErr:   assert.NoError,
		},
		{
			name: "skip invalid json files",
			setupFiles: func(t *testing.T, dir string) {
				catalog := []*orchestrator.Catalog{
					{
						Id:   "catalog-1",
						Name: "Catalog 1",
					},
				}
				data, _ := json.Marshal(catalog)
				os.WriteFile(filepath.Join(dir, "valid.json"), data, 0644)
				os.WriteFile(filepath.Join(dir, "invalid.json"), []byte("not json"), 0644)
			},
			wantCount: 1,
			wantErr:   assert.NoError,
		},
		{
			name: "process nested controls and set parent relationships",
			setupFiles: func(t *testing.T, dir string) {
				catalog := []*orchestrator.Catalog{
					{
						Id:   "catalog-1",
						Name: "Catalog 1",
						Categories: []*orchestrator.Category{
							{
								Name:      "category-1",
								CatalogId: "catalog-1",
								Controls: []*orchestrator.Control{
									{
										Id:                "control-1",
										CategoryName:      "category-1",
										CategoryCatalogId: "catalog-1",
										Controls: []*orchestrator.Control{
											{
												Id: "sub-control-1",
											},
										},
									},
								},
							},
						},
					},
				}
				data, _ := json.Marshal(catalog)
				os.WriteFile(filepath.Join(dir, "catalog.json"), data, 0644)
			},
			wantCount: 1,
			wantErr:   assert.NoError,
		},
		{
			name:      "empty folder",
			wantCount: 0,
			wantErr:   assert.NoError,
		},
		{
			name:      "empty folder path",
			folder:    "",
			wantCount: 0,
			wantErr:   assert.NoError,
		},
		{
			name:      "nonexistent folder",
			folder:    "/nonexistent/path",
			wantCount: 0,
			wantErr: func(t *testing.T, err error, args ...any) bool {
				return assert.ErrorContains(t, err, "could not read catalogs folder")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			if tt.setupFiles != nil {
				tt.setupFiles(t, tempDir)
			}

			folder := tt.folder
			if folder == "" && tt.name != "empty folder path" {
				folder = tempDir
			}

			svc := &Service{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			}

			catalogs, err := svc.loadCatalogsFromFolder(folder)
			tt.wantErr(t, err)

			if err == nil {
				assert.Equal(t, tt.wantCount, len(catalogs))

				// Verify parent relationships for nested controls test
				if tt.name == "process nested controls and set parent relationships" && len(catalogs) > 0 {
					catalog := catalogs[0]
					assert.Equal(t, 1, len(catalog.Categories))
					category := catalog.Categories[0]
					assert.Equal(t, 1, len(category.Controls))
					control := category.Controls[0]
					assert.Equal(t, 1, len(control.Controls))
					subControl := control.Controls[0]

					// Check parent relationships were set correctly
					assert.Equal(t, "category-1", subControl.CategoryName)
					assert.Equal(t, "catalog-1", subControl.CategoryCatalogId)
					assert.NotNil(t, subControl.ParentControlId)
					assert.Equal(t, "control-1", *subControl.ParentControlId)
					assert.NotNil(t, subControl.ParentControlCategoryName)
					assert.Equal(t, "category-1", *subControl.ParentControlCategoryName)
					assert.NotNil(t, subControl.ParentControlCategoryCatalogId)
					assert.Equal(t, "catalog-1", *subControl.ParentControlCategoryCatalogId)
				}
			}
		})
	}
}
