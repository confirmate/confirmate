// Copyright 2026 Fraunhofer AISEC
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
	"connectrpc.com/connect"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestService_CreateControlImplementation(t *testing.T) {
	type args struct {
		req     *orchestrator.CreateControlImplementationRequest
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
		want    assert.Want[*connect.Response[orchestrator.ControlImplementation]]
		wantErr assert.WantErr
		wantDB  assert.Want[persistence.DB]
	}{
		{
			name: "happy path: with allow-all authorization strategy",
			args: args{
				req: &orchestrator.CreateControlImplementationRequest{
					ControlImplementation: &orchestrator.ControlImplementation{
						AuditScopeId:             orchestratortest.MockControlImplementation1.AuditScopeId,
						TargetOfEvaluationId:     orchestratortest.MockControlImplementation1.TargetOfEvaluationId,
						ControlId:                orchestratortest.MockControlImplementation1.ControlId,
						ControlCategoryName:      orchestratortest.MockControlImplementation1.ControlCategoryName,
						ControlCategoryCatalogId: orchestratortest.MockControlImplementation1.ControlCategoryCatalogId,
					},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					assert.NoError(t, d.Create(orchestratortest.MockAuditScope1))
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ControlImplementation], args ...any) bool {
				return assert.NotNil(t, got.Msg) &&
					assert.NotEmpty(t, got.Msg.Id) &&
					assert.Equal(t,
						orchestrator.ControlImplementationState_CONTROL_IMPLEMENTATION_STATE_OPEN,
						got.Msg.State)
			},
			wantErr: assert.NoError,
			wantDB: func(t *testing.T, db persistence.DB, msgAndArgs ...any) bool {
				res := assert.Is[*connect.Response[orchestrator.ControlImplementation]](t, msgAndArgs[0])
				assert.NotNil(t, res)

				got := assert.InDB[orchestrator.ControlImplementation](t, db, res.Msg.Id)
				want := &orchestrator.ControlImplementation{
					AuditScopeId:             orchestratortest.MockControlImplementation1.AuditScopeId,
					TargetOfEvaluationId:     orchestratortest.MockControlImplementation1.TargetOfEvaluationId,
					ControlId:                orchestratortest.MockControlImplementation1.ControlId,
					ControlCategoryName:      orchestratortest.MockControlImplementation1.ControlCategoryName,
					ControlCategoryCatalogId: orchestratortest.MockControlImplementation1.ControlCategoryCatalogId,
					State:                    orchestrator.ControlImplementationState_CONTROL_IMPLEMENTATION_STATE_OPEN,
				}
				return assert.Equal(t, want, got, protocmp.IgnoreFields(&orchestrator.ControlImplementation{},
					"id", "created_at", "updated_at", "transitions"))
			},
		},
		{
			name: "happy path: with permission store and admin token",
			args: args{
				req: &orchestrator.CreateControlImplementationRequest{
					ControlImplementation: &orchestrator.ControlImplementation{
						AuditScopeId:             orchestratortest.MockControlImplementation1.AuditScopeId,
						TargetOfEvaluationId:     orchestratortest.MockControlImplementation1.TargetOfEvaluationId,
						ControlId:                orchestratortest.MockControlImplementation1.ControlId,
						ControlCategoryName:      orchestratortest.MockControlImplementation1.ControlCategoryName,
						ControlCategoryCatalogId: orchestratortest.MockControlImplementation1.ControlCategoryCatalogId,
					},
				},
				context: auth.WithClaims(context.Background(), &auth.OAuthClaims{
					IsAdminToken: true,
				}),
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					assert.NoError(t, d.Create(orchestratortest.MockAuditScope1))
				}),
				authz: &service.AuthorizationStrategyPermissionStore{
					Permissions: permissionStore{
						db: persistencetest.NewInMemoryDB(t, types, joinTables),
					},
				},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ControlImplementation], args ...any) bool {
				return assert.NotNil(t, got.Msg) &&
					assert.NotEmpty(t, got.Msg.Id)
			},
			wantErr: assert.NoError,
			wantDB:  assert.NotNil[persistence.DB],
		},
		{
			name: "validation error - empty request",
			args: args{
				req: &orchestrator.CreateControlImplementationRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.ControlImplementation]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument)
			},
			wantDB: assert.NotNil[persistence.DB],
		},
		{
			name: "authorization failure",
			args: args{
				req: &orchestrator.CreateControlImplementationRequest{
					ControlImplementation: &orchestrator.ControlImplementation{
						AuditScopeId:             orchestratortest.MockControlImplementation1.AuditScopeId,
						TargetOfEvaluationId:     orchestratortest.MockControlImplementation1.TargetOfEvaluationId,
						ControlId:                orchestratortest.MockControlImplementation1.ControlId,
						ControlCategoryName:      orchestratortest.MockControlImplementation1.ControlCategoryName,
						ControlCategoryCatalogId: orchestratortest.MockControlImplementation1.ControlCategoryCatalogId,
					},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					assert.NoError(t, d.Create(orchestratortest.MockAuditScope1))
				}),
				authz: &denyAuthorizationStrategy{},
			},
			want: assert.Nil[*connect.Response[orchestrator.ControlImplementation]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodePermissionDenied)
			},
			wantDB: assert.NotNil[persistence.DB],
		},
		{
			name: "audit scope not found",
			args: args{
				req: &orchestrator.CreateControlImplementationRequest{
					ControlImplementation: &orchestrator.ControlImplementation{
						AuditScopeId:             orchestratortest.MockNonExistentId,
						TargetOfEvaluationId:     orchestratortest.MockControlImplementation1.TargetOfEvaluationId,
						ControlId:                orchestratortest.MockControlImplementation1.ControlId,
						ControlCategoryName:      orchestratortest.MockControlImplementation1.ControlCategoryName,
						ControlCategoryCatalogId: orchestratortest.MockControlImplementation1.ControlCategoryCatalogId,
					},
				},
			},
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: assert.Nil[*connect.Response[orchestrator.ControlImplementation]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeNotFound)
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
			res, err := svc.CreateControlImplementation(tt.args.context, connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
			tt.wantDB(t, tt.fields.db, res)
		})
	}
}

func TestService_GetControlImplementation(t *testing.T) {
	type args struct {
		req     *orchestrator.GetControlImplementationRequest
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
		want    assert.Want[*connect.Response[orchestrator.ControlImplementation]]
		wantErr assert.WantErr
	}{
		{
			name: "happy path",
			args: args{
				req: &orchestrator.GetControlImplementationRequest{
					Id: orchestratortest.MockControlImplementation1.Id,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					assert.NoError(t, d.Create(orchestratortest.MockControlImplementation1))
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ControlImplementation], args ...any) bool {
				return assert.NotNil(t, got.Msg) &&
					assert.Equal(t, orchestratortest.MockControlImplementation1.Id, got.Msg.Id)
			},
			wantErr: assert.NoError,
		},
		{
			name: "authorization failure",
			args: args{
				req: &orchestrator.GetControlImplementationRequest{
					Id: orchestratortest.MockControlImplementation1.Id,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					assert.NoError(t, d.Create(orchestratortest.MockControlImplementation1))
				}),
				authz: &denyAuthorizationStrategy{},
			},
			want: assert.Nil[*connect.Response[orchestrator.ControlImplementation]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodePermissionDenied)
			},
		},
		{
			name: "not found",
			args: args{
				req: &orchestrator.GetControlImplementationRequest{
					Id: orchestratortest.MockNonExistentId,
				},
			},
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: assert.Nil[*connect.Response[orchestrator.ControlImplementation]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeNotFound)
			},
		},
		{
			name: "validation error - empty id",
			args: args{
				req: &orchestrator.GetControlImplementationRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.ControlImplementation]],
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
			res, err := svc.GetControlImplementation(tt.args.context, connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}

func TestService_ListControlImplementations(t *testing.T) {
	type args struct {
		req     *orchestrator.ListControlImplementationsRequest
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
		want    assert.Want[*connect.Response[orchestrator.ListControlImplementationsResponse]]
		wantErr assert.WantErr
	}{
		{
			name: "happy path: list all",
			args: args{
				req: &orchestrator.ListControlImplementationsRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					assert.NoError(t, d.Create(orchestratortest.MockControlImplementation1))
					assert.NoError(t, d.Create(orchestratortest.MockControlImplementation2))
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ListControlImplementationsResponse], args ...any) bool {
				return assert.NotNil(t, got.Msg) &&
					assert.Equal(t, 2, len(got.Msg.ControlImplementations))
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: filter by audit scope",
			args: args{
				req: &orchestrator.ListControlImplementationsRequest{
					Filter: &orchestrator.ListControlImplementationsRequest_Filter{
						AuditScopeId: &orchestratortest.MockControlImplementation1.AuditScopeId,
					},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					assert.NoError(t, d.Create(orchestratortest.MockControlImplementation1))
					assert.NoError(t, d.Create(orchestratortest.MockControlImplementation2))
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ListControlImplementationsResponse], args ...any) bool {
				return assert.NotNil(t, got.Msg) &&
					assert.Equal(t, 1, len(got.Msg.ControlImplementations)) &&
					assert.Equal(t, orchestratortest.MockControlImplementation1.Id, got.Msg.ControlImplementations[0].Id)
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: no access returns empty list",
			args: args{
				req: &orchestrator.ListControlImplementationsRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					assert.NoError(t, d.Create(orchestratortest.MockControlImplementation1))
				}),
				authz: &denyAuthorizationStrategy{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ListControlImplementationsResponse], args ...any) bool {
				return assert.NotNil(t, got.Msg) &&
					assert.Equal(t, 0, len(got.Msg.ControlImplementations))
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
			res, err := svc.ListControlImplementations(tt.args.context, connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}

func TestService_UpdateControlImplementation(t *testing.T) {
	var (
		assigneeId = orchestratortest.MockUserId1
	)
	type args struct {
		req     *orchestrator.UpdateControlImplementationRequest
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
		want    assert.Want[*connect.Response[orchestrator.ControlImplementation]]
		wantErr assert.WantErr
		wantDB  assert.Want[persistence.DB]
	}{
		{
			name: "happy path: update assignee",
			args: args{
				req: &orchestrator.UpdateControlImplementationRequest{
					Id:         orchestratortest.MockControlImplementation1.Id,
					AssigneeId: &assigneeId,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					assert.NoError(t, d.Create(orchestratortest.MockControlImplementation1))
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ControlImplementation], args ...any) bool {
				return assert.NotNil(t, got.Msg) &&
					assert.Equal(t, &assigneeId, got.Msg.AssigneeId)
			},
			wantErr: assert.NoError,
			wantDB: func(t *testing.T, db persistence.DB, msgAndArgs ...any) bool {
				got := assert.InDB[orchestrator.ControlImplementation](t, db, orchestratortest.MockControlImplementation1.Id)
				return assert.Equal(t, &assigneeId, got.AssigneeId)
			},
		},
		{
			name: "authorization failure",
			args: args{
				req: &orchestrator.UpdateControlImplementationRequest{
					Id: orchestratortest.MockControlImplementation1.Id,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					assert.NoError(t, d.Create(orchestratortest.MockControlImplementation1))
				}),
				authz: &denyAuthorizationStrategy{},
			},
			want: assert.Nil[*connect.Response[orchestrator.ControlImplementation]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodePermissionDenied)
			},
			wantDB: assert.NotNil[persistence.DB],
		},
		{
			name: "not found",
			args: args{
				req: &orchestrator.UpdateControlImplementationRequest{
					Id: orchestratortest.MockNonExistentId,
				},
			},
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: assert.Nil[*connect.Response[orchestrator.ControlImplementation]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeNotFound)
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
			res, err := svc.UpdateControlImplementation(tt.args.context, connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
			tt.wantDB(t, tt.fields.db, res)
		})
	}
}

func TestService_TransitionControlImplementationState(t *testing.T) {
	type args struct {
		req     *orchestrator.TransitionControlImplementationStateRequest
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
		want    assert.Want[*connect.Response[orchestrator.ControlImplementation]]
		wantErr assert.WantErr
		wantDB  assert.Want[persistence.DB]
	}{
		{
			name: "happy path: OPEN -> IN_PROGRESS",
			args: args{
				req: &orchestrator.TransitionControlImplementationStateRequest{
					Id:      orchestratortest.MockControlImplementation1.Id,
					ToState: orchestrator.ControlImplementationState_CONTROL_IMPLEMENTATION_STATE_IN_PROGRESS,
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
					assert.NoError(t, d.Create(orchestratortest.MockControlImplementation1))
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ControlImplementation], args ...any) bool {
				return assert.NotNil(t, got.Msg) &&
					assert.Equal(t,
						orchestrator.ControlImplementationState_CONTROL_IMPLEMENTATION_STATE_IN_PROGRESS,
						got.Msg.State) &&
					assert.Equal(t, 1, len(got.Msg.Transitions))
			},
			wantErr: assert.NoError,
			wantDB: func(t *testing.T, db persistence.DB, msgAndArgs ...any) bool {
				got := assert.InDB[orchestrator.ControlImplementation](t, db, orchestratortest.MockControlImplementation1.Id)
				return assert.Equal(t,
					orchestrator.ControlImplementationState_CONTROL_IMPLEMENTATION_STATE_IN_PROGRESS,
					got.State)
			},
		},
		{
			name: "invalid transition: OPEN -> ACCEPTED",
			args: args{
				req: &orchestrator.TransitionControlImplementationStateRequest{
					Id:      orchestratortest.MockControlImplementation1.Id,
					ToState: orchestrator.ControlImplementationState_CONTROL_IMPLEMENTATION_STATE_ACCEPTED,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					assert.NoError(t, d.Create(orchestratortest.MockControlImplementation1))
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: assert.Nil[*connect.Response[orchestrator.ControlImplementation]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument)
			},
			wantDB: assert.NotNil[persistence.DB],
		},
		{
			name: "authorization failure",
			args: args{
				req: &orchestrator.TransitionControlImplementationStateRequest{
					Id:      orchestratortest.MockControlImplementation1.Id,
					ToState: orchestrator.ControlImplementationState_CONTROL_IMPLEMENTATION_STATE_IN_PROGRESS,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					assert.NoError(t, d.Create(orchestratortest.MockControlImplementation1))
				}),
				authz: &denyAuthorizationStrategy{},
			},
			want: assert.Nil[*connect.Response[orchestrator.ControlImplementation]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodePermissionDenied)
			},
			wantDB: assert.NotNil[persistence.DB],
		},
		{
			name: "not found",
			args: args{
				req: &orchestrator.TransitionControlImplementationStateRequest{
					Id:      orchestratortest.MockNonExistentId,
					ToState: orchestrator.ControlImplementationState_CONTROL_IMPLEMENTATION_STATE_IN_PROGRESS,
				},
			},
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: assert.Nil[*connect.Response[orchestrator.ControlImplementation]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeNotFound)
			},
			wantDB: assert.NotNil[persistence.DB],
		},
		{
			name: "validation error - unspecified target state",
			args: args{
				req: &orchestrator.TransitionControlImplementationStateRequest{
					Id:      orchestratortest.MockControlImplementation1.Id,
					ToState: orchestrator.ControlImplementationState_CONTROL_IMPLEMENTATION_STATE_UNSPECIFIED,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.ControlImplementation]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument)
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
			res, err := svc.TransitionControlImplementationState(tt.args.context, connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
			tt.wantDB(t, tt.fields.db, res)
		})
	}
}

func TestService_RemoveControlImplementation(t *testing.T) {
	type args struct {
		req     *orchestrator.RemoveControlImplementationRequest
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
		wantDB  assert.Want[persistence.DB]
	}{
		{
			name: "happy path",
			args: args{
				req: &orchestrator.RemoveControlImplementationRequest{
					Id: orchestratortest.MockControlImplementation1.Id,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					assert.NoError(t, d.Create(orchestratortest.MockControlImplementation1))
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: func(t *testing.T, got *connect.Response[emptypb.Empty], args ...any) bool {
				return assert.NotNil(t, got.Msg)
			},
			wantErr: assert.NoError,
			wantDB: func(t *testing.T, db persistence.DB, msgAndArgs ...any) bool {
				var impl orchestrator.ControlImplementation
				err := db.Get(&impl, "id = ?", orchestratortest.MockControlImplementation1.Id)
				return assert.ErrorIs(t, err, persistence.ErrRecordNotFound)
			},
		},
		{
			name: "authorization failure",
			args: args{
				req: &orchestrator.RemoveControlImplementationRequest{
					Id: orchestratortest.MockControlImplementation1.Id,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					assert.NoError(t, d.Create(orchestratortest.MockControlImplementation1))
				}),
				authz: &denyAuthorizationStrategy{},
			},
			want: assert.Nil[*connect.Response[emptypb.Empty]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodePermissionDenied)
			},
			wantDB: assert.NotNil[persistence.DB],
		},
		{
			name: "not found",
			args: args{
				req: &orchestrator.RemoveControlImplementationRequest{
					Id: orchestratortest.MockNonExistentId,
				},
			},
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: assert.Nil[*connect.Response[emptypb.Empty]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeNotFound)
			},
			wantDB: assert.NotNil[persistence.DB],
		},
		{
			name: "validation error - empty id",
			args: args{
				req: &orchestrator.RemoveControlImplementationRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[emptypb.Empty]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument)
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
			res, err := svc.RemoveControlImplementation(tt.args.context, connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
			tt.wantDB(t, tt.fields.db, res)
		})
	}
}
