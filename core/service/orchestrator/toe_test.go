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

	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/persistence"
	"confirmate.io/core/persistence/persistencetest"
	"confirmate.io/core/service/orchestrator/orchestratortest"
	"confirmate.io/core/util/assert"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestService_CreateTargetOfEvaluation(t *testing.T) {
	type args struct {
		req *orchestrator.CreateTargetOfEvaluationRequest
	}
	type fields struct {
		db persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[orchestrator.TargetOfEvaluation]]
		wantErr assert.WantErr
		wantDB  assert.Want[persistence.DB]
	}{
		{
			name: "happy path",
			args: args{
				req: &orchestrator.CreateTargetOfEvaluationRequest{
					TargetOfEvaluation: &orchestrator.TargetOfEvaluation{
						Name: "test-toe",
					},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.TargetOfEvaluation], args ...any) bool {
				return assert.NotNil(t, got.Msg) &&
					assert.Equal(t, "test-toe", got.Msg.Name) &&
					assert.NotEmpty(t, got.Msg.Id) &&
					assert.NotNil(t, got.Msg.CreatedAt) &&
					assert.NotNil(t, got.Msg.UpdatedAt)
			},
			wantErr: assert.NoError,
			wantDB: func(t *testing.T, db persistence.DB, msgAndArgs ...any) bool {
				res := assert.Is[*connect.Response[orchestrator.TargetOfEvaluation]](t, msgAndArgs[0])
				assert.NotNil(t, res)

				toe := assert.InDB[orchestrator.TargetOfEvaluation](t, db, res.Msg.Id)
				assert.Equal(t, "test-toe", toe.Name)
				return true
			},
		},
		{
			name: "validation error - empty request",
			args: args{
				req: &orchestrator.CreateTargetOfEvaluationRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.TargetOfEvaluation]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument)
			},
			wantDB: assert.NotNil[persistence.DB],
		},
		{
			name: "validation error - missing name",
			args: args{
				req: &orchestrator.CreateTargetOfEvaluationRequest{
					TargetOfEvaluation: &orchestrator.TargetOfEvaluation{},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.TargetOfEvaluation]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.IsValidationError(t, err, "target_of_evaluation.name")
			},
			wantDB: assert.NotNil[persistence.DB],
		},
		{
			name: "db error - unique constraint",
			args: args{
				req: &orchestrator.CreateTargetOfEvaluationRequest{
					TargetOfEvaluation: &orchestrator.TargetOfEvaluation{
						Name: "test-toe",
					},
				},
			},
			fields: fields{
				db: persistencetest.CreateErrorDB(t, persistence.ErrUniqueConstraintFailed, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.TargetOfEvaluation]],
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
			res, err := svc.CreateTargetOfEvaluation(context.Background(), connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
			tt.wantDB(t, tt.fields.db, res)
		})
	}
}

func TestService_GetTargetOfEvaluation(t *testing.T) {
	type args struct {
		req *orchestrator.GetTargetOfEvaluationRequest
	}
	type fields struct {
		db persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[orchestrator.TargetOfEvaluation]]
		wantErr assert.WantErr
	}{
		{
			name: "happy path",
			args: args{
				req: &orchestrator.GetTargetOfEvaluationRequest{
					TargetOfEvaluationId: orchestratortest.MockTargetOfEvaluation1.Id,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockTargetOfEvaluation1)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.TargetOfEvaluation], args ...any) bool {
				return assert.NotNil(t, got.Msg) &&
					assert.Equal(t, orchestratortest.MockTargetOfEvaluation1.Id, got.Msg.Id)
			},
			wantErr: assert.NoError,
		},
		{
			name: "validation error - empty request",
			args: args{
				req: &orchestrator.GetTargetOfEvaluationRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.TargetOfEvaluation]],
			wantErr: func(t *testing.T, err error, args ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument)
			},
		},
		{
			name: "not found",
			args: args{
				req: &orchestrator.GetTargetOfEvaluationRequest{
					TargetOfEvaluationId: orchestratortest.MockEmptyUUID,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.TargetOfEvaluation]],
			wantErr: func(t *testing.T, err error, args ...any) bool {
				return assert.ErrorContains(t, err, "target of evaluation not found")
			},
		},
		{
			name: "db error - not found",
			args: args{
				req: &orchestrator.GetTargetOfEvaluationRequest{
					TargetOfEvaluationId: orchestratortest.MockTargetOfEvaluation1.Id,
				},
			},
			fields: fields{
				db: persistencetest.GetErrorDB(t, persistence.ErrRecordNotFound, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.TargetOfEvaluation]],
			wantErr: func(t *testing.T, err error, args ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: tt.fields.db,
			}
			res, err := svc.GetTargetOfEvaluation(context.Background(), connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}

func TestService_ListTargetsOfEvaluation(t *testing.T) {
	type args struct {
		req *orchestrator.ListTargetsOfEvaluationRequest
	}
	type fields struct {
		db persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[orchestrator.ListTargetsOfEvaluationResponse]]
		wantErr assert.WantErr
	}{
		{
			name: "happy path",
			args: args{
				req: &orchestrator.ListTargetsOfEvaluationRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockTargetOfEvaluation1)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ListTargetsOfEvaluationResponse], args ...any) bool {
				assert.NotNil(t, got.Msg)
				return assert.Equal(t, 1, len(got.Msg.TargetsOfEvaluation)) &&
					assert.Equal(t, orchestratortest.MockTargetOfEvaluation1.Id, got.Msg.TargetsOfEvaluation[0].Id)
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: tt.fields.db,
			}
			res, err := svc.ListTargetsOfEvaluation(context.Background(), connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}

func TestService_UpdateTargetOfEvaluation(t *testing.T) {
	type args struct {
		req *orchestrator.UpdateTargetOfEvaluationRequest
	}
	type fields struct {
		db persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[orchestrator.TargetOfEvaluation]]
		wantErr assert.WantErr
		wantDB  assert.Want[persistence.DB]
	}{
		{
			name: "happy path",
			args: args{
				req: &orchestrator.UpdateTargetOfEvaluationRequest{
					TargetOfEvaluation: &orchestrator.TargetOfEvaluation{
						Id:   orchestratortest.MockTargetOfEvaluation1.Id,
						Name: "updated-name",
					},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockTargetOfEvaluation1)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.TargetOfEvaluation], args ...any) bool {
				return assert.NotNil(t, got.Msg) &&
					assert.Equal(t, "updated-name", got.Msg.Name)
			},
			wantErr: assert.NoError,
			wantDB: func(t *testing.T, db persistence.DB, msgAndArgs ...any) bool {
				res := assert.Is[*connect.Response[orchestrator.TargetOfEvaluation]](t, msgAndArgs[0])
				assert.NotNil(t, res)

				toe := assert.InDB[orchestrator.TargetOfEvaluation](t, db, res.Msg.Id)
				assert.Equal(t, "updated-name", toe.Name)
				return true
			},
		},
		{
			name: "validation error - empty request",
			args: args{
				req: &orchestrator.UpdateTargetOfEvaluationRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.TargetOfEvaluation]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument)
			},
			wantDB: assert.NotNil[persistence.DB],
		},
		{
			name: "validation error - missing id",
			args: args{
				req: &orchestrator.UpdateTargetOfEvaluationRequest{
					TargetOfEvaluation: &orchestrator.TargetOfEvaluation{
						Name: "updated-name",
					},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.TargetOfEvaluation]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.IsValidationError(t, err, "target_of_evaluation.id")
			},
			wantDB: assert.NotNil[persistence.DB],
		},
		{
			name: "not found",
			args: args{
				req: &orchestrator.UpdateTargetOfEvaluationRequest{
					TargetOfEvaluation: &orchestrator.TargetOfEvaluation{
						Id:   orchestratortest.MockEmptyUUID,
						Name: "updated-name",
					},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.TargetOfEvaluation]],
			wantErr: func(t *testing.T, err error, args ...any) bool {
				return assert.ErrorContains(t, err, "target of evaluation not found")
			},
			wantDB: assert.NotNil[persistence.DB],
		},
		{
			name: "db error - constraint",
			args: args{
				req: &orchestrator.UpdateTargetOfEvaluationRequest{
					TargetOfEvaluation: &orchestrator.TargetOfEvaluation{
						Id:   orchestratortest.MockTargetOfEvaluation1.Id,
						Name: "updated-name",
					},
				},
			},
			fields: fields{
				db: persistencetest.UpdateErrorDB(t, persistence.ErrConstraintFailed, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.TargetOfEvaluation]],
			wantErr: func(t *testing.T, err error, args ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument)
			},
			wantDB: assert.NotNil[persistence.DB],
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: tt.fields.db,
			}
			res, err := svc.UpdateTargetOfEvaluation(context.Background(), connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
			tt.wantDB(t, tt.fields.db, res)
		})
	}
}

func TestService_RemoveTargetOfEvaluation(t *testing.T) {
	type args struct {
		req *orchestrator.RemoveTargetOfEvaluationRequest
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
		wantDB  assert.Want[persistence.DB]
	}{
		{
			name: "happy path",
			args: args{
				req: &orchestrator.RemoveTargetOfEvaluationRequest{
					TargetOfEvaluationId: orchestratortest.MockTargetOfEvaluation1.Id,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockTargetOfEvaluation1)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *connect.Response[emptypb.Empty], args ...any) bool {
				return assert.NotNil(t, got.Msg)
			},
			wantErr: assert.NoError,
			wantDB: func(t *testing.T, db persistence.DB, msgAndArgs ...any) bool {
				res := assert.Is[*connect.Response[emptypb.Empty]](t, msgAndArgs[0])
				assert.NotNil(t, res)

				// Verify entity was deleted
				var toe orchestrator.TargetOfEvaluation
				err := db.Get(&toe, "id = ?", orchestratortest.MockTargetOfEvaluation1.Id)
				assert.ErrorIs(t, err, persistence.ErrRecordNotFound)
				return true
			},
		},
		{
			name: "validation error - empty request",
			args: args{
				req: &orchestrator.RemoveTargetOfEvaluationRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[emptypb.Empty]],
			wantErr: func(t *testing.T, err error, args ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument)
			},
			wantDB: func(t *testing.T, db persistence.DB, msgAndArgs ...any) bool {
				return true
			},
		},
		{
			name: "db error - not found",
			args: args{
				req: &orchestrator.RemoveTargetOfEvaluationRequest{
					TargetOfEvaluationId: orchestratortest.MockTargetOfEvaluation1.Id,
				},
			},
			fields: fields{
				db: persistencetest.GetErrorDB(t, persistence.ErrRecordNotFound, types, joinTables),
			},
			want: assert.Nil[*connect.Response[emptypb.Empty]],
			wantErr: func(t *testing.T, err error, args ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeNotFound)
			},
			wantDB: func(t *testing.T, db persistence.DB, msgAndArgs ...any) bool {
				return true
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: tt.fields.db,
			}
			res, err := svc.RemoveTargetOfEvaluation(context.Background(), connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
			tt.wantDB(t, tt.fields.db, res)
		})
	}
}

func TestService_GetTargetOfEvaluationStatistics(t *testing.T) {
	type args struct {
		req *orchestrator.GetTargetOfEvaluationStatisticsRequest
	}
	type fields struct {
		db persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[orchestrator.GetTargetOfEvaluationStatisticsResponse]]
		wantErr assert.WantErr
	}{
		{
			name: "happy path",
			args: args{
				req: &orchestrator.GetTargetOfEvaluationStatisticsRequest{
					TargetOfEvaluationId: orchestratortest.MockTargetOfEvaluation1.Id,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockTargetOfEvaluation1)
					assert.NoError(t, err)

					// Create some assessment results
					err = d.Create(&assessment.AssessmentResult{
						Id:                   "result-1",
						TargetOfEvaluationId: orchestratortest.MockTargetOfEvaluation1.Id,
					})
					assert.NoError(t, err)

					// Create some audit scopes
					err = d.Create(&orchestrator.AuditScope{
						Id:                   "scope-1",
						TargetOfEvaluationId: orchestratortest.MockTargetOfEvaluation1.Id,
					})
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.GetTargetOfEvaluationStatisticsResponse], args ...any) bool {
				return assert.NotNil(t, got.Msg) &&
					assert.Equal(t, int64(1), got.Msg.NumberOfAssessmentResults) &&
					assert.Equal(t, int64(1), got.Msg.NumberOfSelectedCatalogs)
			},
			wantErr: assert.NoError,
		},
		{
			name: "validation error - empty request",
			args: args{
				req: &orchestrator.GetTargetOfEvaluationStatisticsRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.GetTargetOfEvaluationStatisticsResponse]],
			wantErr: func(t *testing.T, err error, args ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.IsValidationError(t, err, "target_of_evaluation_id")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: tt.fields.db,
			}
			res, err := svc.GetTargetOfEvaluationStatistics(context.Background(), connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}

func TestService_CreateDefaultTargetOfEvaluation(t *testing.T) {
	type fields struct {
		db persistence.DB
	}
	tests := []struct {
		name    string
		fields  fields
		want    assert.Want[*orchestrator.TargetOfEvaluation]
		wantErr assert.WantErr
	}{
		{
			name: "create default",
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: func(t *testing.T, got *orchestrator.TargetOfEvaluation, args ...any) bool {
				return assert.NotNil(t, got) &&
					assert.Equal(t, "Default Target of Evaluation", got.Name)
			},
			wantErr: assert.NoError,
		},
		{
			name: "already exists",
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockTargetOfEvaluation1)
					assert.NoError(t, err)
				}),
			},
			want:    assert.Nil[*orchestrator.TargetOfEvaluation],
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: tt.fields.db,
			}
			res, err := svc.CreateDefaultTargetOfEvaluation()
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}
