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

func TestService_CreateControlInScope(t *testing.T) {
	type args struct {
		req     *orchestrator.CreateControlInScopeRequest
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
		want    assert.Want[*connect.Response[orchestrator.ControlInScope]]
		wantErr assert.WantErr
		wantDB  assert.Want[persistence.DB]
	}{
		{
			name: "happy path: with allow-all authorization strategy",
			args: args{
				req: &orchestrator.CreateControlInScopeRequest{
					AuditScopeId:         orchestratortest.MockControlInScope1.AuditScopeId,
					ControlId:            orchestratortest.MockControlInScope1.ControlId,
					TargetOfEvaluationId: orchestratortest.MockControlInScope1.TargetOfEvaluationId,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					assert.NoError(t, d.Create(orchestratortest.MockAuditScope1))
					seedControl(t, d, orchestratortest.MockControl1)
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ControlInScope], args ...any) bool {
				return assert.NotNil(t, got.Msg) &&
					assert.NotEmpty(t, got.Msg.Id) &&
					assert.Equal(t,
						orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_OPEN,
						got.Msg.State)
			},
			wantErr: assert.NoError,
			wantDB: func(t *testing.T, db persistence.DB, msgAndArgs ...any) bool {
				res := assert.Is[*connect.Response[orchestrator.ControlInScope]](t, msgAndArgs[0])
				assert.NotNil(t, res)

				got := assert.InDB[orchestrator.ControlInScope](t, db, res.Msg.Id)
				want := &orchestrator.ControlInScope{
					AuditScopeId:         orchestratortest.MockControlInScope1.AuditScopeId,
					ControlId:            orchestratortest.MockControlInScope1.ControlId,
					TargetOfEvaluationId: orchestratortest.MockControlInScope1.TargetOfEvaluationId,
					State:                orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_OPEN,
				}
				if !assert.Equal(t, want, got, protocmp.IgnoreFields(&orchestrator.ControlInScope{},
					"id", "created_at", "updated_at")) {
					return false
				}

				// Verify AuditTrailEvent was created for the scoping.
				var count int64
				count, _ = db.Count(&orchestrator.AuditTrailEvent{},
					"audit_scope_id = ?", orchestratortest.MockControlInScope1.AuditScopeId)
				return assert.Equal(t, int64(1), count)
			},
		},
		{
			name: "happy path: with permission store and admin token",
			args: args{
				req: &orchestrator.CreateControlInScopeRequest{
					AuditScopeId:         orchestratortest.MockControlInScope1.AuditScopeId,
					ControlId:            orchestratortest.MockControlInScope1.ControlId,
					TargetOfEvaluationId: orchestratortest.MockControlInScope1.TargetOfEvaluationId,
				},
				context: auth.WithClaims(context.Background(), &auth.OAuthClaims{
					IsAdminToken: true,
				}),
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					assert.NoError(t, d.Create(orchestratortest.MockAuditScope1))
					seedControl(t, d, orchestratortest.MockControl1)
				}),
				authz: &service.AuthorizationStrategyPermissionStore{
					Permissions: service.DBPermissionStore{
						DB: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
							assert.NoError(t, d.Create(orchestratortest.MockUserPermissionsToEContributor))
						}),
					},
				},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ControlInScope], args ...any) bool {
				want := &orchestrator.ControlInScope{
					AuditScopeId: orchestratortest.MockControlInScope1.AuditScopeId,
					ControlId:    orchestratortest.MockControlInScope1.ControlId,
					State:        orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_OPEN,
				}
				return assert.NotNil(t, got.Msg) &&
					assert.Equal(t, want, got.Msg, protocmp.IgnoreFields(&orchestrator.ControlInScope{},
						"id", "target_of_evaluation_id", "created_at", "updated_at"))
			},
			wantErr: assert.NoError,
			wantDB:  assert.NotNil[persistence.DB],
		},
		{
			name: "validation error - empty request",
			args: args{
				req: &orchestrator.CreateControlInScopeRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.ControlInScope]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument)
			},
			wantDB: assert.NotNil[persistence.DB],
		},
		{
			name: "authorization failure",
			args: args{
				req: &orchestrator.CreateControlInScopeRequest{
					AuditScopeId:         orchestratortest.MockControlInScope1.AuditScopeId,
					ControlId:            orchestratortest.MockControlInScope1.ControlId,
					TargetOfEvaluationId: orchestratortest.MockControlInScope1.TargetOfEvaluationId,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					assert.NoError(t, d.Create(orchestratortest.MockAuditScope1))
				}),
				authz: &denyAuthorizationStrategy{},
			},
			want: assert.Nil[*connect.Response[orchestrator.ControlInScope]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodePermissionDenied)
			},
			wantDB: assert.NotNil[persistence.DB],
		},
		{
			name: "duplicate: already in scope",
			args: args{
				req: &orchestrator.CreateControlInScopeRequest{
					AuditScopeId:         orchestratortest.MockControlInScope1.AuditScopeId,
					ControlId:            orchestratortest.MockControlInScope1.ControlId,
					TargetOfEvaluationId: orchestratortest.MockControlInScope1.TargetOfEvaluationId,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					assert.NoError(t, d.Create(orchestratortest.MockAuditScope1))
					seedControl(t, d, orchestratortest.MockControl1)
					assert.NoError(t, d.Create(orchestratortest.MockControlInScope1))
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: assert.Nil[*connect.Response[orchestrator.ControlInScope]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeAlreadyExists)
			},
			wantDB: assert.NotNil[persistence.DB],
		},
		{
			name: "control not in catalog",
			args: args{
				req: &orchestrator.CreateControlInScopeRequest{
					AuditScopeId:         orchestratortest.MockControlInScope1.AuditScopeId,
					ControlId:            orchestratortest.MockNonExistentId,
					TargetOfEvaluationId: orchestratortest.MockControlInScope1.TargetOfEvaluationId,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					assert.NoError(t, d.Create(orchestratortest.MockAuditScope1))
					seedControl(t, d, orchestratortest.MockControl1)
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: assert.Nil[*connect.Response[orchestrator.ControlInScope]],
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
			res, err := svc.CreateControlInScope(tt.args.context, connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
			tt.wantDB(t, tt.fields.db, res)
		})
	}
}

// seedControl inserts a Control record into the DB so that CreateControlInScope's existence check passes.
func seedControl(t *testing.T, d persistence.DB, ctrl *orchestrator.Control) {
	t.Helper()
	assert.NoError(t, d.Create(ctrl))
}

// seedControlInScope1 seeds the FK-required parents and then MockControlInScope1.
func seedControlInScope1(t *testing.T, d persistence.DB) {
	t.Helper()
	assert.NoError(t, d.Create(orchestratortest.MockAuditScope1))
	seedControl(t, d, orchestratortest.MockControl1)
	assert.NoError(t, d.Create(orchestratortest.MockControlInScope1))
}

// seedControlInScope2 seeds the FK-required parents and then MockControlInScope2.
func seedControlInScope2(t *testing.T, d persistence.DB) {
	t.Helper()
	assert.NoError(t, d.Create(orchestratortest.MockAuditScope2))
	seedControl(t, d, orchestratortest.MockControl2)
	assert.NoError(t, d.Create(orchestratortest.MockControlInScope2))
}

func TestService_GetControlInScope(t *testing.T) {
	type args struct {
		req     *orchestrator.GetControlInScopeRequest
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
		want    assert.Want[*connect.Response[orchestrator.ControlInScope]]
		wantErr assert.WantErr
	}{
		{
			name: "happy path",
			args: args{
				req: &orchestrator.GetControlInScopeRequest{
					Id: orchestratortest.MockControlInScope1.Id,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					assert.NoError(t, d.Create(orchestratortest.MockAuditScope1))
					seedControl(t, d, orchestratortest.MockControl1)
					assert.NoError(t, d.Create(orchestratortest.MockControlInScope1))
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ControlInScope], args ...any) bool {
				return assert.NotNil(t, got.Msg) &&
					assert.Equal(t, orchestratortest.MockControlInScope1.Id, got.Msg.Id)
			},
			wantErr: assert.NoError,
		},
		{
			name: "authorization failure",
			args: args{
				req: &orchestrator.GetControlInScopeRequest{
					Id: orchestratortest.MockControlInScope1.Id,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					assert.NoError(t, d.Create(orchestratortest.MockAuditScope1))
					seedControl(t, d, orchestratortest.MockControl1)
					assert.NoError(t, d.Create(orchestratortest.MockControlInScope1))
				}),
				authz: &denyAuthorizationStrategy{},
			},
			want: assert.Nil[*connect.Response[orchestrator.ControlInScope]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodePermissionDenied)
			},
		},
		{
			name: "not found",
			args: args{
				req: &orchestrator.GetControlInScopeRequest{
					Id: orchestratortest.MockNonExistentId,
				},
			},
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: assert.Nil[*connect.Response[orchestrator.ControlInScope]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeNotFound)
			},
		},
		{
			name: "validation error - empty id",
			args: args{
				req: &orchestrator.GetControlInScopeRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.ControlInScope]],
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
			res, err := svc.GetControlInScope(tt.args.context, connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}

func TestService_ListControlsInScope(t *testing.T) {
	type args struct {
		req     *orchestrator.ListControlsInScopeRequest
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
		want    assert.Want[*connect.Response[orchestrator.ListControlsInScopeResponse]]
		wantErr assert.WantErr
	}{
		{
			name: "happy path: list IsAdminToken=true",
			args: args{
				req: &orchestrator.ListControlsInScopeRequest{},
				context: auth.WithClaims(context.Background(), &auth.OAuthClaims{
					IsAdminToken: true,
					RegisteredClaims: jwt.RegisteredClaims{
						Subject: orchestratortest.MockUserId1,
						Issuer:  orchestratortest.MockUserIssuer1,
					},
				}),
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					seedControlInScope1(t, d)
					seedControlInScope2(t, d)
				}),
				authz: &service.AuthorizationStrategyPermissionStore{
					Permissions: service.DBPermissionStore{
						DB: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
							err := d.Create(orchestratortest.MockUserPermissionsAuditScopeAdmin)
							assert.NoError(t, err)
						}),
					},
				},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ListControlsInScopeResponse], args ...any) bool {
				return assert.NotNil(t, got.Msg) &&
					assert.Equal(t, 2, len(got.Msg.ControlsInScope))
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: list all: cfadmin=false and permission for audit_scope_1",
			args: args{
				req: &orchestrator.ListControlsInScopeRequest{},
				context: auth.WithClaims(context.Background(), &auth.OAuthClaims{
					IsAdminToken: false,
					RegisteredClaims: jwt.RegisteredClaims{
						Subject: orchestratortest.MockUserId1,
						Issuer:  orchestratortest.MockUserIssuer1,
					},
				}),
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					seedControlInScope1(t, d)
					seedControlInScope2(t, d)
				}),
				authz: &service.AuthorizationStrategyPermissionStore{
					Permissions: service.DBPermissionStore{
						DB: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
							err := d.Create(orchestratortest.MockUserPermissionsAuditScopeAdmin)
							assert.NoError(t, err)
						}),
					},
				},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ListControlsInScopeResponse], args ...any) bool {
				return assert.NotNil(t, got.Msg) &&
					assert.Equal(t, 1, len(got.Msg.ControlsInScope)) &&
					assert.Equal(t, orchestratortest.MockControlInScope1, got.Msg.ControlsInScope[0])
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: list all: cfadmin=false,permission for audit_scope_1 and filter by audit_scope",
			args: args{
				req: &orchestrator.ListControlsInScopeRequest{
					Filter: &orchestrator.ListControlsInScopeRequest_Filter{
						AuditScopeId: new(orchestratortest.MockScopeId1),
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
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					seedControlInScope1(t, d)
					seedControlInScope2(t, d)
				}),
				authz: &service.AuthorizationStrategyPermissionStore{
					Permissions: service.DBPermissionStore{
						DB: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
							err := d.Create(orchestratortest.MockUserPermissionsAuditScopeAdmin)
							assert.NoError(t, err)
						}),
					},
				},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ListControlsInScopeResponse], args ...any) bool {
				return assert.NotNil(t, got.Msg) &&
					assert.Equal(t, 1, len(got.Msg.ControlsInScope)) &&
					assert.Equal(t, orchestratortest.MockControlInScope1, got.Msg.ControlsInScope[0])
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: no access returns empty list",
			args: args{
				req: &orchestrator.ListControlsInScopeRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					seedControlInScope1(t, d)
				}),
				authz: &denyAuthorizationStrategy{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ListControlsInScopeResponse], args ...any) bool {
				return assert.NotNil(t, got.Msg) &&
					assert.Equal(t, 0, len(got.Msg.ControlsInScope))
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
			res, err := svc.ListControlsInScope(tt.args.context, connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}

func TestService_UpdateControlInScope(t *testing.T) {
	var (
		assigneeId = orchestratortest.MockUserId1
	)
	type args struct {
		req     *orchestrator.UpdateControlInScopeRequest
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
		want    assert.Want[*connect.Response[orchestrator.ControlInScope]]
		wantErr assert.WantErr
		wantDB  assert.Want[persistence.DB]
	}{
		{
			name: "happy path: update assignee",
			args: args{
				req: &orchestrator.UpdateControlInScopeRequest{
					Id:         orchestratortest.MockControlInScope1.Id,
					AssigneeId: &assigneeId,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					seedControlInScope1(t, d)
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ControlInScope], args ...any) bool {
				return assert.NotNil(t, got.Msg) &&
					assert.Equal(t, &assigneeId, got.Msg.AssigneeId)
			},
			wantErr: assert.NoError,
			wantDB: func(t *testing.T, db persistence.DB, msgAndArgs ...any) bool {
				got := assert.InDB[orchestrator.ControlInScope](t, db, orchestratortest.MockControlInScope1.Id)
				return assert.Equal(t, &assigneeId, got.AssigneeId)
			},
		},
		{
			name: "happy path: update implementation details",
			args: args{
				req: func() *orchestrator.UpdateControlInScopeRequest {
					details := "We use TLS 1.3 for all connections."
					return &orchestrator.UpdateControlInScopeRequest{
						Id:                    orchestratortest.MockControlInScope1.Id,
						ImplementationDetails: &details,
					}
				}(),
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					seedControlInScope1(t, d)
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ControlInScope], args ...any) bool {
				details := "We use TLS 1.3 for all connections."
				return assert.NotNil(t, got.Msg) &&
					assert.Equal(t, &details, got.Msg.ImplementationDetails)
			},
			wantErr: assert.NoError,
			wantDB: func(t *testing.T, db persistence.DB, msgAndArgs ...any) bool {
				details := "We use TLS 1.3 for all connections."
				got := assert.InDB[orchestrator.ControlInScope](t, db, orchestratortest.MockControlInScope1.Id)
				return assert.Equal(t, &details, got.ImplementationDetails)
			},
		},
		{
			name: "authorization failure",
			args: args{
				req: &orchestrator.UpdateControlInScopeRequest{
					Id: orchestratortest.MockControlInScope1.Id,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					seedControlInScope1(t, d)
				}),
				authz: &denyAuthorizationStrategy{},
			},
			want: assert.Nil[*connect.Response[orchestrator.ControlInScope]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodePermissionDenied)
			},
			wantDB: assert.NotNil[persistence.DB],
		},
		{
			name: "not found",
			args: args{
				req: &orchestrator.UpdateControlInScopeRequest{
					Id: orchestratortest.MockNonExistentId,
				},
			},
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: assert.Nil[*connect.Response[orchestrator.ControlInScope]],
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
			res, err := svc.UpdateControlInScope(tt.args.context, connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
			tt.wantDB(t, tt.fields.db, res)
		})
	}
}

func TestService_TransitionControlInScopeState(t *testing.T) {
	type args struct {
		req     *orchestrator.TransitionControlInScopeStateRequest
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
		want    assert.Want[*connect.Response[orchestrator.ControlInScope]]
		wantErr assert.WantErr
		wantDB  assert.Want[persistence.DB]
	}{
		{
			name: "happy path: OPEN -> IN_PROGRESS creates AuditTrailEvent",
			args: args{
				req: &orchestrator.TransitionControlInScopeStateRequest{
					Id:      orchestratortest.MockControlInScope1.Id,
					ToState: orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_IN_PROGRESS,
					Comment: "Starting implementation work.",
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					seedControlInScope1(t, d)
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ControlInScope], args ...any) bool {
				return assert.NotNil(t, got.Msg) &&
					assert.Equal(t,
						orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_IN_PROGRESS,
						got.Msg.State)
			},
			wantErr: assert.NoError,
			wantDB: func(t *testing.T, db persistence.DB, msgAndArgs ...any) bool {
				// Verify AuditTrailEvent was created.
				var count int64
				count, _ = db.Count(&orchestrator.AuditTrailEvent{},
					"audit_scope_id = ?", orchestratortest.MockControlInScope1.AuditScopeId)
				return assert.Equal(t, int64(1), count)
			},
		},
		{
			name: "happy path: OPEN -> IN_PROGRESS with actor",
			args: args{
				req: &orchestrator.TransitionControlInScopeStateRequest{
					Id:      orchestratortest.MockControlInScope1.Id,
					ToState: orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_IN_PROGRESS,
					Comment: "Picked up by the team.",
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
					seedControlInScope1(t, d)
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ControlInScope], args ...any) bool {
				return assert.NotNil(t, got.Msg) &&
					assert.Equal(t,
						orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_IN_PROGRESS,
						got.Msg.State)
			},
			wantErr: assert.NoError,
			wantDB: func(t *testing.T, db persistence.DB, msgAndArgs ...any) bool {
				got := assert.InDB[orchestrator.ControlInScope](t, db, orchestratortest.MockControlInScope1.Id)
				return assert.Equal(t,
					orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_IN_PROGRESS,
					got.State)
			},
		},
		{
			name: "invalid transition: OPEN -> ACCEPTED",
			args: args{
				req: &orchestrator.TransitionControlInScopeStateRequest{
					Id:      orchestratortest.MockControlInScope1.Id,
					ToState: orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_ACCEPTED,
					Comment: "Marking as accepted.",
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					seedControlInScope1(t, d)
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: assert.Nil[*connect.Response[orchestrator.ControlInScope]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument)
			},
			wantDB: assert.NotNil[persistence.DB],
		},
		{
			name: "authorization failure",
			args: args{
				req: &orchestrator.TransitionControlInScopeStateRequest{
					Id:      orchestratortest.MockControlInScope1.Id,
					ToState: orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_IN_PROGRESS,
					Comment: "Picking up.",
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					seedControlInScope1(t, d)
				}),
				authz: &denyAuthorizationStrategy{},
			},
			want: assert.Nil[*connect.Response[orchestrator.ControlInScope]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodePermissionDenied)
			},
			wantDB: assert.NotNil[persistence.DB],
		},
		{
			name: "not found",
			args: args{
				req: &orchestrator.TransitionControlInScopeStateRequest{
					Id:      orchestratortest.MockNonExistentId,
					ToState: orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_IN_PROGRESS,
					Comment: "Picking up.",
				},
			},
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: assert.Nil[*connect.Response[orchestrator.ControlInScope]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeNotFound)
			},
			wantDB: assert.NotNil[persistence.DB],
		},
		{
			name: "validation error - unspecified target state",
			args: args{
				req: &orchestrator.TransitionControlInScopeStateRequest{
					Id:      orchestratortest.MockControlInScope1.Id,
					ToState: orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_UNSPECIFIED,
					Comment: "Picking up.",
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.ControlInScope]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument)
			},
			wantDB: assert.NotNil[persistence.DB],
		},
		{
			name: "validation error - missing comment",
			args: args{
				req: &orchestrator.TransitionControlInScopeStateRequest{
					Id:      orchestratortest.MockControlInScope1.Id,
					ToState: orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_IN_PROGRESS,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.ControlInScope]],
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
			res, err := svc.TransitionControlInScopeState(tt.args.context, connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
			tt.wantDB(t, tt.fields.db, res)
		})
	}
}

func TestService_RemoveControlInScope(t *testing.T) {
	type args struct {
		req     *orchestrator.RemoveControlInScopeRequest
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
				req: &orchestrator.RemoveControlInScopeRequest{
					Id: orchestratortest.MockControlInScope1.Id,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					seedControlInScope1(t, d)
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: func(t *testing.T, got *connect.Response[emptypb.Empty], args ...any) bool {
				return assert.NotNil(t, got.Msg)
			},
			wantErr: assert.NoError,
			wantDB: func(t *testing.T, db persistence.DB, msgAndArgs ...any) bool {
				// ControlInScope should be gone.
				var cis orchestrator.ControlInScope
				err := db.Get(&cis, "id = ?", orchestratortest.MockControlInScope1.Id)
				if !assert.ErrorIs(t, err, persistence.ErrRecordNotFound) {
					return false
				}
				// AuditTrailEvent should have been created.
				var count int64
				count, _ = db.Count(&orchestrator.AuditTrailEvent{},
					"audit_scope_id = ?", orchestratortest.MockControlInScope1.AuditScopeId)
				return assert.Equal(t, int64(1), count)
			},
		},
		{
			name: "authorization failure",
			args: args{
				req: &orchestrator.RemoveControlInScopeRequest{
					Id: orchestratortest.MockControlInScope1.Id,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					seedControlInScope1(t, d)
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
				req: &orchestrator.RemoveControlInScopeRequest{
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
				req: &orchestrator.RemoveControlInScopeRequest{},
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
			res, err := svc.RemoveControlInScope(tt.args.context, connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
			tt.wantDB(t, tt.fields.db, res)
		})
	}
}

// TestService_RemoveControlInScope_CascadesToDescendants verifies that
// removing a parent ControlInScope also removes the ControlInScope record of
// every descendant control in the same audit scope, and that one audit-trail
// event is recorded per removed record. Without this cascade a sub-control
// could remain "in scope" after its parent was taken out.
func TestService_RemoveControlInScope_CascadesToDescendants(t *testing.T) {
	const subControlInScopeId = "00000000-0000-0000-0004-000000000099"

	db := persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
		// Seed audit scope, parent control (which auto-cascades the
		// MockSubControl1 row in via gorm), and a ControlInScope record for
		// each of them under that audit scope.
		assert.NoError(t, d.Create(orchestratortest.MockAuditScope1))
		seedControl(t, d, orchestratortest.MockControl1)
		assert.NoError(t, d.Create(orchestratortest.MockControlInScope1))
		assert.NoError(t, d.Create(&orchestrator.ControlInScope{
			Id:                   subControlInScopeId,
			AuditScopeId:         orchestratortest.MockScopeId1,
			TargetOfEvaluationId: orchestratortest.MockToeId1,
			ControlId:            orchestratortest.MockSubControlId1,
			State:                orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_OPEN,
		}))
	})

	svc := &Service{db: db, authz: &service.AuthorizationStrategyAllowAll{}}
	_, err := svc.RemoveControlInScope(context.Background(), connect.NewRequest(
		&orchestrator.RemoveControlInScopeRequest{Id: orchestratortest.MockControlInScopeId1},
	))
	assert.NoError(t, err)

	// Both records should be gone.
	var parent, child orchestrator.ControlInScope
	assert.ErrorIs(t, db.Get(&parent, "id = ?", orchestratortest.MockControlInScopeId1), persistence.ErrRecordNotFound)
	assert.ErrorIs(t, db.Get(&child, "id = ?", subControlInScopeId), persistence.ErrRecordNotFound)

	// One audit-trail event per removed ControlInScope record. We can't filter
	// here because of a ramsql quirk on the in-memory test backend that drops
	// rows when WHERE is used on this table; the test seeds no other events, so
	// listing them all is equivalent.
	var events []*orchestrator.AuditTrailEvent
	assert.NoError(t, db.List(&events, "id", true, 0, -1))
	assert.Equal(t, 2, len(events))
	scopingControlIds := map[string]struct{}{}
	for _, e := range events {
		assert.Equal(t, orchestratortest.MockScopeId1, e.AuditScopeId)
		var payload orchestrator.ControlScopingEvent
		assert.NoError(t, e.EventData.UnmarshalTo(&payload))
		assert.False(t, payload.InScope)
		scopingControlIds[payload.ControlId] = struct{}{}
	}
	assert.Contains(t, scopingControlIds, orchestratortest.MockControlId1)
	assert.Contains(t, scopingControlIds, orchestratortest.MockSubControlId1)
}

func TestService_ListAuditTrailEvents(t *testing.T) {
	type args struct {
		req     *orchestrator.ListAuditTrailEventsRequest
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
		want    assert.Want[*connect.Response[orchestrator.ListAuditTrailEventsResponse]]
		wantErr assert.WantErr
	}{
		{
			name: "happy path: list all",
			args: args{
				req: &orchestrator.ListAuditTrailEventsRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					seedControlInScope1(t, d)
					assert.NoError(t, d.Create(&orchestrator.AuditTrailEvent{
						Id:           "00000000-0000-0000-0005-000000000001",
						AuditScopeId: orchestratortest.MockScopeId1,
						ActorId:      orchestratortest.MockUserId1,
					}))
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ListAuditTrailEventsResponse], args ...any) bool {
				return assert.NotNil(t, got.Msg) &&
					assert.Equal(t, 1, len(got.Msg.AuditTrailEvents))
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: filter by audit scope",
			args: args{
				req: func() *orchestrator.ListAuditTrailEventsRequest {
					scopeId := orchestratortest.MockScopeId1
					return &orchestrator.ListAuditTrailEventsRequest{
						Filter: &orchestrator.ListAuditTrailEventsRequest_Filter{
							AuditScopeId: &scopeId,
						},
					}
				}(),
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					assert.NoError(t, d.Create(orchestratortest.MockAuditScope1))
					assert.NoError(t, d.Create(orchestratortest.MockAuditScope2))
					assert.NoError(t, d.Create(&orchestrator.AuditTrailEvent{
						Id:           "00000000-0000-0000-0005-000000000001",
						AuditScopeId: orchestratortest.MockScopeId1,
					}))
					assert.NoError(t, d.Create(&orchestrator.AuditTrailEvent{
						Id:           "00000000-0000-0000-0005-000000000002",
						AuditScopeId: orchestratortest.MockScopeId2,
					}))
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ListAuditTrailEventsResponse], args ...any) bool {
				return assert.NotNil(t, got.Msg) &&
					assert.Equal(t, 1, len(got.Msg.AuditTrailEvents)) &&
					assert.Equal(t, orchestratortest.MockScopeId1, got.Msg.AuditTrailEvents[0].AuditScopeId)
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: no access returns empty list",
			args: args{
				req: &orchestrator.ListAuditTrailEventsRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					assert.NoError(t, d.Create(orchestratortest.MockAuditScope1))
					assert.NoError(t, d.Create(&orchestrator.AuditTrailEvent{
						Id:           "00000000-0000-0000-0005-000000000001",
						AuditScopeId: orchestratortest.MockScopeId1,
					}))
				}),
				authz: &denyAuthorizationStrategy{},
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ListAuditTrailEventsResponse], args ...any) bool {
				return assert.NotNil(t, got.Msg) &&
					assert.Equal(t, 0, len(got.Msg.AuditTrailEvents))
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
			res, err := svc.ListAuditTrailEvents(tt.args.context, connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}
