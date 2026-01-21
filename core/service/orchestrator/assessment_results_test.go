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
	"time"

	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/persistence"
	"confirmate.io/core/persistence/persistencetest"
	"confirmate.io/core/service/orchestrator/orchestratortest"
	"confirmate.io/core/util"
	"confirmate.io/core/util/assert"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestService_StoreAssessmentResult(t *testing.T) {
	type args struct {
		req *orchestrator.StoreAssessmentResultRequest
	}
	type fields struct {
		db persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[orchestrator.StoreAssessmentResultResponse]]
		wantErr assert.WantErr
		wantDB  assert.Want[persistence.DB]
	}{
		{
			name: "validation error - empty request",
			args: args{
				req: &orchestrator.StoreAssessmentResultRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.StoreAssessmentResultResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.ErrorContains(t, err, "invalid request")
			},
			wantDB: assert.NotNil[persistence.DB],
		},
		{
			name: "validation error - missing metric id",
			args: args{
				req: &orchestrator.StoreAssessmentResultRequest{
					Result: &assessment.AssessmentResult{
						EvidenceId:           orchestratortest.MockEvidenceId1,
						TargetOfEvaluationId: orchestratortest.MockToeId1,
					},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.StoreAssessmentResultResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.ErrorContains(t, err, "metric_id")
			},
			wantDB: assert.NotNil[persistence.DB],
		},
		{
			name: "db error - unique constraint",
			args: args{
				req: &orchestrator.StoreAssessmentResultRequest{
					Result: orchestratortest.MockAssessmentResult1,
				},
			},
			fields: fields{
				db: persistencetest.CreateErrorDB(t, persistence.ErrUniqueConstraintFailed, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.StoreAssessmentResultResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeAlreadyExists)
			},
			wantDB: assert.NotNil[persistence.DB],
		},
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
			want:    assert.NotNil[*connect.Response[orchestrator.StoreAssessmentResultResponse]],
			wantErr: assert.NoError,
			wantDB: func(t *testing.T, db persistence.DB, msgAndArgs ...any) bool {
				// Verify the result was persisted with correct timestamp
				result := assert.InDB[assessment.AssessmentResult](t, db, orchestratortest.MockResultId3)
				assert.NotNil(t, result.CreatedAt)
				assert.True(t, time.Since(result.CreatedAt.AsTime()) < 5*time.Second)
				assert.Equal(t, orchestratortest.MockMetricId1, result.MetricId)
				assert.Equal(t, orchestratortest.MockResourceIdNew, result.ResourceId)
				return true
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: tt.fields.db,
			}
			res, err := svc.StoreAssessmentResult(context.Background(), connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
			tt.wantDB(t, tt.fields.db)
		})
	}
}

func TestService_GetAssessmentResult(t *testing.T) {
	type args struct {
		req *orchestrator.GetAssessmentResultRequest
	}
	type fields struct {
		db persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[assessment.AssessmentResult]]
		wantErr assert.WantErr
	}{
		{
			name: "validation error - empty id",
			args: args{
				req: &orchestrator.GetAssessmentResultRequest{
					Id: "",
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[assessment.AssessmentResult]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.ErrorContains(t, err, "id")
			},
		},
		{
			name: "validation error - invalid uuid",
			args: args{
				req: &orchestrator.GetAssessmentResultRequest{
					Id: "not-a-uuid",
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[assessment.AssessmentResult]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.ErrorContains(t, err, "id")
			},
		},
		{
			name: "happy path",
			args: args{
				req: &orchestrator.GetAssessmentResultRequest{
					Id: orchestratortest.MockAssessmentResult1.Id,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockAssessmentResult1)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *connect.Response[assessment.AssessmentResult], args ...any) bool {
				assert.NotNil(t, got.Msg)
				return assert.Equal(t, orchestratortest.MockAssessmentResult1.Id, got.Msg.Id)
			},
			wantErr: assert.NoError,
		},
		{
			name: "not found",
			args: args{
				req: &orchestrator.GetAssessmentResultRequest{
					Id: orchestratortest.MockNonExistentId,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[assessment.AssessmentResult]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeNotFound)
			},
		},
		{
			name: "db error - not found",
			args: args{
				req: &orchestrator.GetAssessmentResultRequest{
					Id: orchestratortest.MockAssessmentResult1.Id,
				},
			},
			fields: fields{
				db: persistencetest.GetErrorDB(t, persistence.ErrRecordNotFound, types, joinTables),
			},
			want: assert.Nil[*connect.Response[assessment.AssessmentResult]],
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
			res, err := svc.GetAssessmentResult(context.Background(), connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}

func TestService_ListAssessmentResults(t *testing.T) {
	type args struct {
		req *orchestrator.ListAssessmentResultsRequest
	}
	type fields struct {
		db persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[orchestrator.ListAssessmentResultsResponse]]
		wantErr assert.WantErr
	}{
		{
			name: "validation error",
			args: args{
				req: &orchestrator.ListAssessmentResultsRequest{
					PageToken: "!!!invalid-base64!!!",
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.ListAssessmentResultsResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument)
			},
		},
		{
			name: "list all",
			args: args{
				req: &orchestrator.ListAssessmentResultsRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockAssessmentResult1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockAssessmentResult2)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ListAssessmentResultsResponse], args ...any) bool {
				return assert.NotNil(t, got.Msg) &&
					assert.Equal(t, 2, len(got.Msg.Results))
			},
			wantErr: assert.NoError,
		},
		{
			name: "filter by metric ID",
			args: args{
				req: &orchestrator.ListAssessmentResultsRequest{
					Filter: &orchestrator.ListAssessmentResultsRequest_Filter{
						MetricId: &orchestratortest.MockAssessmentResult1.MetricId,
					},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockAssessmentResult1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockAssessmentResult2)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ListAssessmentResultsResponse], args ...any) bool {
				return assert.NotNil(t, got.Msg) &&
					assert.Equal(t, 1, len(got.Msg.Results))
			},
			wantErr: assert.NoError,
		},
		{
			name: "filter by compliant",
			args: args{
				req: &orchestrator.ListAssessmentResultsRequest{
					Filter: &orchestrator.ListAssessmentResultsRequest_Filter{
						Compliant: &[]bool{true}[0],
					},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockAssessmentResult1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockAssessmentResult2)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ListAssessmentResultsResponse], args ...any) bool {
				return assert.NotNil(t, got.Msg) &&
					assert.Equal(t, 1, len(got.Msg.Results))
			},
			wantErr: assert.NoError,
		},
		{
			name: "filter by tool ID",
			args: args{
				req: &orchestrator.ListAssessmentResultsRequest{
					Filter: &orchestrator.ListAssessmentResultsRequest_Filter{
						ToolId: orchestratortest.MockAssessmentResult1.ToolId,
					},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockAssessmentResult1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockAssessmentResult2)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ListAssessmentResultsResponse], args ...any) bool {
				// Both MockAssessmentResult1 and MockAssessmentResult2 have tool-1
				return assert.NotNil(t, got.Msg) &&
					assert.Equal(t, 2, len(got.Msg.Results))
			},
			wantErr: assert.NoError,
		},
		{
			name: "filter by target of evaluation ID",
			args: args{
				req: &orchestrator.ListAssessmentResultsRequest{
					Filter: &orchestrator.ListAssessmentResultsRequest_Filter{
						TargetOfEvaluationId: util.Ref(orchestratortest.MockToeId1),
					},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockAssessmentResult1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockAssessmentResult2)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ListAssessmentResultsResponse], args ...any) bool {
				// Both MockAssessmentResult1 and MockAssessmentResult2 have the same TOE ID
				return assert.NotNil(t, got.Msg) &&
					assert.Equal(t, 2, len(got.Msg.Results))
			},
			wantErr: assert.NoError,
		},
		{
			name: "filter by assessment result IDs",
			args: args{
				req: &orchestrator.ListAssessmentResultsRequest{
					Filter: &orchestrator.ListAssessmentResultsRequest_Filter{
						AssessmentResultIds: []string{orchestratortest.MockAssessmentResult1.Id},
					},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockAssessmentResult1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockAssessmentResult2)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ListAssessmentResultsResponse], args ...any) bool {
				return assert.NotNil(t, got.Msg) &&
					assert.Equal(t, 1, len(got.Msg.Results)) &&
					assert.Equal(t, orchestratortest.MockAssessmentResult1.Id, got.Msg.Results[0].Id)
			},
			wantErr: assert.NoError,
		},
		{
			name: "filter by latest_by_resource_id",
			args: args{
				req: &orchestrator.ListAssessmentResultsRequest{
					LatestByResourceId: &[]bool{true}[0],
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					// Create multiple results with different combinations of resource_id and metric_id
					// to test that we get the latest for each unique (resource_id, metric_id) pair

					// Resource 1, Metric 1: 3 results, latest should be result-1-1-latest
					result11old := &assessment.AssessmentResult{
						Id:                   "result-1-1-old",
						CreatedAt:            timestamppb.New(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
						MetricId:             "metric-1",
						ResourceId:           "resource-1",
						TargetOfEvaluationId: orchestratortest.MockToeId1,
					}
					result11middle := &assessment.AssessmentResult{
						Id:                   "result-1-1-middle",
						CreatedAt:            timestamppb.New(time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)),
						MetricId:             "metric-1",
						ResourceId:           "resource-1",
						TargetOfEvaluationId: orchestratortest.MockToeId1,
					}
					result11latest := &assessment.AssessmentResult{
						Id:                   "result-1-1-latest",
						CreatedAt:            timestamppb.New(time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)),
						MetricId:             "metric-1",
						ResourceId:           "resource-1",
						TargetOfEvaluationId: orchestratortest.MockToeId1,
					}

					// Resource 1, Metric 2: 2 results, latest should be result-1-2-latest
					result12old := &assessment.AssessmentResult{
						Id:                   "result-1-2-old",
						CreatedAt:            timestamppb.New(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
						MetricId:             "metric-2",
						ResourceId:           "resource-1",
						TargetOfEvaluationId: orchestratortest.MockToeId1,
					}
					result12latest := &assessment.AssessmentResult{
						Id:                   "result-1-2-latest",
						CreatedAt:            timestamppb.New(time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)),
						MetricId:             "metric-2",
						ResourceId:           "resource-1",
						TargetOfEvaluationId: orchestratortest.MockToeId1,
					}

					// Resource 2, Metric 1: 2 results, latest should be result-2-1-latest
					result21old := &assessment.AssessmentResult{
						Id:                   "result-2-1-old",
						CreatedAt:            timestamppb.New(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
						MetricId:             "metric-1",
						ResourceId:           "resource-2",
						TargetOfEvaluationId: orchestratortest.MockToeId1,
					}
					result21latest := &assessment.AssessmentResult{
						Id:                   "result-2-1-latest",
						CreatedAt:            timestamppb.New(time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)),
						MetricId:             "metric-1",
						ResourceId:           "resource-2",
						TargetOfEvaluationId: orchestratortest.MockToeId1,
					}

					// Resource 2, Metric 2: 1 result, should be returned
					result22single := &assessment.AssessmentResult{
						Id:                   "result-2-2-single",
						CreatedAt:            timestamppb.New(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
						MetricId:             "metric-2",
						ResourceId:           "resource-2",
						TargetOfEvaluationId: orchestratortest.MockToeId1,
					}

					// Insert in random order to ensure ordering by created_at works
					results := []*assessment.AssessmentResult{
						result11middle, result21old, result12latest, result11old,
						result21latest, result12old, result22single, result11latest,
					}
					for _, r := range results {
						err := d.Create(r)
						assert.NoError(t, err)
					}
				}),
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ListAssessmentResultsResponse], args ...any) bool {
				// Should return exactly 4 results (one per unique resource_id/metric_id combination)
				if !assert.NotNil(t, got.Msg) || !assert.Equal(t, 4, len(got.Msg.Results)) {
					return false
				}

				// Collect returned IDs
				ids := make(map[string]bool)
				for _, r := range got.Msg.Results {
					ids[r.Id] = true
				}

				// Verify we got the latest result for each (resource_id, metric_id) pair
				expectedIds := []string{
					"result-1-1-latest", // resource-1, metric-1: latest of 3
					"result-1-2-latest", // resource-1, metric-2: latest of 2
					"result-2-1-latest", // resource-2, metric-1: latest of 2
					"result-2-2-single", // resource-2, metric-2: only 1
				}

				for _, expectedId := range expectedIds {
					if !ids[expectedId] {
						t.Errorf("Expected result %s not found in response", expectedId)
						return false
					}
				}

				return true
			},
			wantErr: assert.NoError,
		},
		{
			name: "filter by latest_by_resource_id with conditions",
			args: args{
				req: &orchestrator.ListAssessmentResultsRequest{
					LatestByResourceId: &[]bool{true}[0],
					Filter: &orchestrator.ListAssessmentResultsRequest_Filter{
						MetricId: &[]string{"metric-1"}[0],
					},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					// Create results for different metrics and resources
					result11old := &assessment.AssessmentResult{
						Id:                   "result-1-1-old",
						CreatedAt:            timestamppb.New(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
						MetricId:             "metric-1",
						ResourceId:           "resource-1",
						TargetOfEvaluationId: orchestratortest.MockToeId1,
					}
					result11latest := &assessment.AssessmentResult{
						Id:                   "result-1-1-latest",
						CreatedAt:            timestamppb.New(time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)),
						MetricId:             "metric-1",
						ResourceId:           "resource-1",
						TargetOfEvaluationId: orchestratortest.MockToeId1,
					}
					// This should be filtered out due to metric-2
					result12latest := &assessment.AssessmentResult{
						Id:                   "result-1-2-latest",
						CreatedAt:            timestamppb.New(time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)),
						MetricId:             "metric-2",
						ResourceId:           "resource-1",
						TargetOfEvaluationId: orchestratortest.MockToeId1,
					}
					result21latest := &assessment.AssessmentResult{
						Id:                   "result-2-1-latest",
						CreatedAt:            timestamppb.New(time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)),
						MetricId:             "metric-1",
						ResourceId:           "resource-2",
						TargetOfEvaluationId: orchestratortest.MockToeId1,
					}

					results := []*assessment.AssessmentResult{
						result11old, result11latest, result12latest, result21latest,
					}
					for _, r := range results {
						err := d.Create(r)
						assert.NoError(t, err)
					}
				}),
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ListAssessmentResultsResponse], args ...any) bool {
				// Should return exactly 2 results (latest for metric-1 only, for each resource)
				if !assert.NotNil(t, got.Msg) || !assert.Equal(t, 2, len(got.Msg.Results)) {
					return false
				}

				// Collect returned IDs
				ids := make(map[string]bool)
				for _, r := range got.Msg.Results {
					ids[r.Id] = true
					// Verify all results are for metric-1
					if r.MetricId != "metric-1" {
						t.Errorf("Expected only metric-1 results, got %s", r.MetricId)
						return false
					}
				}

				// Verify we got the latest result for each resource with metric-1
				expectedIds := []string{
					"result-1-1-latest", // resource-1, metric-1: latest
					"result-2-1-latest", // resource-2, metric-1: latest
				}

				for _, expectedId := range expectedIds {
					if !ids[expectedId] {
						t.Errorf("Expected result %s not found in response", expectedId)
						return false
					}
				}

				return true
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: tt.fields.db,
			}
			res, err := svc.ListAssessmentResults(context.Background(), connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}
