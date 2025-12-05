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
		want    assert.Want[*orchestrator.Catalog]
		wantErr assert.WantErr[*connect.Error]
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
			want: func(t *testing.T, got *orchestrator.Catalog, args ...any) bool {
				return assert.NotNil(t, got) &&
					assert.Equal(t, orchestratortest.MockCatalog1.Id, got.Id)
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: tt.fields.db,
			}
			res, err := svc.CreateCatalog(context.Background(), connect.NewRequest(tt.args.req))
			assert.WantResponse(t, res, err, tt.want, tt.wantErr)
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
		want    assert.Want[*orchestrator.Catalog]
		wantErr assert.WantErr[*connect.Error]
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
			want: func(t *testing.T, got *orchestrator.Catalog, args ...any) bool {
				return assert.NotNil(t, got) &&
					assert.Equal(t, orchestratortest.MockCatalog1.Id, got.Id)
			},
			wantErr: nil,
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
			want: nil,
			wantErr: func(t *testing.T, err *connect.Error, msgAndArgs ...any) bool {
				return assert.Equal(t, connect.CodeNotFound, err.Code())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: tt.fields.db,
			}
			res, err := svc.GetCatalog(context.Background(), connect.NewRequest(tt.args.req))
			assert.WantResponse(t, res, err, tt.want, tt.wantErr)
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
		want    assert.Want[*orchestrator.ListCatalogsResponse]
		wantErr assert.WantErr[*connect.Error]
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
			want: func(t *testing.T, got *orchestrator.ListCatalogsResponse, args ...any) bool {
				return assert.NotNil(t, got) &&
					assert.Equal(t, 2, len(got.Catalogs))
			},
			wantErr: nil,
		},
		{
			name: "empty list",
			args: args{
				req: &orchestrator.ListCatalogsRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: func(t *testing.T, got *orchestrator.ListCatalogsResponse, args ...any) bool {
				return assert.NotNil(t, got) &&
					assert.Equal(t, 0, len(got.Catalogs))
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: tt.fields.db,
			}
			res, err := svc.ListCatalogs(context.Background(), connect.NewRequest(tt.args.req))
			assert.WantResponse(t, res, err, tt.want, tt.wantErr)
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
		want    assert.Want[*orchestrator.Catalog]
		wantErr assert.WantErr[*connect.Error]
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
			want: func(t *testing.T, got *orchestrator.Catalog, args ...any) bool {
				return assert.NotNil(t, got) &&
					assert.Equal(t, "Updated Catalog", got.Name)
			},
			wantErr: nil,
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
			want: nil,
			wantErr: func(t *testing.T, err *connect.Error, msgAndArgs ...any) bool {
				return assert.Equal(t, connect.CodeNotFound, err.Code())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: tt.fields.db,
			}
			res, err := svc.UpdateCatalog(context.Background(), connect.NewRequest(tt.args.req))
			assert.WantResponse(t, res, err, tt.want, tt.wantErr)
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
		want    assert.Want[*emptypb.Empty]
		wantErr assert.WantErr[*connect.Error]
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
			want: func(t *testing.T, got *emptypb.Empty, args ...any) bool {
				return assert.NotNil(t, got)
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: tt.fields.db,
			}
			res, err := svc.RemoveCatalog(context.Background(), connect.NewRequest(tt.args.req))
			assert.WantResponse(t, res, err, tt.want, tt.wantErr)
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
		want    assert.Want[*orchestrator.Category]
		wantErr assert.WantErr[*connect.Error]
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
			want: func(t *testing.T, got *orchestrator.Category, args ...any) bool {
				return assert.NotNil(t, got) &&
					assert.Equal(t, orchestratortest.MockCategory1.Name, got.Name)
			},
			wantErr: nil,
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
			want: nil,
			wantErr: func(t *testing.T, err *connect.Error, msgAndArgs ...any) bool {
				return assert.Equal(t, connect.CodeNotFound, err.Code())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: tt.fields.db,
			}
			res, err := svc.GetCategory(context.Background(), connect.NewRequest(tt.args.req))
			assert.WantResponse(t, res, err, tt.want, tt.wantErr)
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
		want    assert.Want[*orchestrator.ListControlsResponse]
		wantErr assert.WantErr[*connect.Error]
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
			want: func(t *testing.T, got *orchestrator.ListControlsResponse, args ...any) bool {
				return assert.NotNil(t, got) &&
					assert.Equal(t, 2, len(got.Controls))
			},
			wantErr: nil,
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
			want: func(t *testing.T, got *orchestrator.ListControlsResponse, args ...any) bool {
				return assert.NotNil(t, got) &&
					assert.Equal(t, 1, len(got.Controls))
			},
			wantErr: nil,
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
			want: func(t *testing.T, got *orchestrator.ListControlsResponse, args ...any) bool {
				return assert.NotNil(t, got) &&
					assert.Equal(t, 1, len(got.Controls))
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: tt.fields.db,
			}
			res, err := svc.ListControls(context.Background(), connect.NewRequest(tt.args.req))
			assert.WantResponse(t, res, err, tt.want, tt.wantErr)
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
		want    assert.Want[*orchestrator.Control]
		wantErr assert.WantErr[*connect.Error]
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
			want: func(t *testing.T, got *orchestrator.Control, args ...any) bool {
				return assert.NotNil(t, got) &&
					assert.Equal(t, orchestratortest.MockControl1.Id, got.Id)
			},
			wantErr: nil,
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
			want: nil,
			wantErr: func(t *testing.T, err *connect.Error, msgAndArgs ...any) bool {
				return assert.Equal(t, connect.CodeNotFound, err.Code())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: tt.fields.db,
			}
			res, err := svc.GetControl(context.Background(), connect.NewRequest(tt.args.req))
			assert.WantResponse(t, res, err, tt.want, tt.wantErr)
		})
	}
}
