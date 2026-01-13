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
				cErr := assert.Is[*connect.Error](t, err)
				return assert.Equal(t, connect.CodeNotFound, cErr.Code())
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
				cErr := assert.Is[*connect.Error](t, err)
				return assert.Equal(t, connect.CodeNotFound, cErr.Code())
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
				cErr := assert.Is[*connect.Error](t, err)
				return assert.Equal(t, connect.CodeNotFound, cErr.Code())
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
				cErr := assert.Is[*connect.Error](t, err)
				return assert.Equal(t, connect.CodeNotFound, cErr.Code())
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
