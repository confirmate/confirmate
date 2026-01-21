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

func TestService_CreateAuditScope(t *testing.T) {
	type args struct {
		req *orchestrator.CreateAuditScopeRequest
	}
	type fields struct {
		db persistence.DB
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
			name: "happy path",
			args: args{
				req: &orchestrator.CreateAuditScopeRequest{
					AuditScope: &orchestrator.AuditScope{
						TargetOfEvaluationId: orchestratortest.MockAuditScope1.TargetOfEvaluationId,
						CatalogId:            orchestratortest.MockAuditScope1.CatalogId,
					},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.AuditScope], args ...any) bool {
				return assert.NotNil(t, got.Msg) &&
					assert.NotEmpty(t, got.Msg.Id)
			},
			wantErr: assert.NoError,
			wantDB: func(t *testing.T, db persistence.DB, msgAndArgs ...any) bool {
				res := assert.Is[*connect.Response[orchestrator.AuditScope]](t, msgAndArgs[0])
				assert.NotNil(t, res)

				scope := assert.InDB[orchestrator.AuditScope](t, db, res.Msg.Id)
				assert.Equal(t, orchestratortest.MockAuditScope1.TargetOfEvaluationId, scope.TargetOfEvaluationId)
				assert.Equal(t, orchestratortest.MockAuditScope1.CatalogId, scope.CatalogId)
				return true
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
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument)
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
			name: "db error - unique constraint",
			args: args{
				req: &orchestrator.CreateAuditScopeRequest{
					AuditScope: &orchestrator.AuditScope{
						TargetOfEvaluationId: orchestratortest.MockAuditScope1.TargetOfEvaluationId,
						CatalogId:            orchestratortest.MockAuditScope1.CatalogId,
					},
				},
			},
			fields: fields{
				db: persistencetest.CreateErrorDB(t, persistence.ErrUniqueConstraintFailed, types, joinTables),
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
				db: tt.fields.db,
			}
			res, err := svc.CreateAuditScope(context.Background(), connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
			tt.wantDB(t, tt.fields.db, res)
		})
	}
}

func TestService_GetAuditScope(t *testing.T) {
	type args struct {
		req *orchestrator.GetAuditScopeRequest
	}
	type fields struct {
		db persistence.DB
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
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.AuditScope], args ...any) bool {
				return assert.NotNil(t, got.Msg) &&
					assert.Equal(t, orchestratortest.MockAuditScope1.Id, got.Msg.Id)
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
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument)
			},
		},
		{
			name: "not found",
			args: args{
				req: &orchestrator.GetAuditScopeRequest{
					AuditScopeId: orchestratortest.MockNonExistentID,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.AuditScope]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeNotFound)
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
				db: persistencetest.GetErrorDB(t, persistence.ErrRecordNotFound, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.AuditScope]],
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
			res, err := svc.GetAuditScope(context.Background(), connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}

func TestService_ListAuditScopes(t *testing.T) {
	type args struct {
		req *orchestrator.ListAuditScopesRequest
	}
	type fields struct {
		db persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[orchestrator.ListAuditScopesResponse]]
		wantErr assert.WantErr
	}{
		{
			name: "list all",
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
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ListAuditScopesResponse], args ...any) bool {
				return assert.NotNil(t, got.Msg) &&
					assert.Equal(t, 2, len(got.Msg.AuditScopes))
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
				db: tt.fields.db,
			}
			res, err := svc.ListAuditScopes(context.Background(), connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}

func TestService_UpdateAuditScope(t *testing.T) {
	type args struct {
		req *orchestrator.UpdateAuditScopeRequest
	}
	type fields struct {
		db persistence.DB
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
				req: &orchestrator.UpdateAuditScopeRequest{
					AuditScope: &orchestrator.AuditScope{
						Id:                   orchestratortest.MockAuditScope1.Id,
						TargetOfEvaluationId: orchestratortest.MockToeID2,
						CatalogId:            "catalog-1-updated",
					},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockAuditScope1)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.AuditScope], args ...any) bool {
				return assert.NotNil(t, got.Msg) &&
					assert.Equal(t, orchestratortest.MockAuditScope1.Id, got.Msg.Id)
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
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument)
			},
		},
		{
			name: "validation error - missing id",
			args: args{
				req: &orchestrator.UpdateAuditScopeRequest{
					AuditScope: &orchestrator.AuditScope{
						TargetOfEvaluationId: orchestratortest.MockToeID1,
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
						Id:                   orchestratortest.MockNonExistentID,
						TargetOfEvaluationId: orchestratortest.MockAuditScope1.TargetOfEvaluationId,
						CatalogId:            orchestratortest.MockAuditScope1.CatalogId,
					},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.AuditScope]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeNotFound)
			},
		},
		{
			name: "db error - constraint",
			args: args{
				req: &orchestrator.UpdateAuditScopeRequest{
					AuditScope: &orchestrator.AuditScope{
						Id:                   orchestratortest.MockAuditScope1.Id,
						TargetOfEvaluationId: orchestratortest.MockAuditScope1.TargetOfEvaluationId,
						CatalogId:            orchestratortest.MockAuditScope1.CatalogId,
					},
				},
			},
			fields: fields{
				db: persistencetest.UpdateErrorDB(t, persistence.ErrConstraintFailed, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.AuditScope]],
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
			res, err := svc.UpdateAuditScope(context.Background(), connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}

func TestService_RemoveAuditScope(t *testing.T) {
	type args struct {
		req *orchestrator.RemoveAuditScopeRequest
	}
	type fields struct {
		db persistence.DB
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
				req: &orchestrator.RemoveAuditScopeRequest{
					AuditScopeId: orchestratortest.MockAuditScope1.Id,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockAuditScope1)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *connect.Response[emptypb.Empty], args ...any) bool {
				return assert.NotNil(t, got)
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
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument)
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
				db: persistencetest.GetErrorDB(t, persistence.ErrRecordNotFound, types, joinTables),
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
				db: tt.fields.db,
			}
			res, err := svc.RemoveAuditScope(context.Background(), connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}
