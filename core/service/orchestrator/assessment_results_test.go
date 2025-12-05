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
)

func TestService_StoreAssessmentResult(t *testing.T) {
	type args struct {
		req *orchestrator.StoreAssessmentResultRequest
	}
	type fields struct {
		db *persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*orchestrator.StoreAssessmentResultResponse]
		wantErr assert.WantErr
	}{
		{
			name: "happy path",
			args: args{
				req: &orchestrator.StoreAssessmentResultRequest{
					Result: orchestratortest.MockNewAssessmentResult,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want:    assert.NotNil[*orchestrator.StoreAssessmentResultResponse],
			wantErr: nil,
		},
		{
			name: "with existing ID",
			args: args{
				req: &orchestrator.StoreAssessmentResultRequest{
					Result: orchestratortest.MockNewAssessmentResultWithId,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want:    assert.NotNil[*orchestrator.StoreAssessmentResultResponse],
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: tt.fields.db,
			}
			res, err := svc.StoreAssessmentResult(context.Background(), connect.NewRequest(tt.args.req))
			assert.WantResponse(t, res, err, tt.want, tt.wantErr)
		})
	}
}

func TestService_GetAssessmentResult(t *testing.T) {
	type args struct {
		req *orchestrator.GetAssessmentResultRequest
	}
	type fields struct {
		db *persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*assessment.AssessmentResult]
		wantErr assert.WantErr
	}{
		{
			name: "happy path",
			args: args{
				req: &orchestrator.GetAssessmentResultRequest{
					Id: orchestratortest.MockAssessmentResult1.Id,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d *persistence.DB) {
					err := d.Create(orchestratortest.MockAssessmentResult1)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *assessment.AssessmentResult, args ...any) bool {
				return assert.NotNil(t, got) &&
					assert.Equal(t, orchestratortest.MockAssessmentResult1.Id, got.Id)
			},
			wantErr: nil,
		},
		{
			name: "not found",
			args: args{
				req: &orchestrator.GetAssessmentResultRequest{
					Id: "non-existent",
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: nil,
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
			res, err := svc.GetAssessmentResult(context.Background(), connect.NewRequest(tt.args.req))
			assert.WantResponse(t, res, err, tt.want, tt.wantErr)
		})
	}
}

func TestService_ListAssessmentResults(t *testing.T) {
	type args struct {
		req *orchestrator.ListAssessmentResultsRequest
	}
	type fields struct {
		db *persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*orchestrator.ListAssessmentResultsResponse]
		wantErr assert.WantErr
	}{
		{
			name: "list all",
			args: args{
				req: &orchestrator.ListAssessmentResultsRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d *persistence.DB) {
					err := d.Create(orchestratortest.MockAssessmentResult1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockAssessmentResult2)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *orchestrator.ListAssessmentResultsResponse, args ...any) bool {
				return assert.NotNil(t, got) &&
					assert.Equal(t, 2, len(got.Results))
			},
			wantErr: nil,
		},
		{
			name: "filter by metric ID",
			args: args{
				req: &orchestrator.ListAssessmentResultsRequest{
					Filter: &orchestrator.ListAssessmentResultsRequest_Filter{
						MetricId: stringPtr(orchestratortest.MockAssessmentResult1.MetricId),
					},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d *persistence.DB) {
					err := d.Create(orchestratortest.MockAssessmentResult1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockAssessmentResult2)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *orchestrator.ListAssessmentResultsResponse, args ...any) bool {
				return assert.NotNil(t, got) &&
					assert.Equal(t, 1, len(got.Results))
			},
			wantErr: nil,
		},
		{
			name: "filter by compliant",
			args: args{
				req: &orchestrator.ListAssessmentResultsRequest{
					Filter: &orchestrator.ListAssessmentResultsRequest_Filter{
						Compliant: boolPtr(true),
					},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d *persistence.DB) {
					err := d.Create(orchestratortest.MockAssessmentResult1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockAssessmentResult2)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *orchestrator.ListAssessmentResultsResponse, args ...any) bool {
				return assert.NotNil(t, got) &&
					assert.Equal(t, 1, len(got.Results))
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: tt.fields.db,
			}
			res, err := svc.ListAssessmentResults(context.Background(), connect.NewRequest(tt.args.req))
			assert.WantResponse(t, res, err, tt.want, tt.wantErr)
		})
	}
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
