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

func TestService_CreateMetric(t *testing.T) {
	type fields struct {
		db *persistence.DB
	}
	type args struct {
		req *orchestrator.CreateMetricRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    assert.Want[*assessment.Metric]
		wantErr assert.WantErr[*connect.Error]
	}{
		{
			name: "happy path",
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			args: args{
				req: &orchestrator.CreateMetricRequest{
					Metric: orchestratortest.MockMetric1,
				},
			},
			want: func(t *testing.T, got *assessment.Metric, args ...any) bool {
				return assert.Equal(t, orchestratortest.MockMetric1.Id, got.Id) &&
					assert.Equal(t, orchestratortest.MockMetric1.Description, got.Description)
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: tt.fields.db,
			}
			res, err := svc.CreateMetric(context.Background(), connect.NewRequest(tt.args.req))
			assert.WantResponse(t, res, err, tt.want, tt.wantErr)
		})
	}
}

func TestService_GetMetric(t *testing.T) {
	type args struct {
		req *orchestrator.GetMetricRequest
	}
	type fields struct {
		db *persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*assessment.Metric]
		wantErr assert.WantErr[*connect.Error]
	}{
		{
			name: "happy path",
			args: args{
				req: &orchestrator.GetMetricRequest{
					MetricId: orchestratortest.MockMetric1.Id,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d *persistence.DB) {
					err := d.Create(orchestratortest.MockMetric1)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *assessment.Metric, args ...any) bool {
				return assert.NotNil(t, got) &&
					assert.Equal(t, orchestratortest.MockMetric1.Id, got.Id)
			},
			wantErr: assert.Nil[*connect.Error],
		},
		{
			name: "not found",
			args: args{
				req: &orchestrator.GetMetricRequest{
					MetricId: "non-existent",
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*assessment.Metric],
			wantErr: func(t *testing.T, err *connect.Error, msgAndArgs ...any) bool {
				return assert.Equal(t, connect.CodeNotFound, err.Code())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: tt.fields.db,
			}
			res, err := svc.GetMetric(context.Background(), connect.NewRequest(tt.args.req))
			assert.WantResponse(t, res, err, tt.want, tt.wantErr)
		})
	}
}

func TestService_ListMetrics(t *testing.T) {
	type args struct {
		req *orchestrator.ListMetricsRequest
	}
	type fields struct {
		db *persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*orchestrator.ListMetricsResponse]
		wantErr assert.WantErr[*connect.Error]
	}{
		{
			name: "list all",
			args: args{
				req: &orchestrator.ListMetricsRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d *persistence.DB) {
					err := d.Create(orchestratortest.MockMetric1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockMetric2)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *orchestrator.ListMetricsResponse, args ...any) bool {
				return assert.NotNil(t, got) &&
					assert.Equal(t, 2, len(got.Metrics))
			},
			wantErr: nil,
		},
		{
			name: "empty list",
			args: args{
				req: &orchestrator.ListMetricsRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: func(t *testing.T, got *orchestrator.ListMetricsResponse, args ...any) bool {
				return assert.NotNil(t, got) &&
					assert.Equal(t, 0, len(got.Metrics))
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: tt.fields.db,
			}
			res, err := svc.ListMetrics(context.Background(), connect.NewRequest(tt.args.req))
			assert.WantResponse(t, res, err, tt.want, tt.wantErr)
		})
	}
}

func TestService_UpdateMetric(t *testing.T) {
	type args struct {
		req *orchestrator.UpdateMetricRequest
	}
	type fields struct {
		db *persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*assessment.Metric]
		wantErr assert.WantErr[*connect.Error]
	}{
		{
			name: "happy path",
			args: args{
				req: &orchestrator.UpdateMetricRequest{
					Metric: &assessment.Metric{
						Id:          orchestratortest.MockMetric1.Id,
						Description: "Updated description",
					},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d *persistence.DB) {
					err := d.Create(orchestratortest.MockMetric1)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *assessment.Metric, args ...any) bool {
				return assert.NotNil(t, got) &&
					assert.Equal(t, "Updated description", got.Description)
			},
			wantErr: nil,
		},
		{
			name: "not found",
			args: args{
				req: &orchestrator.UpdateMetricRequest{
					Metric: &assessment.Metric{
						Id:          "non-existent",
						Description: "Updated description",
					},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: nil,
			wantErr: func(t *testing.T, err *connect.Error, msgAndArgs ...any) bool {
				return assert.Equal(t, connect.CodeNotFound, err.Code())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: tt.fields.db,
			}
			res, err := svc.UpdateMetric(context.Background(), connect.NewRequest(tt.args.req))
			assert.WantResponse(t, res, err, tt.want, tt.wantErr)
		})
	}
}

func TestService_RemoveMetric(t *testing.T) {
	type args struct {
		req *orchestrator.RemoveMetricRequest
	}
	type fields struct {
		db *persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*emptypb.Empty]
		wantErr assert.WantErr[*connect.Error]
	}{
		{
			name: "happy path",
			args: args{
				req: &orchestrator.RemoveMetricRequest{
					MetricId: orchestratortest.MockMetric1.Id,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d *persistence.DB) {
					err := d.Create(orchestratortest.MockMetric1)
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
			res, err := svc.RemoveMetric(context.Background(), connect.NewRequest(tt.args.req))
			assert.WantResponse(t, res, err, tt.want, tt.wantErr)
		})
	}
}

func TestService_GetMetricImplementation(t *testing.T) {
	type args struct {
		req *orchestrator.GetMetricImplementationRequest
	}
	type fields struct {
		db *persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*assessment.MetricImplementation]
		wantErr assert.WantErr[*connect.Error]
	}{
		{
			name: "happy path",
			args: args{
				req: &orchestrator.GetMetricImplementationRequest{
					MetricId: orchestratortest.MockMetricImplementation1.MetricId,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d *persistence.DB) {
					err := d.Create(orchestratortest.MockMetric1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockMetricImplementation1)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *assessment.MetricImplementation, args ...any) bool {
				return assert.NotNil(t, got) &&
					assert.Equal(t, orchestratortest.MockMetricImplementation1.MetricId, got.MetricId)
			},
			wantErr: nil,
		},
		{
			name: "not found",
			args: args{
				req: &orchestrator.GetMetricImplementationRequest{
					MetricId: "non-existent",
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: nil,
			wantErr: func(t *testing.T, err *connect.Error, msgAndArgs ...any) bool {
				return assert.Equal(t, connect.CodeNotFound, err.Code())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: tt.fields.db,
			}
			res, err := svc.GetMetricImplementation(context.Background(), connect.NewRequest(tt.args.req))
			assert.WantResponse(t, res, err, tt.want, tt.wantErr)
		})
	}
}

func TestService_UpdateMetricImplementation(t *testing.T) {
	type args struct {
		req *orchestrator.UpdateMetricImplementationRequest
	}
	type fields struct {
		db *persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*assessment.MetricImplementation]
		wantErr assert.WantErr[*connect.Error]
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
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d *persistence.DB) {
					err := d.Create(orchestratortest.MockMetric1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockMetricImplementation1)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *assessment.MetricImplementation, args ...any) bool {
				return assert.NotNil(t, got) &&
					assert.Equal(t, "updated code", got.Code)
			},
			wantErr: nil,
		},
		{
			name: "not found",
			args: args{
				req: &orchestrator.UpdateMetricImplementationRequest{
					Implementation: &assessment.MetricImplementation{
						MetricId: "non-existent",
						Lang:     assessment.MetricImplementation_LANGUAGE_REGO,
					},
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: nil,
			wantErr: func(t *testing.T, err *connect.Error, msgAndArgs ...any) bool {
				return assert.Equal(t, connect.CodeNotFound, err.Code())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: tt.fields.db,
			}
			res, err := svc.UpdateMetricImplementation(context.Background(), connect.NewRequest(tt.args.req))
			assert.WantResponse(t, res, err, tt.want, tt.wantErr)
		})
	}
}

func TestService_GetMetricConfiguration(t *testing.T) {
	type args struct {
		req *orchestrator.GetMetricConfigurationRequest
	}
	type fields struct {
		db *persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*assessment.MetricConfiguration]
		wantErr assert.WantErr[*connect.Error]
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
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d *persistence.DB) {
					err := d.Create(orchestratortest.MockTargetOfEvaluation1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockMetric1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockMetricConfiguration1)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *assessment.MetricConfiguration, args ...any) bool {
				return assert.NotNil(t, got) &&
					assert.Equal(t, orchestratortest.MockMetricConfiguration1.MetricId, got.MetricId)
			},
			wantErr: nil,
		},
		{
			name: "not found",
			args: args{
				req: &orchestrator.GetMetricConfigurationRequest{
					TargetOfEvaluationId: orchestratortest.MockMetricConfiguration1.TargetOfEvaluationId,
					MetricId:             "non-existent",
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: nil,
			wantErr: func(t *testing.T, err *connect.Error, msgAndArgs ...any) bool {
				return assert.Equal(t, connect.CodeNotFound, err.Code())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: tt.fields.db,
			}
			res, err := svc.GetMetricConfiguration(context.Background(), connect.NewRequest(tt.args.req))
			assert.WantResponse(t, res, err, tt.want, tt.wantErr)
		})
	}
}

func TestService_ListMetricConfigurations(t *testing.T) {
	type args struct {
		req *orchestrator.ListMetricConfigurationRequest
	}
	type fields struct {
		db *persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*orchestrator.ListMetricConfigurationResponse]
		wantErr assert.WantErr[*connect.Error]
	}{
		{
			name: "list all for TOE",
			args: args{
				req: &orchestrator.ListMetricConfigurationRequest{
					TargetOfEvaluationId: orchestratortest.MockMetricConfiguration1.TargetOfEvaluationId,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d *persistence.DB) {
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
			want: func(t *testing.T, got *orchestrator.ListMetricConfigurationResponse, args ...any) bool {
				return assert.NotNil(t, got) &&
					assert.Equal(t, 2, len(got.Configurations))
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: tt.fields.db,
			}
			res, err := svc.ListMetricConfigurations(context.Background(), connect.NewRequest(tt.args.req))
			assert.WantResponse(t, res, err, tt.want, tt.wantErr)
		})
	}
}

/*
func TestService_UpdateMetricConfiguration(t *testing.T) {
	var (
		tests = []struct {
			name    string
			req     *orchestrator.UpdateMetricConfigurationRequest
			setup   func(*service)
			wantErr bool
		}{
			{
				name: "happy path",
				req: &orchestrator.UpdateMetricConfigurationRequest{
					TargetOfEvaluationId: "toe-1",
					MetricId:             "metric-1",
					Configuration: &assessment.MetricConfiguration{
						TargetOfEvaluationId: "toe-1",
						MetricId:             "metric-1",
						IsDefault:            false,
					},
				},
				setup: func(svc *service) {
					err := svc.db.Create(&assessment.MetricConfiguration{
						TargetOfEvaluationId: "toe-1",
						MetricId:             "metric-1",
						IsDefault:            true,
					})
					assert.NoError(t, err)
				},
				wantErr: false,
			},
			{
				name: "not found",
				req: &orchestrator.UpdateMetricConfigurationRequest{
					TargetOfEvaluationId: "toe-1",
					MetricId:             "non-existent",
					Configuration: &assessment.MetricConfiguration{
						TargetOfEvaluationId: "toe-1",
						MetricId:             "non-existent",
						IsDefault:            false,
					},
				},
				setup:   func(svc *service) {},
				wantErr: true,
			},
		}
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				svc = &service{
					db: persistencetest.NewInMemoryDB(t, types, joinTables),
				}
				req = connect.NewRequest(tt.req)
			)

			tt.setup(svc)

			res, err := svc.UpdateMetricConfiguration(context.Background(), req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, res)
			assert.Equal(t, tt.req.Configuration.IsEnabled, res.Msg.IsEnabled)
		})
	}
}
*/
