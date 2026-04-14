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
	"os"
	"testing"

	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/auth"
	"confirmate.io/core/persistence"
	"confirmate.io/core/persistence/persistencetest"
	"confirmate.io/core/service"
	"confirmate.io/core/service/orchestrator/orchestratortest"
	"confirmate.io/core/util/assert"
	"confirmate.io/core/util/clitest"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/go-cmp/cmp"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestMain(m *testing.M) {
	clitest.AutoChdir()
	code := m.Run()
	os.Exit(code)
}

func TestService_CreateMetric(t *testing.T) {
	type args struct {
		req *orchestrator.CreateMetricRequest
		ctx context.Context
	}
	type fields struct {
		db    persistence.DB
		authz service.AuthorizationStrategy
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[assessment.Metric]]
		wantErr assert.WantErr
	}{
		{
			name: "err: request validation error",
			args: args{
				req: nil,
			},
			want: assert.Nil[*connect.Response[assessment.Metric]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.ErrorContains(t, err, "empty request")
			},
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
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.ErrorContains(t, err, "invalid request")
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
				db:    persistencetest.CreateErrorDB(t, persistence.ErrUniqueConstraintFailed, types, joinTables),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: assert.Nil[*connect.Response[assessment.Metric]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeAlreadyExists)
			},
		},
		{
			name: "authorization error",
			args: args{
				req: &orchestrator.CreateMetricRequest{
					Metric: orchestratortest.MockMetric1,
				},
			},
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables),
				authz: &service.AuthorizationStrategyPermissionStore{},
			},
			want: assert.Nil[*connect.Response[assessment.Metric]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodePermissionDenied)
			},
		},
		{
			name: "happy path: with allow-all authorization strategy",
			args: args{
				req: &orchestrator.CreateMetricRequest{
					Metric: orchestratortest.MockMetric1,
				},
			},
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: func(t *testing.T, got *connect.Response[assessment.Metric], args ...any) bool {
				assert.NotNil(t, got.Msg)
				return assert.Equal(t, orchestratortest.MockMetric1, got.Msg, cmp.Options{
					protocmp.IgnoreFields(&assessment.Metric{}, "id"),
				}) &&
					assert.NotEmpty(t, got.Msg.Id)
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: with authorization strategy with permission store and admin token",
			args: args{
				req: &orchestrator.CreateMetricRequest{
					Metric: orchestratortest.MockMetric1,
				},
				ctx: auth.WithClaims(context.Background(), &auth.OAuthClaims{
					IsAdminToken: true,
				}),
			},
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables),
				authz: &service.AuthorizationStrategyPermissionStore{},
			},
			want: func(t *testing.T, got *connect.Response[assessment.Metric], args ...any) bool {
				assert.NotNil(t, got.Msg)
				return assert.Equal(t, orchestratortest.MockMetric1, got.Msg, cmp.Options{
					protocmp.IgnoreFields(&assessment.Metric{}, "id"),
				}) &&
					assert.NotEmpty(t, got.Msg.Id)
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
			res, err := svc.CreateMetric(tt.args.ctx, connect.NewRequest(tt.args.req))
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
		db    persistence.DB
		authz service.AuthorizationStrategy
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
				return assert.Equal(t, orchestratortest.MockMetric1, got.Msg)
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
			name: "err: request validation error",
			args: args{
				req: nil,
			},
			want: assert.Nil[*connect.Response[orchestrator.ListMetricsResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.ErrorContains(t, err, "empty request")
			},
		},
		{
			name: "err: db error",
			args: args{
				req: &orchestrator.ListMetricsRequest{},
			},
			fields: fields{
				db: persistencetest.ListErrorDB(t, persistence.ErrRecordNotFound, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.ListMetricsResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeNotFound) &&
					errors.Is(err, persistence.ErrRecordNotFound)
			},
		},
		{
			name: "happy path: list all metrics without deprecated metrics",
			args: args{
				req: &orchestrator.ListMetricsRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockMetric1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockMetric2)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockMetricDeprecated)
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
			name: "happy path: empty list",
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
		ctx context.Context
	}
	type fields struct {
		db    persistence.DB
		authz service.AuthorizationStrategy
	}

	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[assessment.Metric]]
		wantErr assert.WantErr
	}{
		{
			name: "happy path: with allow-all authorization strategy",
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
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: func(t *testing.T, got *connect.Response[assessment.Metric], args ...any) bool {
				assert.NotNil(t, got.Msg)
				assert.Equal(t, orchestratortest.MockMetric1, got.Msg, cmp.Options{
					protocmp.IgnoreFields(&assessment.Metric{}, "description"),
				})
				return assert.Equal(t, "Updated description", got.Msg.Description)
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: with authorization strategy with permission store and admin token",
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
				ctx: auth.WithClaims(context.Background(), &auth.OAuthClaims{
					IsAdminToken: true,
				}),
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockMetric1)
					assert.NoError(t, err)
				}),
				authz: &service.AuthorizationStrategyPermissionStore{},
			},
			want: func(t *testing.T, got *connect.Response[assessment.Metric], args ...any) bool {
				assert.NotNil(t, got.Msg)
				assert.Equal(t, orchestratortest.MockMetric1, got.Msg, cmp.Options{
					protocmp.IgnoreFields(&assessment.Metric{}, "description"),
				})
				return assert.Equal(t, "Updated description", got.Msg.Description)
			},
			wantErr: assert.NoError,
		},
		{
			name: "authorization error",
			args: args{
				req: &orchestrator.UpdateMetricRequest{
					Metric: orchestratortest.MockMetric1,
				},
			},
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables),
				authz: &service.AuthorizationStrategyPermissionStore{},
			},
			want: assert.Nil[*connect.Response[assessment.Metric]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodePermissionDenied)
			},
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
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.ErrorContains(t, err, "invalid request")
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
				db:    persistencetest.NewInMemoryDB(t, types, joinTables),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: assert.Nil[*connect.Response[assessment.Metric]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeNotFound) &&
					assert.ErrorContains(t, err, "metric not found")
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
				db:    persistencetest.UpdateErrorDB(t, persistence.ErrConstraintFailed, types, joinTables),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: assert.Nil[*connect.Response[assessment.Metric]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					errors.Is(err, persistence.ErrConstraintFailed)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db:    tt.fields.db,
				authz: tt.fields.authz,
			}
			res, err := svc.UpdateMetric(tt.args.ctx, connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}

func TestService_RemoveMetric(t *testing.T) {
	type args struct {
		req *orchestrator.RemoveMetricRequest
		ctx context.Context
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
				req: &orchestrator.RemoveMetricRequest{
					MetricId: orchestratortest.MockMetric1.Id,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockMetric1)
					assert.NoError(t, err)
				}),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: func(t *testing.T, got *connect.Response[emptypb.Empty], args ...any) bool {
				return assert.Empty(t, got.Msg)
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: with authorization strategy with permission store and admin token",
			args: args{
				req: &orchestrator.RemoveMetricRequest{
					MetricId: orchestratortest.MockMetric1.Id,
				},
				ctx: auth.WithClaims(context.Background(), &auth.OAuthClaims{
					IsAdminToken: true,
				}),
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockMetric1)
					assert.NoError(t, err)
				}),
				authz: &service.AuthorizationStrategyPermissionStore{},
			},
			want: func(t *testing.T, got *connect.Response[emptypb.Empty], args ...any) bool {
				return assert.Empty(t, got.Msg)
			},
			wantErr: assert.NoError,
		},
		{
			name: "authorization error",
			args: args{
				req: &orchestrator.RemoveMetricRequest{
					MetricId: orchestratortest.MockMetric1.Id,
				},
			},
			fields: fields{
				db:    persistencetest.NewInMemoryDB(t, types, joinTables),
				authz: &service.AuthorizationStrategyPermissionStore{},
			},
			want: assert.Nil[*connect.Response[emptypb.Empty]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodePermissionDenied)
			},
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
			name: "db error - GET",
			args: args{
				req: &orchestrator.RemoveMetricRequest{
					MetricId: orchestratortest.MockMetric1.Id,
				},
			},
			fields: fields{
				db:    persistencetest.GetErrorDB(t, persistence.ErrRecordNotFound, types, joinTables),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: assert.Nil[*connect.Response[emptypb.Empty]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeNotFound) &&
					errors.Is(err, persistence.ErrRecordNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db:    tt.fields.db,
				authz: tt.fields.authz,
			}
			res, err := svc.RemoveMetric(tt.args.ctx, connect.NewRequest(tt.args.req))
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
				return assert.Equal(t, orchestratortest.MockMetricImplementation1, got.Msg)
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
		ctx context.Context
	}
	type fields struct {
		db    persistence.DB
		authz service.AuthorizationStrategy
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[assessment.MetricImplementation]]
		wantErr assert.WantErr
	}{
		{
			name: "happy path: with allow-all authorization strategy",
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
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: func(t *testing.T, got *connect.Response[assessment.MetricImplementation], args ...any) bool {
				assert.NotNil(t, got.Msg)
				assert.Equal(t, orchestratortest.MockMetricImplementation1, got.Msg, cmp.Options{
					protocmp.IgnoreFields(&assessment.MetricImplementation{}, "code", "updated_at"),
				})
				return assert.Equal(t, "updated code", got.Msg.Code)
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path: with authorization strategy with permission store and admin token",
			args: args{
				req: &orchestrator.UpdateMetricImplementationRequest{
					Implementation: &assessment.MetricImplementation{
						MetricId: orchestratortest.MockMetricImplementation1.MetricId,
						Lang:     assessment.MetricImplementation_LANGUAGE_REGO,
						Code:     "updated code",
					},
				},
				ctx: auth.WithClaims(context.Background(), &auth.OAuthClaims{
					IsAdminToken: true,
				}),
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					err := d.Create(orchestratortest.MockMetric1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockMetricImplementation1)
					assert.NoError(t, err)
				}),
				authz: &service.AuthorizationStrategyPermissionStore{},
			},
			want: func(t *testing.T, got *connect.Response[assessment.MetricImplementation], args ...any) bool {
				assert.NotNil(t, got.Msg)
				assert.Equal(t, orchestratortest.MockMetricImplementation1, got.Msg, cmp.Options{
					protocmp.IgnoreFields(&assessment.MetricImplementation{}, "code", "updated_at"),
				})
				return assert.Equal(t, "updated code", got.Msg.Code)
			},
			wantErr: assert.NoError,
		},
		{
			name: "authorization error",
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
				db:    persistencetest.NewInMemoryDB(t, types, joinTables),
				authz: &service.AuthorizationStrategyPermissionStore{},
			},
			want: assert.Nil[*connect.Response[assessment.MetricImplementation]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodePermissionDenied)
			},
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
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.ErrorContains(t, err, "invalid request")
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
				db:    persistencetest.NewInMemoryDB(t, types, joinTables),
				authz: &service.AuthorizationStrategyAllowAll{},
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
				db:    tt.fields.db,
				authz: tt.fields.authz,
			}
			res, err := svc.UpdateMetricImplementation(tt.args.ctx, connect.NewRequest(tt.args.req))
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
				return assert.Equal(t, orchestratortest.MockMetricConfiguration1, got.Msg)
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
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.ErrorContains(t, err, "invalid request")
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
			name: "validation error - empty request",
			args: args{
				req: nil,
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.ListMetricConfigurationResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.ErrorContains(t, err, "empty request")
			},
		},
		{
			name: "err: db error",
			args: args{
				req: &orchestrator.ListMetricConfigurationRequest{
					TargetOfEvaluationId: orchestratortest.MockToeId1,
				},
			},
			fields: fields{
				db: persistencetest.ListErrorDB(t, persistence.ErrRecordNotFound, types, joinTables),
			},
			want: assert.Nil[*connect.Response[orchestrator.ListMetricConfigurationResponse]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeNotFound) &&
					errors.Is(err, persistence.ErrRecordNotFound)
			},
		},
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
					err = d.Create(orchestratortest.MockMetric4)
					assert.NoError(t, err)
					err = d.Create(&assessment.Metric{Id: "metric-3", Description: "Mock Metric 3"})
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockMetricConfiguration1)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockMetricConfiguration2)
					assert.NoError(t, err)
					err = d.Create(orchestratortest.MockMetricConfiguration4)
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
		req     *orchestrator.UpdateMetricConfigurationRequest
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
		want    assert.Want[*connect.Response[assessment.MetricConfiguration]]
		wantErr assert.WantErr
	}{
		{
			name: "happy path: with allow-all authorization strategy",
			args: args{
				req: &orchestrator.UpdateMetricConfigurationRequest{
					Configuration: &assessment.MetricConfiguration{
						TargetOfEvaluationId: orchestratortest.MockToeId1,
						MetricId:             orchestratortest.MockMetricId1,
						Operator:             "!=", // updates the operator from "==" to "!="
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
				authz: &service.AuthorizationStrategyAllowAll{},
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
			name: "happy path: with authorization strategy with permission store and admin token",
			args: args{
				req: &orchestrator.UpdateMetricConfigurationRequest{
					Configuration: &assessment.MetricConfiguration{
						TargetOfEvaluationId: orchestratortest.MockToeId1,
						MetricId:             orchestratortest.MockMetricId1,
						Operator:             "!=", // updates the operator from "==" to "!="
						TargetValue:          structpb.NewBoolValue(false),
						IsDefault:            false,
					},
				},
				context: auth.WithClaims(context.Background(), &auth.OAuthClaims{
					IsAdminToken: true,
				}),
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
				authz: &service.AuthorizationStrategyPermissionStore{},
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
			name: "happy path: with authorization strategy with permission store and user permissions allowing access",
			args: args{
				req: &orchestrator.UpdateMetricConfigurationRequest{
					Configuration: &assessment.MetricConfiguration{
						TargetOfEvaluationId: orchestratortest.MockToeId1,
						MetricId:             orchestratortest.MockMetricId1,
						Operator:             "!=", // updates the operator from "==" to "!="
						TargetValue:          structpb.NewBoolValue(false),
						IsDefault:            false,
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
				authz: &service.AuthorizationStrategyPermissionStore{
					Permissions: permissionStore{
						db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
							err := d.Create(orchestratortest.MockUserPermissionsToEAdmin)
							assert.NoError(t, err)
						}),
					},
				},
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
			name: "authorization failure",
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
				db:    persistencetest.NewInMemoryDB(t, types, joinTables),
				authz: &denyAuthorizationStrategy{},
			},
			want: assert.Nil[*connect.Response[assessment.MetricConfiguration]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodePermissionDenied)
			},
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
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					assert.ErrorContains(t, err, "invalid request")
			},
		},
		{
			name: "error - db error on update",
			args: args{
				req: &orchestrator.UpdateMetricConfigurationRequest{
					Configuration: &assessment.MetricConfiguration{
						TargetOfEvaluationId: orchestratortest.MockToeId1,
						MetricId:             orchestratortest.MockMetricId1,
						Operator:             "!=",
						TargetValue:          structpb.NewBoolValue(false),
					},
				},
			},
			fields: fields{
				db:    persistencetest.SaveErrorDB(t, persistence.ErrConstraintFailed, types, joinTables),
				authz: &service.AuthorizationStrategyAllowAll{},
			},
			want: assert.Nil[*connect.Response[assessment.MetricConfiguration]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsConnectError(t, err, connect.CodeInvalidArgument) &&
					errors.Is(err, persistence.ErrConstraintFailed)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db:    tt.fields.db,
				authz: tt.fields.authz,
			}
			res, err := svc.UpdateMetricConfiguration(tt.args.context, connect.NewRequest(tt.args.req))
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
		{
			name: "load default",
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
				cfg: Config{
					LoadDefaultMetrics: true,
					DefaultMetricsPath: "./policies/security-metrics/metrics",
				},
			},
			wantErr: assert.NoError,
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
