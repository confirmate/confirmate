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
		db *persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*orchestrator.AuditScope]
		wantErr assert.WantErr
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
			want: func(t *testing.T, got *orchestrator.AuditScope, args ...any) bool {
				return assert.NotNil(t, got) &&
					assert.NotEmpty(t, got.Id)
			},
			wantErr: nil,
		},
		{
			name: "with existing ID",
			args: args{
				req: &orchestrator.CreateAuditScopeRequest{
					AuditScope: &orchestrator.AuditScope{
						Id:                   "existing-id",
						TargetOfEvaluationId: orchestratortest.MockTargetOfEvaluation2.Id,
						CatalogId:            orchestratortest.MockCatalog2.Id,
					},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: func(t *testing.T, got *orchestrator.AuditScope, args ...any) bool {
				return assert.NotNil(t, got) &&
					assert.NotEmpty(t, got.Id)
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: tt.fields.db,
			}
			res, err := svc.CreateAuditScope(context.Background(), connect.NewRequest(tt.args.req))
			assert.WantResponse(t, res, err, tt.want, tt.wantErr)
		})
	}
}

func TestService_GetAuditScope(t *testing.T) {
	type args struct {
		req *orchestrator.GetAuditScopeRequest
	}
	type fields struct {
		db *persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*orchestrator.AuditScope]
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
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d *persistence.DB) {
					err := d.Create(orchestratortest.MockAuditScope1)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *orchestrator.AuditScope, args ...any) bool {
				return assert.NotNil(t, got) &&
					assert.Equal(t, orchestratortest.MockAuditScope1.Id, got.Id)
			},
			wantErr: nil,
		},
		{
			name: "not found",
			args: args{
				req: &orchestrator.GetAuditScopeRequest{
					AuditScopeId: "non-existent",
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: nil,
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool { cErr := assert.Is[*connect.Error](t, err);
				return assert.Equal(t, connect.CodeNotFound, cErr.Code())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: tt.fields.db,
			}
			res, err := svc.GetAuditScope(context.Background(), connect.NewRequest(tt.args.req))
			assert.WantResponse(t, res, err, tt.want, tt.wantErr)
		})
	}
}

func TestService_ListAuditScopes(t *testing.T) {
	type args struct {
		req *orchestrator.ListAuditScopesRequest
	}
	type fields struct {
		db *persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*orchestrator.ListAuditScopesResponse]
		wantErr assert.WantErr
	}{
		{
			name: "list all",
			args: args{
				req: &orchestrator.ListAuditScopesRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d *persistence.DB) {
					err := d.Create(orchestratortest.MockAuditScope1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockAuditScope2)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *orchestrator.ListAuditScopesResponse, args ...any) bool {
				return assert.NotNil(t, got) &&
					assert.Equal(t, 2, len(got.AuditScopes))
			},
			wantErr: nil,
		},
		{
			name: "filter by target of evaluation",
			args: args{
				req: &orchestrator.ListAuditScopesRequest{
					Filter: &orchestrator.ListAuditScopesRequest_Filter{
						TargetOfEvaluationId: stringPtr(orchestratortest.MockAuditScope1.TargetOfEvaluationId),
					},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d *persistence.DB) {
					err := d.Create(orchestratortest.MockAuditScope1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockAuditScope2)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *orchestrator.ListAuditScopesResponse, args ...any) bool {
				return assert.NotNil(t, got) &&
					assert.Equal(t, 1, len(got.AuditScopes))
			},
			wantErr: nil,
		},
		{
			name: "filter by catalog",
			args: args{
				req: &orchestrator.ListAuditScopesRequest{
					Filter: &orchestrator.ListAuditScopesRequest_Filter{
						CatalogId: stringPtr(orchestratortest.MockAuditScope1.CatalogId),
					},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d *persistence.DB) {
					err := d.Create(orchestratortest.MockAuditScope1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockAuditScope2)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *orchestrator.ListAuditScopesResponse, args ...any) bool {
				return assert.NotNil(t, got) &&
					assert.Equal(t, 1, len(got.AuditScopes))
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: tt.fields.db,
			}
			res, err := svc.ListAuditScopes(context.Background(), connect.NewRequest(tt.args.req))
			assert.WantResponse(t, res, err, tt.want, tt.wantErr)
		})
	}
}

func TestService_UpdateAuditScope(t *testing.T) {
	type args struct {
		req *orchestrator.UpdateAuditScopeRequest
	}
	type fields struct {
		db *persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*orchestrator.AuditScope]
		wantErr assert.WantErr
	}{
		{
			name: "happy path",
			args: args{
				req: &orchestrator.UpdateAuditScopeRequest{
					AuditScope: &orchestrator.AuditScope{
						Id:                   orchestratortest.MockAuditScope1.Id,
						TargetOfEvaluationId: "toe-1-updated",
						CatalogId:            "catalog-1-updated",
					},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d *persistence.DB) {
					err := d.Create(orchestratortest.MockAuditScope1)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *orchestrator.AuditScope, args ...any) bool {
				return assert.NotNil(t, got) &&
					assert.Equal(t, orchestratortest.MockAuditScope1.Id, got.Id)
			},
			wantErr: nil,
		},
		{
			name: "not found",
			args: args{
				req: &orchestrator.UpdateAuditScopeRequest{
					AuditScope: &orchestrator.AuditScope{
						Id:                   "non-existent",
						TargetOfEvaluationId: orchestratortest.MockAuditScope1.TargetOfEvaluationId,
						CatalogId:            orchestratortest.MockAuditScope1.CatalogId,
					},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: nil,
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool { cErr := assert.Is[*connect.Error](t, err);
				return assert.Equal(t, connect.CodeNotFound, cErr.Code())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: tt.fields.db,
			}
			res, err := svc.UpdateAuditScope(context.Background(), connect.NewRequest(tt.args.req))
			assert.WantResponse(t, res, err, tt.want, tt.wantErr)
		})
	}
}

func TestService_RemoveAuditScope(t *testing.T) {
	type args struct {
		req *orchestrator.RemoveAuditScopeRequest
	}
	type fields struct {
		db *persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*emptypb.Empty]
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
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d *persistence.DB) {
					err := d.Create(orchestratortest.MockAuditScope1)
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
			res, err := svc.RemoveAuditScope(context.Background(), connect.NewRequest(tt.args.req))
			assert.WantResponse(t, res, err, tt.want, tt.wantErr)
		})
	}
}
