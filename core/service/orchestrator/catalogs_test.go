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

	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/auth"
	"confirmate.io/core/persistence"
	"confirmate.io/core/persistence/persistencetest"
	"confirmate.io/core/service"
	"confirmate.io/core/service/orchestrator/orchestratortest"
	"confirmate.io/core/util/assert"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestService_CreateCatalog(t *testing.T) {
	type args struct {
		req *orchestrator.CreateCatalogRequest
		ctx context.Context
	}
	type fields struct {
		db    persistence.DB
		authz service.AuthorizationStrategy
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[orchestrator.Catalog]]
		wantErr assert.WantErr
	}{
		{
			name: "happy path: with allow-all authorization strategy",
			args: args{
				req: &orchestrator.CreateCatalogRequest{
					Catalog: orchestratortest.MockCatalog1,
				},
			},
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.Catalog], args ...any) bool {
				assert.NotNil(t, got.Msg)
				return assert.Equal(t, orchestratortest.MockCatalog1.Id, got.Msg.Id) &&
					assert.Equal(t, orchestratortest.MockCatalog1.Name, got.Msg.Name) &&
					assert.Equal(t, orchestratortest.MockCatalog1.Description, got.Msg.Description)
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: with authorization strategy with permission store and admin token",
			args: args{
				req: &orchestrator.CreateCatalogRequest{
					Catalog: orchestratortest.MockCatalog1,
				},
				ctx: auth.WithClaims(
					context.Background(),
					&auth.OAuthClaims{
						IsAdminToken: true,
					},
				),
			},
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables),
				authz: &service.AuthorizationStrategyPermissionStore{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.Catalog], args ...any) bool {
				want := orchestratortest.MockCatalog1
				normalizeCatalogControls(want)
				return assert.NotNil(t, got.Msg) &&
					assert.Equal(t, want, got.Msg)
			},
			wantErr: assert.NoError,
		},
		{
			name: "authorization error",
			args: args{
				req: &orchestrator.CreateCatalogRequest{
					Catalog: orchestratortest.MockCatalog1,
				},
				ctx: auth.WithClaims(
					context.Background(),
					&auth.OAuthClaims{
						IsAdminToken: false,
					},
				),
			},
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables),
				authz: &service.AuthorizationStrategyPermissionStore{},
			},
			want: assert.Nil[*connect.Response[orchestrator.Catalog]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodePermissionDenied)
			},
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
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.ErrorContains(t, err, "invalid request")
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
		{
			name: "db error - unique constraint",
			args: args{
				req: &orchestrator.CreateCatalogRequest{
					Catalog: orchestratortest.MockCatalog1,
				},
			},
			fields: fields{
				db:    persistencetest.CreateErrorDB(t, persistence.ErrUniqueConstraintFailed, types, joinTables),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: assert.Nil[*connect.Response[orchestrator.Catalog]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeAlreadyExists)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db:    tt.fields.db,
				authz: tt.fields.authz,
			}
			res, err := svc.CreateCatalog(tt.args.ctx, connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}

func TestService_GetCatalog(t *testing.T) {
	catalog1 := orchestratortest.MockCatalog1
	normalizeCatalogControls(catalog1)

	type args struct {
		req *orchestrator.GetCatalogRequest
	}
	type fields struct {
		db persistence.DB
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
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(catalog1)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.Catalog], args ...any) bool {
				want := &orchestrator.Catalog{
					Id:          orchestratortest.MockCatalogId1,
					Name:        orchestratortest.MockCatalogName1,
					Description: orchestratortest.MockCatalogDescription1,
					Categories: []*orchestrator.Category{
						{
							Name:      orchestratortest.MockCategoryName1,
							CatalogId: orchestratortest.MockCatalogId1,
							Controls: []*orchestrator.Control{
								{
									Id:        orchestratortest.MockControlId1,
									Name:      orchestratortest.MockControlName1,
									ShortName: orchestratortest.MockControlShortName1,
									CatalogId: orchestratortest.MockCatalogId1,
									Controls:  []*orchestrator.Control{},
								},
							},
						},
						{
							Name:      orchestratortest.MockCategoryName2,
							CatalogId: orchestratortest.MockCatalogId1,
							Controls: []*orchestrator.Control{
								{
									Id:        orchestratortest.MockControlId2,
									Name:      orchestratortest.MockControlName2,
									ShortName: orchestratortest.MockControlShortName2,
									CatalogId: orchestratortest.MockCatalogId1,
									Controls:  []*orchestrator.Control{},
								},
							},
						},
					},
				}
				assert.NotNil(t, got.Msg)
				return assert.Equal(t, want, got.Msg)
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
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.ErrorContains(t, err, "invalid request")
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
		{
			name: "db error - not found",
			args: args{
				req: &orchestrator.GetCatalogRequest{
					CatalogId: orchestratortest.MockCatalog1.Id,
				},
			},
			fields: fields{
				db: persistencetest.GetErrorDB(t, persistence.ErrRecordNotFound, types, joinTables),
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
		db persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[orchestrator.ListCatalogsResponse]]
		wantErr assert.WantErr
	}{
		{
			name: "validation error",
			args: args{
				req: &orchestrator.ListCatalogsRequest{
					PageToken: "!!!invalid-base64!!!",
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.ListCatalogsResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.ErrorContains(t, err, "invalid page_token")
			},
		},
		{
			name: "db error - not found",
			args: args{
				req: &orchestrator.ListCatalogsRequest{},
			},
			fields: fields{
				db: persistencetest.ListErrorDB(t, persistence.ErrRecordNotFound, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.ListCatalogsResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeNotFound)
			},
		},
		{
			name: "happy path: list all catalogs",
			args: args{
				req: &orchestrator.ListCatalogsRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockCatalog1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockCatalog2)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ListCatalogsResponse], args ...any) bool {
				want := &orchestrator.Catalog{
					Id:          orchestratortest.MockCatalogId1,
					Name:        orchestratortest.MockCatalogName1,
					Description: orchestratortest.MockCatalogDescription1,
					Categories: []*orchestrator.Category{
						{
							Name:      orchestratortest.MockCategoryName1,
							CatalogId: orchestratortest.MockCatalogId1,
						},
						{
							Name:      orchestratortest.MockCategoryName2,
							CatalogId: orchestratortest.MockCatalogId1,
						},
					},
				}

				assert.NotNil(t, got.Msg)
				assert.Equal(t, 2, len(got.Msg.Catalogs))
				return assert.Equal(t, want, got.Msg.Catalogs[0])
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: empty list",
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
		ctx context.Context
	}
	type fields struct {
		db    persistence.DB
		authz service.AuthorizationStrategy
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[orchestrator.Catalog]]
		wantErr assert.WantErr
	}{
		{
			name: "happy path: with allow-all authorization strategy",
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
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockCatalog1)
					assert.NoError(t, err)
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.Catalog], args ...any) bool {
				assert.NotNil(t, got.Msg)
				return assert.Equal(t, orchestratortest.MockCatalog1.Id, got.Msg.Id) &&
					assert.Equal(t, "Updated Catalog", got.Msg.Name) &&
					assert.Equal(t, "Updated description", got.Msg.Description)
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: with authorization strategy with permission store and admin token",
			args: args{
				req: &orchestrator.UpdateCatalogRequest{
					Catalog: &orchestrator.Catalog{
						Id:          orchestratortest.MockCatalog1.Id,
						Name:        "Updated Catalog",
						Description: "Updated description",
					},
				},
				ctx: auth.WithClaims(
					context.Background(),
					&auth.OAuthClaims{
						IsAdminToken: true,
					},
				),
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockCatalog1)
					assert.NoError(t, err)
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.Catalog], args ...any) bool {
				assert.NotNil(t, got.Msg)
				return assert.Equal(t, orchestratortest.MockCatalog1.Id, got.Msg.Id) &&
					assert.Equal(t, "Updated Catalog", got.Msg.Name) &&
					assert.Equal(t, "Updated description", got.Msg.Description)
			},
			wantErr: assert.NoError,
		},
		{
			name: "authorization error",
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
				db:    persistencetest.NewInMemoryDB(t, types, joinTables),
				authz: &service.AuthorizationStrategyPermissionStore{},
			},
			want: assert.Nil[*connect.Response[orchestrator.Catalog]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodePermissionDenied)
			},
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
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.ErrorContains(t, err, "invalid request")
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
				db:    persistencetest.NewInMemoryDB(t, types, joinTables),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: assert.Nil[*connect.Response[orchestrator.Catalog]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeNotFound)
			},
		},
		{
			name: "db error - constraint",
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
				db:    persistencetest.UpdateErrorDB(t, persistence.ErrConstraintFailed, types, joinTables),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: assert.Nil[*connect.Response[orchestrator.Catalog]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db:    tt.fields.db,
				authz: tt.fields.authz,
			}
			res, err := svc.UpdateCatalog(tt.args.ctx, connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}

func TestService_RemoveCatalog(t *testing.T) {
	type args struct {
		req *orchestrator.RemoveCatalogRequest
		ctx context.Context
	}
	type fields struct {
		db    persistence.DB
		authz service.AuthorizationStrategy
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[emptypb.Empty]]
		wantErr assert.WantErr
	}{
		{
			name: "happy path: with allow-all authorization strategy",
			args: args{
				req: &orchestrator.RemoveCatalogRequest{
					CatalogId: orchestratortest.MockCatalog1.Id,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockCatalog1)
					assert.NoError(t, err)
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: func(t *testing.T, got *connect.Response[emptypb.Empty], args ...any) bool {
				return assert.NotNil(t, got.Msg)
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: with authorization strategy with permission store and admin token",
			args: args{
				req: &orchestrator.RemoveCatalogRequest{
					CatalogId: orchestratortest.MockCatalog1.Id,
				},
				ctx: auth.WithClaims(
					context.Background(),
					&auth.OAuthClaims{
						IsAdminToken: true,
					},
				),
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockCatalog1)
					assert.NoError(t, err)
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: func(t *testing.T, got *connect.Response[emptypb.Empty], args ...any) bool {
				return assert.NotNil(t, got.Msg)
			},
			wantErr: assert.NoError,
		},
		{
			name: "authorization error",
			args: args{
				req: &orchestrator.RemoveCatalogRequest{
					CatalogId: orchestratortest.MockCatalog1.Id,
				},
			},
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables),
				authz: &service.AuthorizationStrategyPermissionStore{},
			},
			want: assert.Nil[*connect.Response[emptypb.Empty]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodePermissionDenied)
			},
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
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.ErrorContains(t, err, "invalid request")
			},
		},
		{
			name: "db error - not found",
			args: args{
				req: &orchestrator.RemoveCatalogRequest{
					CatalogId: orchestratortest.MockCatalog1.Id,
				},
			},
			fields: fields{
				db:    persistencetest.GetErrorDB(t, persistence.ErrRecordNotFound, types, joinTables),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: assert.Nil[*connect.Response[emptypb.Empty]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db:    tt.fields.db,
				authz: tt.fields.authz,
			}
			res, err := svc.RemoveCatalog(tt.args.ctx, connect.NewRequest(tt.args.req))
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
		db persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[orchestrator.Category]]
		wantErr assert.WantErr
	}{
		{
			name: "validation error - empty request",
			args: args{
				req: &orchestrator.GetCategoryRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.Category]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.ErrorContains(t, err, "invalid request")
			},
		},
		{
			name: "happy path",
			args: args{
				req: &orchestrator.GetCategoryRequest{
					CatalogId:    orchestratortest.MockCatalogId1,
					CategoryName: orchestratortest.MockCategoryName1,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockCatalog1)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.Category], args ...any) bool {
				want := &orchestrator.Category{
					Name:      orchestratortest.MockCategoryName1,
					CatalogId: orchestratortest.MockCatalogId1,
					Controls: []*orchestrator.Control{
						{
							Id:        orchestratortest.MockControlId1,
							Name:      orchestratortest.MockControlName1,
							ShortName: orchestratortest.MockControlShortName1,
							CatalogId: orchestratortest.MockCatalogId1,
						},
					},
				}
				assert.NotNil(t, got.Msg)
				return assert.Equal(t, want, got.Msg)
			},
			wantErr: assert.NoError,
		},
		{
			name: "not found",
			args: args{
				req: &orchestrator.GetCategoryRequest{
					CatalogId:    orchestratortest.MockCatalogId1,
					CategoryName: "does-not-exist",
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
		db persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[orchestrator.ListControlsResponse]]
		wantErr assert.WantErr
	}{
		{
			name: "validation error - empty request",
			args: args{
				req: &orchestrator.ListControlsRequest{
					PageToken: "!!!invalid-base64!!!",
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.ListControlsResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.ErrorContains(t, err, "invalid page_token")
			},
		},
		{
			name: "db error - not found",
			args: args{
				req: &orchestrator.ListControlsRequest{},
			},
			fields: fields{
				db: persistencetest.ListErrorDB(t, persistence.ErrRecordNotFound, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.ListControlsResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeNotFound)
			},
		},
		{
			name: "error: filter by category_name",
			args: args{
				req: &orchestrator.ListControlsRequest{
					Filter: &orchestrator.ListControlsRequest_Filter{
						CategoryName: new(orchestratortest.MockCategoryName1),
					},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockCatalog1)
					assert.NoError(t, err)
				}),
			},
			want: assert.Nil[*connect.Response[orchestrator.ListControlsResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeUnimplemented) &&
					assert.ErrorContains(t, err, "filtering by category name is not yet implemented")
			},
		},
		{
			name: "error: filter by assurance_level",
			args: args{
				req: &orchestrator.ListControlsRequest{
					Filter: &orchestrator.ListControlsRequest_Filter{
						AssuranceLevels: []string{"high"},
					},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockCatalog1)
					assert.NoError(t, err)
				}),
			},
			want: assert.Nil[*connect.Response[orchestrator.ListControlsResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeUnimplemented) &&
					assert.ErrorContains(t, err, "filtering by assurance levels is not yet implemented")
			},
		},
		{
			name: "happy path: with filter catalog id",
			args: args{
				req: &orchestrator.ListControlsRequest{
					Filter: &orchestrator.ListControlsRequest_Filter{
						CatalogId: new(orchestratortest.MockCatalogId1),
					},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockCatalog1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockCatalog2)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ListControlsResponse], args ...any) bool {
				want := []*orchestrator.Control{
					{
						Id:        orchestratortest.MockControlId1,
						Name:      orchestratortest.MockControlName1,
						ShortName: orchestratortest.MockControlShortName1,
						CatalogId: orchestratortest.MockCatalogId1,
						Controls: []*orchestrator.Control{
							{
								Id:              orchestratortest.MockControl1SubControlId1,
								Name:            orchestratortest.MockSubControlName1,
								ShortName:       orchestratortest.MockSubControlShortName1,
								CatalogId:       orchestratortest.MockCatalogId1,
								AssuranceLevel:  new("high"),
								ParentControlId: new(orchestratortest.MockControlId1),
							},
							{
								Id:              orchestratortest.MockControl1SubControlId2,
								Name:            orchestratortest.MockSubControlName2,
								ShortName:       orchestratortest.MockSubControlShortName2,
								CatalogId:       orchestratortest.MockCatalogId1,
								AssuranceLevel:  new("medium"),
								ParentControlId: new(orchestratortest.MockControlId1),
							},
						},
					},
					{

						Id:        orchestratortest.MockControlId2,
						Name:      orchestratortest.MockControlName2,
						ShortName: orchestratortest.MockControlShortName2,
						CatalogId: orchestratortest.MockCatalogId1,
						Controls: []*orchestrator.Control{
							{
								Id:              orchestratortest.MockControl2SubControlId1,
								Name:            orchestratortest.MockSubControlName2,
								ShortName:       orchestratortest.MockSubControlShortName1,
								ParentControlId: new(orchestratortest.MockControlId2),
								CatalogId:       orchestratortest.MockCatalogId1,
							},
						},
					},
				}

				assert.NotNil(t, got.Msg)
				return assert.Equal(t, 2, len(got.Msg.Controls)) && assert.Equal(t, want, got.Msg.Controls)
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: list all",
			args: args{
				req: &orchestrator.ListControlsRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockCatalog1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockCatalog2)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ListControlsResponse], args ...any) bool {
				want := []*orchestrator.Control{
					{
						Id:        orchestratortest.MockControlId1,
						Name:      orchestratortest.MockControlName1,
						ShortName: orchestratortest.MockControlShortName1,
						CatalogId: orchestratortest.MockCatalogId1,
						Controls: []*orchestrator.Control{
							{
								Id:        orchestratortest.MockControl1SubControlId1,
								CatalogId: orchestratortest.MockCatalogId1,
								Name:      orchestratortest.MockSubControlName1,
								ShortName: orchestratortest.MockSubControlShortName1,
								// Metrics:         ,
								ParentControlId: new(orchestratortest.MockControlId1),
								AssuranceLevel:  new("high"),
							},
							{
								Id:        orchestratortest.MockControl1SubControlId2,
								CatalogId: orchestratortest.MockCatalogId1,
								Name:      orchestratortest.MockSubControlName2,
								ShortName: orchestratortest.MockSubControlShortName2,
								// Metrics:         []*assessment.Metric{MockMetric2},
								ParentControlId: new(orchestratortest.MockControlId1),
								AssuranceLevel:  new("medium"),
							},
						},
					},
					{
						Id:        orchestratortest.MockControlId2,
						CatalogId: orchestratortest.MockCatalogId1,
						Name:      orchestratortest.MockControlName2,
						ShortName: orchestratortest.MockControlShortName2,
						Controls: []*orchestrator.Control{
							{
								Id:              orchestratortest.MockControl2SubControlId1,
								CatalogId:       orchestratortest.MockCatalogId1,
								Name:            orchestratortest.MockSubControlName2,
								ShortName:       orchestratortest.MockSubControlShortName1,
								ParentControlId: new(orchestratortest.MockControlId2),
							},
						},
					},
				}

				assert.NotNil(t, got.Msg)
				assert.Equal(t, 2, len(got.Msg.Controls))
				return assert.Equal(t, want, got.Msg.Controls)
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: list full control tree",
			args: args{
				req: &orchestrator.ListControlsRequest{
					Filter: &orchestrator.ListControlsRequest_Filter{
						Full: new(true),
					},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockCatalog1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockCatalog3)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ListControlsResponse], args ...any) bool {
				want := []*orchestrator.Control{
					{
						Id:        orchestratortest.MockControlId1,
						Name:      orchestratortest.MockControlName1,
						ShortName: orchestratortest.MockControlShortName1,
						CatalogId: orchestratortest.MockCatalogId1,
						Controls: []*orchestrator.Control{
							{
								Id:              orchestratortest.MockControl1SubControlId1,
								CatalogId:       orchestratortest.MockCatalogId1,
								Name:            orchestratortest.MockSubControlName1,
								ShortName:       orchestratortest.MockSubControlShortName1,
								Metrics:         []*assessment.Metric{orchestratortest.MockMetric1},
								ParentControlId: new(orchestratortest.MockControlId1),
								AssuranceLevel:  new("high"),
							},
							{
								Id:              orchestratortest.MockControl1SubControlId2,
								CatalogId:       orchestratortest.MockCatalogId1,
								Name:            orchestratortest.MockSubControlName2,
								ShortName:       orchestratortest.MockSubControlShortName2,
								Metrics:         []*assessment.Metric{orchestratortest.MockMetric2},
								ParentControlId: new(orchestratortest.MockControlId1),
								AssuranceLevel:  new("medium"),
							},
						},
					},
					{
						Id:        orchestratortest.MockControlId2,
						CatalogId: orchestratortest.MockCatalogId1,
						Name:      orchestratortest.MockControlName2,
						ShortName: orchestratortest.MockControlShortName2,
						Controls: []*orchestrator.Control{
							{
								Id:              orchestratortest.MockControl2SubControlId1,
								CatalogId:       orchestratortest.MockCatalogId1,
								Name:            orchestratortest.MockSubControlName2,
								ShortName:       orchestratortest.MockSubControlShortName1,
								Metrics:         []*assessment.Metric{orchestratortest.MockMetric1},
								ParentControlId: new(orchestratortest.MockControlId2),
							},
						},
					},
					{

						Id:        orchestratortest.MockControlId31,
						Name:      orchestratortest.MockControlName31,
						ShortName: orchestratortest.MockControlShortName31,
						CatalogId: orchestratortest.MockCatalogId3,
						Controls: []*orchestrator.Control{
							{
								Id:              orchestratortest.MockControl31SubControlId1,
								Name:            orchestratortest.MockControl31SubControlName1,
								ShortName:       orchestratortest.MockControl31SubControlShortName1,
								Metrics:         []*assessment.Metric{orchestratortest.MockMetric2},
								ParentControlId: new(orchestratortest.MockControlId31),
								CatalogId:       orchestratortest.MockCatalogId3,
							},
						},
					},
				}

				assert.NotNil(t, got.Msg)
				assert.Equal(t, 3, len(got.Msg.Controls))
				return assert.Equal(t, want, got.Msg.Controls)
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
		db persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[orchestrator.Control]]
		wantErr assert.WantErr
	}{
		{
			name: "validation error - empty request",
			args: args{
				req: &orchestrator.GetControlRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.Control]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.ErrorContains(t, err, "invalid request")
			},
		},
		{
			name: "happy path",
			args: args{
				req: &orchestrator.GetControlRequest{
					ControlId: orchestratortest.MockControlId1,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockCatalog1)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.Control], args ...any) bool {
				want := &orchestrator.Control{
					Id:        orchestratortest.MockControlId1,
					Name:      orchestratortest.MockControlName1,
					ShortName: orchestratortest.MockControlShortName1,
					CatalogId: orchestratortest.MockCatalogId1,
					Controls: []*orchestrator.Control{
						{
							Id:              orchestratortest.MockControl1SubControlId1,
							Name:            orchestratortest.MockSubControlName1,
							CatalogId:       orchestratortest.MockCatalogId1,
							ShortName:       orchestratortest.MockSubControlShortName1,
							ParentControlId: new(orchestratortest.MockControlId1),
							Metrics:         []*assessment.Metric{orchestratortest.MockMetric1},
							AssuranceLevel:  new("high"),
						},
						{
							Id:              orchestratortest.MockControl1SubControlId2,
							Name:            orchestratortest.MockSubControlName2,
							CatalogId:       orchestratortest.MockCatalogId1,
							ShortName:       orchestratortest.MockSubControlShortName2,
							ParentControlId: new(orchestratortest.MockControlId1),
							Metrics:         []*assessment.Metric{orchestratortest.MockMetric2},
							AssuranceLevel:  new("medium"),
						},
					},
				}
				assert.NotNil(t, got.Msg)
				return assert.Equal(t, want, got.Msg)
			},
			wantErr: assert.NoError,
		},
		{
			name: "not found",
			args: args{
				req: &orchestrator.GetControlRequest{
					ControlId: "non-existent",
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
		wantDB           assert.Want[persistence.DB]
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
			wantDB: func(t *testing.T, db persistence.DB, args ...any) bool {
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
			wantDB: func(t *testing.T, db persistence.DB, args ...any) bool {
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
			wantDB: func(t *testing.T, db persistence.DB, args ...any) bool {
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
			wantDB:          assert.NotNil[persistence.DB],
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
			wantDB: assert.NotNil[persistence.DB],
		},
		{
			name:            "invalid catalogs path",
			loadDefaultCats: true,
			catalogsPath:    "/nonexistent/path",
			wantErr: func(t *testing.T, err error, args ...any) bool {
				return assert.ErrorContains(t, err, "could not load default catalogs")
			},
			wantDB: assert.NotNil[persistence.DB],
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
										Id: "control-1",
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
					var control *orchestrator.Control
					for _, candidate := range category.Controls {
						if candidate.ParentControlId == nil {
							control = candidate
							break
						}
					}
					assert.NotNil(t, control)
					assert.Equal(t, 1, len(control.Controls))
					subControl := control.Controls[0]

					// Check parent relationships were set correctly
					assert.NotEmpty(t, control.ShortName)
					assert.Equal(t, "control-1", control.ShortName)
					assert.NoError(t, uuid.Validate(control.Id))
					assert.NotEmpty(t, subControl.ShortName)
					assert.Equal(t, "sub-control-1", subControl.ShortName)
					assert.NoError(t, uuid.Validate(subControl.Id))
					assert.NotNil(t, subControl.ParentControlId)
					assert.Equal(t, control.Id, *subControl.ParentControlId)
				}
			}
		})
	}
}
