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
	"errors"
	"testing"

	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/persistence"
	"confirmate.io/core/persistence/persistencetest"
	"confirmate.io/core/service/orchestrator/orchestratortest"
	"confirmate.io/core/util/assert"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestService_CreateMetric(t *testing.T) {
	type args struct {
		req *orchestrator.CreateMetricRequest
	}
	type fields struct {
		db persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[assessment.Metric]]
		wantErr assert.WantErr
	}{
		{
			name: "happy path",
			args: args{
				req: &orchestrator.CreateMetricRequest{
					Metric: orchestratortest.MockMetric1,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: func(t *testing.T, got *connect.Response[assessment.Metric], args ...any) bool {
				assert.NotNil(t, got.Msg)
				return assert.Equal(t, orchestratortest.MockMetric1.Id, got.Msg.Id) &&
					assert.Equal(t, orchestratortest.MockMetric1.Description, got.Msg.Description)
			},
			wantErr: assert.NoError,
		},
		{
			name: "validation error - empty request",
			args: args{
				req: &orchestrator.CreateMetricRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[assessment.Metric]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument)
			},
		},
		{
			name: "validation error - missing metric",
			args: args{
				req: &orchestrator.CreateMetricRequest{
					Metric: &assessment.Metric{},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[assessment.Metric]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.IsValidationError(t, err, "metric.id")
			},
		},
		{
			name: "db error - unique constraint",
			args: args{
				req: &orchestrator.CreateMetricRequest{
					Metric: orchestratortest.MockMetric1,
				},
			},
			fields: fields{
				db: persistencetest.CreateErrorDB(t, persistence.ErrUniqueConstraintFailed, types, joinTables),
			},
			want: assert.Nil[*connect.Response[assessment.Metric]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeAlreadyExists)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: tt.fields.db,
			}
			res, err := svc.CreateMetric(context.Background(), connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}

func TestService_GetMetric(t *testing.T) {
	type args struct {
		req *orchestrator.GetMetricRequest
	}
	type fields struct {
		db persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[assessment.Metric]]
		wantErr assert.WantErr
	}{
		{
			name: "happy path",
			args: args{
				req: &orchestrator.GetMetricRequest{
					MetricId: orchestratortest.MockMetric1.Id,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					assert.NoError(t, d.Create(orchestratortest.MockMetric1))
				}),
			},
			want: func(t *testing.T, got *connect.Response[assessment.Metric], args ...any) bool {
				assert.NotNil(t, got.Msg)
				assert.Equal(t, orchestratortest.MockMetric1.Id, got.Msg.Id)
				return true
			},
			wantErr: assert.NoError,
		},
		{
			name: "validation error - missing id",
			args: args{
				req: &orchestrator.GetMetricRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[assessment.Metric]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsValidationError(t, err, "metric_id")
			},
		},
		{
			name: "not found",
			args: args{
				req: &orchestrator.GetMetricRequest{
					MetricId: orchestratortest.MockNonExistentId,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[assessment.Metric]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeNotFound)
			},
		},
		{
			name: "db error - not found",
			args: args{
				req: &orchestrator.GetMetricRequest{
					MetricId: orchestratortest.MockMetric1.Id,
				},
			},
			fields: fields{
				db: persistencetest.GetErrorDB(t, persistence.ErrRecordNotFound, types, joinTables),
			},
			want: assert.Nil[*connect.Response[assessment.Metric]],
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

			res, err := svc.GetMetric(context.Background(), connect.NewRequest(tt.args.req))
			tt.wantErr(t, err)
			tt.want(t, res)
		})
	}
}

func TestService_ListMetrics(t *testing.T) {
	type args struct {
		req *orchestrator.ListMetricsRequest
	}
	type fields struct {
		db persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[orchestrator.ListMetricsResponse]]
		wantErr assert.WantErr
	}{
		{
			name: "list all",
			args: args{
				req: &orchestrator.ListMetricsRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockMetric1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockMetric2)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ListMetricsResponse], args ...any) bool {
				assert.NotNil(t, got)
				return assert.Equal(t, 2, len(got.Msg.Metrics))
			},
			wantErr: assert.NoError,
		},
		{
			name: "empty list",
			args: args{
				req: &orchestrator.ListMetricsRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ListMetricsResponse], args ...any) bool {
				assert.NotNil(t, got)
				return assert.Equal(t, 0, len(got.Msg.Metrics))
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: tt.fields.db,
			}

			res, err := svc.ListMetrics(context.Background(), connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}

func TestService_UpdateMetric(t *testing.T) {
	type args struct {
		req *orchestrator.UpdateMetricRequest
	}
	type fields struct {
		db persistence.DB
	}

	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[assessment.Metric]]
		wantErr assert.WantErr
	}{
		{
			name: "happy path",
			args: args{
				req: &orchestrator.UpdateMetricRequest{
					Metric: &assessment.Metric{
						Id:          orchestratortest.MockMetric1.Id,
						Name:        orchestratortest.MockMetricName1,
						Description: "Updated description",
						Version:     "v1",
						Category:    "test-category",
					},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockMetric1)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *connect.Response[assessment.Metric], args ...any) bool {
				assert.NotNil(t, got.Msg)
				return assert.Equal(t, "Updated description", got.Msg.Description)
			},
			wantErr: assert.NoError,
		},
		{
			name: "validation error - empty request",
			args: args{
				req: &orchestrator.UpdateMetricRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[assessment.Metric]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument)
			},
		},
		{
			name: "validation error - missing id",
			args: args{
				req: &orchestrator.UpdateMetricRequest{
					Metric: &assessment.Metric{
						Description: "Updated description",
					},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[assessment.Metric]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.IsValidationError(t, err, "metric.id")
			},
		},
		{
			name: "not found",
			args: args{
				req: &orchestrator.UpdateMetricRequest{
					Metric: &assessment.Metric{
						Id:          "99999999-9999-9999-9999-999999999999",
						Name:        orchestratortest.MockMetricName1,
						Description: "Updated description",
						Version:     "v1",
						Category:    "test-category",
					},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[assessment.Metric]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeNotFound)
			},
		},
		{
			name: "db error - constraint",
			args: args{
				req: &orchestrator.UpdateMetricRequest{
					Metric: &assessment.Metric{
						Id:          orchestratortest.MockMetric1.Id,
						Name:        orchestratortest.MockMetricName1,
						Description: "Updated description",
						Version:     "v1",
						Category:    "test-category",
					},
				},
			},
			fields: fields{
				db: persistencetest.UpdateErrorDB(t, persistence.ErrConstraintFailed, types, joinTables),
			},
			want: assert.Nil[*connect.Response[assessment.Metric]],
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
			res, err := svc.UpdateMetric(context.Background(), connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}

func TestService_RemoveMetric(t *testing.T) {
	type args struct {
		req *orchestrator.RemoveMetricRequest
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
				req: &orchestrator.RemoveMetricRequest{
					MetricId: orchestratortest.MockMetric1.Id,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockMetric1)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *connect.Response[emptypb.Empty], args ...any) bool {
				return assert.NotNil(t, got)
			},
			wantErr: assert.NoError,
		},
		{
			name: "validation error - missing id",
			args: args{
				req: &orchestrator.RemoveMetricRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[emptypb.Empty]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsValidationError(t, err, "metric_id")
			},
		},
		{
			name: "db error - not found",
			args: args{
				req: &orchestrator.RemoveMetricRequest{
					MetricId: orchestratortest.MockMetric1.Id,
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
			res, err := svc.RemoveMetric(context.Background(), connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}

func TestService_GetMetricImplementation(t *testing.T) {
	type args struct {
		req *orchestrator.GetMetricImplementationRequest
	}
	type fields struct {
		db persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[assessment.MetricImplementation]]
		wantErr assert.WantErr
	}{
		{
			name: "happy path",
			args: args{
				req: &orchestrator.GetMetricImplementationRequest{
					MetricId: orchestratortest.MockMetricImplementation1.MetricId,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockMetric1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockMetricImplementation1)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *connect.Response[assessment.MetricImplementation], args ...any) bool {
				assert.NotNil(t, got.Msg)
				return assert.Equal(t, orchestratortest.MockMetricImplementation1.MetricId, got.Msg.MetricId)
			},
			wantErr: assert.NoError,
		},
		{
			name: "validation error - missing id",
			args: args{
				req: &orchestrator.GetMetricImplementationRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[assessment.MetricImplementation]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsValidationError(t, err, "metric_id")
			},
		},
		{
			name: "not found",
			args: args{
				req: &orchestrator.GetMetricImplementationRequest{
					MetricId: orchestratortest.MockNonExistentId,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[assessment.MetricImplementation]],
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
			res, err := svc.GetMetricImplementation(context.Background(), connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}

func TestService_UpdateMetricImplementation(t *testing.T) {
	type args struct {
		req *orchestrator.UpdateMetricImplementationRequest
	}
	type fields struct {
		db persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[assessment.MetricImplementation]]
		wantErr assert.WantErr
	}{
		{
			name: "happy path",
			args: args{
				req: &orchestrator.UpdateMetricImplementationRequest{
					Implementation: &assessment.MetricImplementation{
						MetricId: orchestratortest.MockMetricImplementation1.MetricId,
						Lang:     assessment.MetricImplementation_LANGUAGE_REGO,
						Code:     "updated code",
					},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockMetric1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockMetricImplementation1)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *connect.Response[assessment.MetricImplementation], args ...any) bool {
				assert.NotNil(t, got.Msg)
				return assert.Equal(t, "updated code", got.Msg.Code)
			},
			wantErr: assert.NoError,
		},
		{
			name: "validation error - empty request",
			args: args{
				req: &orchestrator.UpdateMetricImplementationRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[assessment.MetricImplementation]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument)
			},
		},
		{
			name: "validation error - missing id",
			args: args{
				req: &orchestrator.UpdateMetricImplementationRequest{
					Implementation: &assessment.MetricImplementation{},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[assessment.MetricImplementation]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.IsValidationError(t, err, "implementation.metric_id")
			},
		},
		{
			name: "not found",
			args: args{
				req: &orchestrator.UpdateMetricImplementationRequest{
					Implementation: &assessment.MetricImplementation{
						MetricId: orchestratortest.MockNonExistentId,
						Lang:     assessment.MetricImplementation_LANGUAGE_REGO,
						Code:     "some code",
					},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[assessment.MetricImplementation]],
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
			res, err := svc.UpdateMetricImplementation(context.Background(), connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}

func TestService_GetMetricConfiguration(t *testing.T) {
	type args struct {
		req *orchestrator.GetMetricConfigurationRequest
	}
	type fields struct {
		db persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[assessment.MetricConfiguration]]
		wantErr assert.WantErr
	}{
		{
			name: "happy path",
			args: args{
				req: &orchestrator.GetMetricConfigurationRequest{
					TargetOfEvaluationId: orchestratortest.MockMetricConfiguration1.TargetOfEvaluationId,
					MetricId:             orchestratortest.MockMetricConfiguration1.MetricId,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					assert.NoError(t, d.Create(orchestratortest.MockTargetOfEvaluation1))
					assert.NoError(t, d.Create(orchestratortest.MockMetric1))
					assert.NoError(t, d.Create(orchestratortest.MockMetricConfiguration1))
				}),
			},
			want: func(t *testing.T, got *connect.Response[assessment.MetricConfiguration], args ...any) bool {
				assert.NotNil(t, got.Msg)
				return assert.Equal(t, orchestratortest.MockMetricConfiguration1.MetricId, got.Msg.MetricId)
			},
			wantErr: assert.NoError,
		},
		{
			name: "validation error - empty request",
			args: args{
				req: &orchestrator.GetMetricConfigurationRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[assessment.MetricConfiguration]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument)
			},
		},
		{
			name: "validation error - missing metric id",
			args: args{
				req: &orchestrator.GetMetricConfigurationRequest{
					TargetOfEvaluationId: orchestratortest.MockToeId1,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[assessment.MetricConfiguration]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsValidationError(t, err, "metric_id")
			},
		},
		{
			name: "not found",
			args: args{
				req: &orchestrator.GetMetricConfigurationRequest{
					TargetOfEvaluationId: orchestratortest.MockMetricConfiguration1.TargetOfEvaluationId,
					MetricId:             orchestratortest.MockNonExistentId,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[assessment.MetricConfiguration]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeNotFound)
			},
		},
		{
			name: "fallback to default configuration",
			args: args{
				req: &orchestrator.GetMetricConfigurationRequest{
					TargetOfEvaluationId: orchestratortest.MockToeId1,
					MetricId:             orchestratortest.MockMetricWithDefault.Id,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: func(t *testing.T, got *connect.Response[assessment.MetricConfiguration], args ...any) bool {
				if !assert.NotNil(t, got) {
					return false
				}
				assert.NotNil(t, got.Msg)
				assert.Equal(t, orchestratortest.MockMetricWithDefault.Id, got.Msg.MetricId)
				assert.Equal(t, orchestratortest.MockToeId1, got.Msg.TargetOfEvaluationId)
				assert.Equal(t, "==", got.Msg.Operator)
				assert.True(t, got.Msg.IsDefault)
				return assert.True(t, got.Msg.TargetValue.GetBoolValue())
			},
			wantErr: assert.NoError,
		},
	}

	// Set up a default configuration for the fallback test
	defaultMetricConfigurations[orchestratortest.MockMetricWithDefault.Id] = orchestratortest.MockMetricConfigurationDefault
	defer delete(defaultMetricConfigurations, orchestratortest.MockMetricWithDefault.Id)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: tt.fields.db,
			}
			res, err := svc.GetMetricConfiguration(context.Background(), connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}

func TestService_ListMetricConfigurations(t *testing.T) {
	type args struct {
		req *orchestrator.ListMetricConfigurationRequest
	}
	type fields struct {
		db persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[orchestrator.ListMetricConfigurationResponse]]
		wantErr assert.WantErr
	}{
		{
			name: "list all for TOE",
			args: args{
				req: &orchestrator.ListMetricConfigurationRequest{
					TargetOfEvaluationId: orchestratortest.MockMetricConfiguration1.TargetOfEvaluationId,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockTargetOfEvaluation1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockTargetOfEvaluation2)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockMetric1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockMetric2)
					assert.NoError(t, err)
					err = d.Create(&assessment.Metric{Id: "metric-3", Description: "Mock Metric 3"})
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockMetricConfiguration1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockMetricConfiguration2)
					assert.NoError(t, err)
					err = d.Create(&assessment.MetricConfiguration{
						TargetOfEvaluationId: orchestratortest.MockTargetOfEvaluation2.Id,
						MetricId:             "metric-3",
						IsDefault:            true,
					})
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.ListMetricConfigurationResponse], args ...any) bool {
				assert.NotNil(t, got)
				return assert.Equal(t, 2, len(got.Msg.Configurations))
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: tt.fields.db,
			}
			res, err := svc.ListMetricConfigurations(context.Background(), connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}

func TestService_UpdateMetricConfiguration(t *testing.T) {
	type args struct {
		req *orchestrator.UpdateMetricConfigurationRequest
	}
	type fields struct {
		db persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[assessment.MetricConfiguration]]
		wantErr assert.WantErr
	}{
		{
			name: "happy path",
			args: args{
				req: &orchestrator.UpdateMetricConfigurationRequest{
					Configuration: &assessment.MetricConfiguration{
						TargetOfEvaluationId: orchestratortest.MockToeId1,
						MetricId:             orchestratortest.MockMetricId1,
						Operator:             "!=",
						TargetValue:          structpb.NewBoolValue(false),
						IsDefault:            false,
					},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					// Create the TOE first (required by foreign key constraint)
					err := d.Create(orchestratortest.MockTargetOfEvaluation1)
					assert.NoError(t, err)
					// Create the metric (required by foreign key constraint)
					err = d.Create(orchestratortest.MockMetric1)
					assert.NoError(t, err)
					// Then create the configuration
					err = d.Create(orchestratortest.MockMetricConfiguration1)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *connect.Response[assessment.MetricConfiguration], args ...any) bool {
				return assert.Equal(t, orchestratortest.MockToeId1, got.Msg.TargetOfEvaluationId) &&
					assert.Equal(t, orchestratortest.MockMetricId1, got.Msg.MetricId) &&
					assert.Equal(t, "!=", got.Msg.Operator) &&
					assert.False(t, got.Msg.IsDefault)
			},
			wantErr: assert.NoError,
		},
		{
			name: "validation error - empty request",
			args: args{
				req: &orchestrator.UpdateMetricConfigurationRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[assessment.MetricConfiguration]],
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
			res, err := svc.UpdateMetricConfiguration(context.Background(), connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}

func TestService_loadMetrics(t *testing.T) {
	type fields struct {
		db  persistence.DB
		cfg Config
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr assert.WantErr
	}{
		{
			name: "no metrics to load",
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
				cfg: Config{
					LoadDefaultMetrics: false,
					LoadMetricsFunc:    nil,
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "load from custom function",
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
				cfg: Config{
					LoadDefaultMetrics: false,
					LoadMetricsFunc: func(svc *Service) ([]*assessment.Metric, error) {
						return []*assessment.Metric{
							{
								Id:          "custom-metric",
								Description: "Custom Metric",
							},
						}, nil
					},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "custom function returns error",
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
				cfg: Config{
					LoadDefaultMetrics: false,
					LoadMetricsFunc: func(svc *Service) ([]*assessment.Metric, error) {
						return nil, errors.New("custom error")
					},
				},
			},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.Error(t, err) &&
					assert.ErrorContains(t, err, "could not load additional metrics")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db:  tt.fields.db,
				cfg: tt.fields.cfg,
			}
			err := svc.loadMetrics()
			tt.wantErr(t, err)
		})
	}
}

func TestService_loadMetricsFromRepository(t *testing.T) {
	type fields struct {
		db  persistence.DB
		cfg Config
	}
	tests := []struct {
		name        string
		fields      fields
		wantMetrics int
		wantErr     assert.WantErr
	}{
		{
			name: "directory does not exist",
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
				cfg: Config{
					DefaultMetricsPath: "/nonexistent/path",
				},
			},
			wantMetrics: 0,
			wantErr:     assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db:  tt.fields.db,
				cfg: tt.fields.cfg,
			}
			metrics, err := svc.loadMetricsFromRepository()
			tt.wantErr(t, err)
			assert.Equal(t, tt.wantMetrics, len(metrics))
		})
	}
}

func Test_prepareMetric(t *testing.T) {
	type args struct {
		m          *assessment.Metric
		metricPath string
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.WantErr
	}{
		{
			name: "no data.json file",
			args: args{
				m: &assessment.Metric{
					Id:          "test-metric",
					Description: "Test Metric",
				},
				metricPath: "/tmp/nonexistent/metric.yaml",
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := prepareMetric(tt.args.m, tt.args.metricPath)
			tt.wantErr(t, err)
		})
	}
}
