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
	"confirmate.io/core/auth"
	"confirmate.io/core/persistence"
	"confirmate.io/core/persistence/persistencetest"
	"confirmate.io/core/service"
	"confirmate.io/core/service/orchestrator/orchestratortest"
	"confirmate.io/core/util/assert"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/go-cmp/cmp"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestService_CreateAuditScope(t *testing.T) {
	type args struct {
		req     *orchestrator.CreateAuditScopeRequest
		context context.Context
	}
	type fields struct {
		db    persistence.DB
		authz service.AuthorizationStrategy
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[orchestrator.AuditScope]]
		wantErr assert.WantErr
		wantDB  assert.Want[persistence.DB]
	}{
		{
			name: "happy path: with allow-all authorization strategy",
			args: args{
				req: &orchestrator.CreateAuditScopeRequest{
					AuditScope: &orchestrator.AuditScope{
						TargetOfEvaluationId: orchestratortest.MockAuditScope1.TargetOfEvaluationId,
						CatalogId:            orchestratortest.MockAuditScope1.CatalogId,
						Name:                 orchestratortest.MockScopeName1,
					},
				},
			},
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.AuditScope], args ...any) bool {
				return assert.NotNil(t, got.Msg) &&
					assert.NotEmpty(t, got.Msg.Id)
			},
			wantErr: assert.NoError,
			wantDB: func(t *testing.T, db persistence.DB, msgAndArgs ...any) bool {
				res := assert.Is[*connect.Response[orchestrator.AuditScope]](t, msgAndArgs[0])
				assert.NotNil(t, res)

				got := assert.InDB[orchestrator.AuditScope](t, db, res.Msg.Id)
				want := &orchestrator.AuditScope{
					// ID is generated, so we can't assert on it
					TargetOfEvaluationId: orchestratortest.MockAuditScope1.TargetOfEvaluationId,
					CatalogId:            orchestratortest.MockAuditScope1.CatalogId,
					Name:                 orchestratortest.MockScopeName1,
				}

				// Check if ID is generated and not empty
				assert.NotEmpty(t, got.Id)
				// Remove ID from got for comparison since it's generated
				got.Id = ""
				return assert.Equal(t, want, got)
			},
		},
		{
			name: "happy path: with authorization strategy with permission store and admin token",
			args: args{
				req: &orchestrator.CreateAuditScopeRequest{
					AuditScope: &orchestrator.AuditScope{
						TargetOfEvaluationId: orchestratortest.MockAuditScope1.TargetOfEvaluationId,
						CatalogId:            orchestratortest.MockAuditScope1.CatalogId,
						Name:                 orchestratortest.MockScopeName1,
					},
				},
				context: auth.WithClaims(context.Background(), &auth.OAuthClaims{
					IsAdminToken: true,
				}),
			},
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.AuditScope], args ...any) bool {
				return assert.NotNil(t, got.Msg) &&
					assert.NotEmpty(t, got.Msg.Id)
			},
			wantErr: assert.NoError,
			wantDB: func(t *testing.T, db persistence.DB, msgAndArgs ...any) bool {
				res := assert.Is[*connect.Response[orchestrator.AuditScope]](t, msgAndArgs[0])
				assert.NotNil(t, res)

				got := assert.InDB[orchestrator.AuditScope](t, db, res.Msg.Id)
				want := &orchestrator.AuditScope{
					// ID is generated, so we can't assert on it
					TargetOfEvaluationId: orchestratortest.MockAuditScope1.TargetOfEvaluationId,
					CatalogId:            orchestratortest.MockAuditScope1.CatalogId,
					Name:                 orchestratortest.MockScopeName1,
				}

				// Check if ID is generated and not empty
				assert.NotEmpty(t, got.Id)
				// Remove ID from got for comparison since it's generated
				got.Id = ""
				return assert.Equal(t, want, got)
			},
		},
		{
			name: "happy path: with authorization strategy with permission store and user permissions allowing access",
			args: args{
				req: &orchestrator.CreateAuditScopeRequest{
					AuditScope: &orchestrator.AuditScope{
						TargetOfEvaluationId: orchestratortest.MockAuditScope1.TargetOfEvaluationId,
						CatalogId:            orchestratortest.MockAuditScope1.CatalogId,
						Name:                 orchestratortest.MockScopeName1,
					},
				},
				context: auth.WithClaims(context.Background(), &auth.OAuthClaims{
					IsAdminToken: false,
					RegisteredClaims: jwt.RegisteredClaims{
						Subject: orchestratortest.MockUserId1,
						Issuer:  orchestratortest.MockUserIssuer1,
					},
				}),
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
				authz: &service.AuthorizationStrategyPermissionStore{
					Permissions: permissionStore{
						db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
							err := d.Create(orchestratortest.MockUserPermissionsToEAdmin)
							assert.NoError(t, err)
						}),
					},
				},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.AuditScope], args ...any) bool {
				want := &orchestrator.AuditScope{
					// ID is generated, so we can't assert on it
					TargetOfEvaluationId: orchestratortest.MockAuditScope1.TargetOfEvaluationId,
					CatalogId:            orchestratortest.MockAuditScope1.CatalogId,
					Name:                 orchestratortest.MockScopeName1,
				}
				return assert.Equal(t, want, got.Msg, cmp.Options{
					protocmp.IgnoreFields(&orchestrator.AuditScope{}, "id"),
				})
			},
			wantErr: assert.NoError,
			wantDB: func(t *testing.T, db persistence.DB, msgAndArgs ...any) bool {
				res := assert.Is[*connect.Response[orchestrator.AuditScope]](t, msgAndArgs[0])
				assert.NotNil(t, res)

				got := assert.InDB[orchestrator.AuditScope](t, db, res.Msg.Id)
				want := &orchestrator.AuditScope{
					// ID is generated, so we can't assert on it
					TargetOfEvaluationId: orchestratortest.MockAuditScope1.TargetOfEvaluationId,
					CatalogId:            orchestratortest.MockAuditScope1.CatalogId,
					Name:                 orchestratortest.MockScopeName1,
				}

				// Check if ID is generated and not empty
				assert.NotEmpty(t, got.Id)
				return assert.Equal(t, want, got, cmp.Options{
					protocmp.IgnoreFields(&orchestrator.AuditScope{}, "id"),
				})
			},
		},
		{
			name: "validation error - empty request",
			args: args{
				req: &orchestrator.CreateAuditScopeRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.AuditScope]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.ErrorContains(t, err, "invalid request:")
			},
			wantDB: assert.NotNil[persistence.DB],
		},
		{
			name: "validation error - missing target of evaluation id",
			args: args{
				req: &orchestrator.CreateAuditScopeRequest{
					AuditScope: &orchestrator.AuditScope{
						CatalogId: orchestratortest.MockAuditScope1.CatalogId,
					},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.AuditScope]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.IsValidationError(t, err, "audit_scope.target_of_evaluation_id")
			},
			wantDB: assert.NotNil[persistence.DB],
		},
		{
			name: "authorization failure",
			args: args{
				req: &orchestrator.CreateAuditScopeRequest{
					AuditScope: &orchestrator.AuditScope{
						TargetOfEvaluationId: orchestratortest.MockAuditScope1.TargetOfEvaluationId,
						CatalogId:            orchestratortest.MockAuditScope1.CatalogId,
						Name:                 orchestratortest.MockScopeName1,
					},
				},
			},
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables),
				authz: &denyAuthorizationStrategy{},
			},
			want: assert.Nil[*connect.Response[orchestrator.AuditScope]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodePermissionDenied)
			},
			wantDB: assert.NotNil[persistence.DB],
		},
		{
			name: "db error - unique constraint",
			args: args{
				req: &orchestrator.CreateAuditScopeRequest{
					AuditScope: &orchestrator.AuditScope{
						TargetOfEvaluationId: orchestratortest.MockAuditScope1.TargetOfEvaluationId,
						CatalogId:            orchestratortest.MockAuditScope1.CatalogId,
						Name:                 orchestratortest.MockScopeName1,
					},
				},
			},
			fields: fields{
				db:    persistencetest.CreateErrorDB(t, persistence.ErrUniqueConstraintFailed, types, joinTables),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: assert.Nil[*connect.Response[orchestrator.AuditScope]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeAlreadyExists)
			},
			wantDB: assert.NotNil[persistence.DB],
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db:    tt.fields.db,
				authz: tt.fields.authz,
			}
			res, err := svc.CreateAuditScope(tt.args.context, connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
			tt.wantDB(t, tt.fields.db, res)
		})
	}
}

func TestService_GetAuditScope(t *testing.T) {
	type args struct {
		req     *orchestrator.GetAuditScopeRequest
		context context.Context
	}
	type fields struct {
		db    persistence.DB
		authz service.AuthorizationStrategy
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[orchestrator.AuditScope]]
		wantErr assert.WantErr
	}{
		{
			name: "happy path",
			args: args{
				req: &orchestrator.GetAuditScopeRequest{
					AuditScopeId: orchestratortest.MockAuditScope1.Id,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockAuditScope1)
					assert.NoError(t, err)
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.AuditScope], args ...any) bool {
				return assert.NotNil(t, got.Msg) &&
					assert.Equal(t, orchestratortest.MockAuditScope1.Id, got.Msg.Id)
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: with allow-all authorization strategy",
			args: args{
				req: &orchestrator.GetAuditScopeRequest{
					AuditScopeId: orchestratortest.MockAuditScope1.Id,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockAuditScope1)
					assert.NoError(t, err)
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.AuditScope], args ...any) bool {
				return assert.NotNil(t, got.Msg) &&
					assert.Equal(t, orchestratortest.MockAuditScope1.Id, got.Msg.Id)
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: with authorization strategy with permission store and admin token",
			args: args{
				req: &orchestrator.GetAuditScopeRequest{
					AuditScopeId: orchestratortest.MockAuditScope1.Id,
				},
				context: auth.WithClaims(context.Background(), &auth.OAuthClaims{
					IsAdminToken: true,
				}),
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockAuditScope1)
					assert.NoError(t, err)
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.AuditScope], args ...any) bool {
				return assert.NotNil(t, got.Msg) &&
					assert.Equal(t, orchestratortest.MockAuditScope1.Id, got.Msg.Id)
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: with authorization strategy with permission store and user permissions allowing access", args: args{
				req: &orchestrator.GetAuditScopeRequest{
					AuditScopeId: orchestratortest.MockAuditScope1.Id,
				},
				context: auth.WithClaims(context.Background(), &auth.OAuthClaims{
					RegisteredClaims: jwt.RegisteredClaims{
						Subject: orchestratortest.MockUserId1,
						Issuer:  orchestratortest.MockUserIssuer1,
					},
				}),
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockAuditScope1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockAuditScope2)
					assert.NoError(t, err)
				}),
				authz: &service.AuthorizationStrategyPermissionStore{
					Permissions: permissionStore{
						db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
							err := d.Create(orchestratortest.MockUserPermissionsAuditScopeAdmin)
							assert.NoError(t, err)
						}),
					},
				},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.AuditScope], args ...any) bool {
				return assert.NotNil(t, got.Msg) &&
					assert.Equal(t, orchestratortest.MockAuditScope1, got.Msg)
			},
			wantErr: assert.NoError,
		},
		{
			name: "validation error - empty request",
			args: args{
				req: &orchestrator.GetAuditScopeRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.AuditScope]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.ErrorContains(t, err, "invalid request:")
			},
		},
		{
			name: "authorization failure",
			args: args{
				req: &orchestrator.GetAuditScopeRequest{
					AuditScopeId: orchestratortest.MockAuditScope1.Id,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockAuditScope1)
					assert.NoError(t, err)
				}),
				authz: &denyAuthorizationStrategy{},
			},
			want: assert.Nil[*connect.Response[orchestrator.AuditScope]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodePermissionDenied)
			},
		},
		{
			name: "db error - not found",
			args: args{
				req: &orchestrator.GetAuditScopeRequest{
					AuditScopeId: orchestratortest.MockAuditScope1.Id,
				},
			},
			fields: fields{
				db:    persistencetest.GetErrorDB(t, persistence.ErrRecordNotFound, types, joinTables),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: assert.Nil[*connect.Response[orchestrator.AuditScope]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeNotFound) &&
					assert.ErrorContains(t, err, "audit scope not found")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db:    tt.fields.db,
				authz: tt.fields.authz,
			}
			res, err := svc.GetAuditScope(tt.args.context, connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}

func TestService_ListAuditScopes(t *testing.T) {
	type args struct {
		req     *orchestrator.ListAuditScopesRequest
		context context.Context
	}
	type fields struct {
		db    persistence.DB
		authz service.AuthorizationStrategy
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[orchestrator.ListAuditScopesResponse]]
		wantErr assert.WantErr
	}{
		{
			name: "validation error",
			args: args{
				req: &orchestrator.ListAuditScopesRequest{
					PageToken: "!!!invalid-base64!!!",
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.ListAuditScopesResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.ErrorContains(t, err, "invalid page_token")
			},
		},
		{
			name: "authorization failure returns empty list",
			args: args{
				req: &orchestrator.ListAuditScopesRequest{
					Filter: &orchestrator.ListAuditScopesRequest_Filter{
						TargetOfEvaluationId: &orchestratortest.MockAuditScope1.TargetOfEvaluationId,
					},
				},
			},
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables),
				authz: &denyAuthorizationStrategy{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ListAuditScopesResponse], _ ...any) bool {
				return assert.NotNil(t, got) && assert.Equal(t, 0, len(got.Msg.AuditScopes))
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: user is not authorized to view any results for the specified target of evaluation",
			args: args{
				req: &orchestrator.ListAuditScopesRequest{},
				context: auth.WithClaims(context.Background(), &auth.OAuthClaims{
					RegisteredClaims: jwt.RegisteredClaims{
						Subject: orchestratortest.MockUserId1,
						Issuer:  orchestratortest.MockUserIssuer1,
					},
				}),
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockAuditScope1)
					assert.NoError(t, err)
				}),
				authz: &service.AuthorizationStrategyPermissionStore{
					Permissions: permissionStore{
						db: persistencetest.NewInMemoryDB(t, types, joinTables),
					},
				},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ListAuditScopesResponse], _ ...any) bool {
				return assert.Empty(t, got.Msg.AuditScopes)
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: with allow-all authorization strategy",
			args: args{
				req: &orchestrator.ListAuditScopesRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockAuditScope1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockAuditScope2)
					assert.NoError(t, err)
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ListAuditScopesResponse], args ...any) bool {
				return assert.NotNil(t, got.Msg) &&
					assert.Equal(t, 2, len(got.Msg.AuditScopes))
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: with authorization strategy with permission store and admin token",
			args: args{
				req: &orchestrator.ListAuditScopesRequest{},
				context: auth.WithClaims(context.Background(), &auth.OAuthClaims{
					IsAdminToken: true,
				}),
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockAuditScope1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockAuditScope2)
					assert.NoError(t, err)
				}),
				authz: &service.AuthorizationStrategyPermissionStore{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ListAuditScopesResponse], args ...any) bool {
				return assert.NotNil(t, got.Msg) &&
					assert.Equal(t, 2, len(got.Msg.AuditScopes))
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: with authorization strategy with permission store and user permissions allowing access",
			args: args{
				req: &orchestrator.ListAuditScopesRequest{},
				context: auth.WithClaims(context.Background(), &auth.OAuthClaims{
					RegisteredClaims: jwt.RegisteredClaims{
						Subject: orchestratortest.MockUserId1,
						Issuer:  orchestratortest.MockUserIssuer1,
					},
				}),
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockAuditScope1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockAuditScope2)
					assert.NoError(t, err)
				}),
				authz: &service.AuthorizationStrategyPermissionStore{
					Permissions: permissionStore{
						db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
							err := d.Create(orchestratortest.MockUserPermissionsAuditScopeAdmin)
							assert.NoError(t, err)
						}),
					},
				},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ListAuditScopesResponse], args ...any) bool {
				return assert.NotNil(t, got.Msg) &&
					assert.Equal(t, 1, len(got.Msg.AuditScopes)) &&
					assert.Equal(t, orchestratortest.MockAuditScope1, got.Msg.AuditScopes[0])
			},
			wantErr: assert.NoError,
		},
		{
			name: "filter by target of evaluation",
			args: args{
				req: &orchestrator.ListAuditScopesRequest{
					Filter: &orchestrator.ListAuditScopesRequest_Filter{
						TargetOfEvaluationId: &orchestratortest.MockAuditScope1.TargetOfEvaluationId,
					},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockAuditScope1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockAuditScope2)
					assert.NoError(t, err)
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ListAuditScopesResponse], args ...any) bool {
				return assert.NotNil(t, got.Msg) &&
					assert.Equal(t, 1, len(got.Msg.AuditScopes))
			},
			wantErr: assert.NoError,
		},
		{
			name: "filter by catalog",
			args: args{
				req: &orchestrator.ListAuditScopesRequest{
					Filter: &orchestrator.ListAuditScopesRequest_Filter{
						CatalogId: &orchestratortest.MockAuditScope1.CatalogId,
					},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockAuditScope1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockAuditScope2)
					assert.NoError(t, err)
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ListAuditScopesResponse], args ...any) bool {
				return assert.NotNil(t, got.Msg) &&
					assert.Equal(t, 1, len(got.Msg.AuditScopes))
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db:    tt.fields.db,
				authz: tt.fields.authz,
			}
			res, err := svc.ListAuditScopes(tt.args.context, connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}

func TestService_UpdateAuditScope(t *testing.T) {
	type args struct {
		req     *orchestrator.UpdateAuditScopeRequest
		context context.Context
	}
	type fields struct {
		db    persistence.DB
		authz service.AuthorizationStrategy
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[orchestrator.AuditScope]]
		wantErr assert.WantErr
	}{
		{
			name: "happy path: with allow-all authorization strategy",
			args: args{
				req: &orchestrator.UpdateAuditScopeRequest{
					AuditScope: &orchestrator.AuditScope{
						Id:                   orchestratortest.MockAuditScope1.Id,
						Name:                 orchestratortest.MockAuditScope1.Name + " Updated",
						TargetOfEvaluationId: orchestratortest.MockToeId2,
						CatalogId:            "catalog-1-updated",
					},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockAuditScope1)
					assert.NoError(t, err)
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.AuditScope], args ...any) bool {
				want := &orchestrator.AuditScope{
					Id:                   orchestratortest.MockAuditScope1.Id,
					Name:                 orchestratortest.MockAuditScope1.Name + " Updated",
					TargetOfEvaluationId: orchestratortest.MockToeId2,
					CatalogId:            "catalog-1-updated",
				}
				return assert.NotNil(t, got.Msg) &&
					assert.Equal(t, want, got.Msg)
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: with authorization strategy with permission store and admin token",
			args: args{
				req: &orchestrator.UpdateAuditScopeRequest{
					AuditScope: &orchestrator.AuditScope{
						Id:                   orchestratortest.MockAuditScope1.Id,
						Name:                 orchestratortest.MockAuditScope1.Name + " Updated",
						TargetOfEvaluationId: orchestratortest.MockToeId2,
						CatalogId:            "catalog-1-updated",
					},
				},
				context: auth.WithClaims(context.Background(), &auth.OAuthClaims{
					IsAdminToken: true,
				}),
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockAuditScope1)
					assert.NoError(t, err)
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.AuditScope], args ...any) bool {
				want := &orchestrator.AuditScope{
					Id:                   orchestratortest.MockAuditScope1.Id,
					Name:                 orchestratortest.MockAuditScope1.Name + " Updated",
					TargetOfEvaluationId: orchestratortest.MockToeId2,
					CatalogId:            "catalog-1-updated",
				}
				return assert.NotNil(t, got.Msg) &&
					assert.Equal(t, want, got.Msg)
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: with authorization strategy with permission store and user permissions allowing access",
			args: args{
				req: &orchestrator.UpdateAuditScopeRequest{
					AuditScope: &orchestrator.AuditScope{
						Id:                   orchestratortest.MockScopeId1,
						Name:                 orchestratortest.MockAuditScope1.Name + " Updated",
						TargetOfEvaluationId: orchestratortest.MockToeId2,
						CatalogId:            "catalog-1-updated",
					},
				},
				context: auth.WithClaims(context.Background(), &auth.OAuthClaims{
					RegisteredClaims: jwt.RegisteredClaims{
						Subject: orchestratortest.MockUserId1,
						Issuer:  orchestratortest.MockUserIssuer1,
					},
				}),
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockUser1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockAuditScope1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockAuditScope2)
					assert.NoError(t, err)
				}),
				authz: &service.AuthorizationStrategyPermissionStore{
					Permissions: permissionStore{
						db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
							err := d.Create(orchestratortest.MockUserPermissionsAuditScopeAdmin)
							assert.NoError(t, err)
						}),
					},
				},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.AuditScope], args ...any) bool {
				want := &orchestrator.AuditScope{
					Id:                   orchestratortest.MockAuditScope1.Id,
					Name:                 orchestratortest.MockAuditScope1.Name + " Updated",
					TargetOfEvaluationId: orchestratortest.MockToeId2,
					CatalogId:            "catalog-1-updated",
				}
				return assert.NotNil(t, got.Msg) &&
					assert.Equal(t, want, got.Msg)
			},
			wantErr: assert.NoError,
		},
		{
			name: "validation error - empty request",
			args: args{
				req: &orchestrator.UpdateAuditScopeRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.AuditScope]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.ErrorContains(t, err, "invalid request")
			},
		},
		{
			name: "validation error - missing id",
			args: args{
				req: &orchestrator.UpdateAuditScopeRequest{
					AuditScope: &orchestrator.AuditScope{
						TargetOfEvaluationId: orchestratortest.MockToeId1,
					},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.AuditScope]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.IsValidationError(t, err, "audit_scope.id")
			},
		},
		{
			name: "not found",
			args: args{
				req: &orchestrator.UpdateAuditScopeRequest{
					AuditScope: &orchestrator.AuditScope{
						Id:                   orchestratortest.MockNonExistentId,
						Name:                 "Non-existent Scope",
						TargetOfEvaluationId: orchestratortest.MockAuditScope1.TargetOfEvaluationId,
						CatalogId:            orchestratortest.MockAuditScope1.CatalogId,
					},
				},
			},
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: assert.Nil[*connect.Response[orchestrator.AuditScope]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeNotFound) &&
					assert.ErrorContains(t, err, "audit scope not found")
			},
		},
		{
			name: "authorization failure",
			args: args{
				req: &orchestrator.UpdateAuditScopeRequest{
					AuditScope: &orchestrator.AuditScope{
						Id:                   orchestratortest.MockAuditScope1.Id,
						Name:                 orchestratortest.MockAuditScope1.Name + " Updated",
						TargetOfEvaluationId: orchestratortest.MockAuditScope1.TargetOfEvaluationId,
						CatalogId:            orchestratortest.MockAuditScope1.CatalogId,
					},
				},
			},
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables),
				authz: &denyAuthorizationStrategy{},
			},
			want: assert.Nil[*connect.Response[orchestrator.AuditScope]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodePermissionDenied)
			},
		},
		{
			name: "db error - constraint",
			args: args{
				req: &orchestrator.UpdateAuditScopeRequest{
					AuditScope: &orchestrator.AuditScope{
						Id:                   orchestratortest.MockAuditScope1.Id,
						Name:                 orchestratortest.MockAuditScope1.Name + " Updated",
						TargetOfEvaluationId: orchestratortest.MockAuditScope1.TargetOfEvaluationId,
						CatalogId:            orchestratortest.MockAuditScope1.CatalogId,
					},
				},
			},
			fields: fields{
				db:    persistencetest.UpdateErrorDB(t, persistence.ErrConstraintFailed, types, joinTables),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: assert.Nil[*connect.Response[orchestrator.AuditScope]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.ErrorContains(t, err, service.ErrConstraintFailed.Error())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db:    tt.fields.db,
				authz: tt.fields.authz,
			}
			res, err := svc.UpdateAuditScope(tt.args.context, connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}

func TestService_RemoveAuditScope(t *testing.T) {
	type args struct {
		req     *orchestrator.RemoveAuditScopeRequest
		context context.Context
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
				req: &orchestrator.RemoveAuditScopeRequest{
					AuditScopeId: orchestratortest.MockAuditScope1.Id,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockAuditScope1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockAuditScope2)
					assert.NoError(t, err)
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: func(t *testing.T, got *connect.Response[emptypb.Empty], args ...any) bool {
				return assert.NotNil(t, got) &&
					assert.Empty(t, got.Msg)
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: with authorization strategy with permission store and admin token",
			args: args{
				req: &orchestrator.RemoveAuditScopeRequest{
					AuditScopeId: orchestratortest.MockAuditScope1.Id,
				},
				context: auth.WithClaims(context.Background(), &auth.OAuthClaims{
					IsAdminToken: true,
				}),
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockAuditScope1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockAuditScope2)
					assert.NoError(t, err)
				}),
				authz: &service.AuthorizationStrategyPermissionStore{},
			},
			want: func(t *testing.T, got *connect.Response[emptypb.Empty], args ...any) bool {
				return assert.NotNil(t, got) &&
					assert.Empty(t, got.Msg)
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: with authorization strategy with permission store and user permissions allowing access",
			args: args{
				req: &orchestrator.RemoveAuditScopeRequest{
					AuditScopeId: orchestratortest.MockAuditScope1.Id,
				},
				context: auth.WithClaims(context.Background(), &auth.OAuthClaims{
					RegisteredClaims: jwt.RegisteredClaims{
						Subject: orchestratortest.MockUserId1,
						Issuer:  orchestratortest.MockUserIssuer1,
					},
				}),
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockAuditScope1)
					assert.NoError(t, err)
				}),
				authz: &service.AuthorizationStrategyPermissionStore{
					Permissions: permissionStore{
						db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
							err := d.Create(orchestratortest.MockUserPermissionsAuditScopeAdmin)
							assert.NoError(t, err)
						}),
					},
				},
			},
			want: func(t *testing.T, got *connect.Response[emptypb.Empty], args ...any) bool {
				return assert.NotNil(t, got) &&
					assert.Empty(t, got.Msg)
			},
			wantErr: assert.NoError,
		},
		{
			name: "validation error - empty request",
			args: args{
				req: &orchestrator.RemoveAuditScopeRequest{},
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
			name: "authorization failure",
			args: args{
				req: &orchestrator.RemoveAuditScopeRequest{
					AuditScopeId: orchestratortest.MockAuditScope1.Id,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockAuditScope1)
					assert.NoError(t, err)
				}),
				authz: &denyAuthorizationStrategy{},
			},
			want: assert.Nil[*connect.Response[emptypb.Empty]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodePermissionDenied) &&
					assert.ErrorContains(t, err, service.ErrPermissionDenied.Error())
			},
		},
		{
			name: "authorization error: with authorization strategy with permission store and user permissions ",
			args: args{
				req: &orchestrator.RemoveAuditScopeRequest{
					AuditScopeId: orchestratortest.MockAuditScope2.Id,
				},
				context: auth.WithClaims(context.Background(), &auth.OAuthClaims{
					RegisteredClaims: jwt.RegisteredClaims{
						Subject: orchestratortest.MockUserId1,
						Issuer:  orchestratortest.MockUserIssuer1,
					},
				}),
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockAuditScope2)
					assert.NoError(t, err)
				}),
				authz: &service.AuthorizationStrategyPermissionStore{
					Permissions: permissionStore{
						db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
							err := d.Create(orchestratortest.MockUserPermissionsAuditScopeAdmin)
							assert.NoError(t, err)
						}),
					},
				},
			},
			want: assert.Nil[*connect.Response[emptypb.Empty]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodePermissionDenied) &&
					assert.ErrorContains(t, err, service.ErrPermissionDenied.Error())
			},
		},
		{
			name: "db error - not found",
			args: args{
				req: &orchestrator.RemoveAuditScopeRequest{
					AuditScopeId: orchestratortest.MockAuditScope1.Id,
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
			res, err := svc.RemoveAuditScope(tt.args.context, connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}
