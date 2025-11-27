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
	"confirmate.io/core/persistence/persistencetest"
	"confirmate.io/core/util/assert"

	"connectrpc.com/connect"
)

func TestService_CreateMetric(t *testing.T) {
	var (
		tests = []struct {
			name    string
			req     *orchestrator.CreateMetricRequest
			wantErr bool
		}{
			{
				name: "happy path",
				req: &orchestrator.CreateMetricRequest{
					Metric: &assessment.Metric{
						Id:          "metric-1",
						Description: "A test metric",
					},
				},
				wantErr: false,
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

			res, err := svc.CreateMetric(context.Background(), req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, res)
			assert.Equal(t, tt.req.Metric.Id, res.Msg.Id)
		})
	}
}

func TestService_GetMetric(t *testing.T) {
	var (
		tests = []struct {
			name    string
			id      string
			setup   func(*service)
			wantErr bool
		}{
			{
				name: "happy path",
				id:   "metric-1",
				setup: func(svc *service) {
					err := svc.db.Create(&assessment.Metric{
						Id:          "metric-1",
						Description: "A test metric",
					})
					assert.NoError(t, err)
				},
				wantErr: false,
			},
			{
				name:    "not found",
				id:      "non-existent",
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
				req = connect.NewRequest(&orchestrator.GetMetricRequest{
					MetricId: tt.id,
				})
			)

			tt.setup(svc)

			res, err := svc.GetMetric(context.Background(), req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, res)
			assert.Equal(t, tt.id, res.Msg.Id)
		})
	}
}

func TestService_ListMetrics(t *testing.T) {
	var (
		tests = []struct {
			name      string
			setup     func(*service)
			wantCount int
		}{
			{
				name: "list all",
				setup: func(svc *service) {
					err := svc.db.Create(&assessment.Metric{
						Id:          "metric-1",
						Description: "First metric",
					})
					assert.NoError(t, err)

					err = svc.db.Create(&assessment.Metric{
						Id:          "metric-2",
						Description: "Second metric",
					})
					assert.NoError(t, err)
				},
				wantCount: 2,
			},
			{
				name:      "empty list",
				setup:     func(svc *service) {},
				wantCount: 0,
			},
		}
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				svc = &service{
					db: persistencetest.NewInMemoryDB(t, types, joinTables),
				}
				req = connect.NewRequest(&orchestrator.ListMetricsRequest{})
			)

			tt.setup(svc)

			res, err := svc.ListMetrics(context.Background(), req)

			assert.NoError(t, err)
			assert.NotNil(t, res)
			assert.Equal(t, tt.wantCount, len(res.Msg.Metrics))
		})
	}
}

func TestService_UpdateMetric(t *testing.T) {
	var (
		tests = []struct {
			name    string
			req     *orchestrator.UpdateMetricRequest
			setup   func(*service)
			wantErr bool
		}{
			{
				name: "happy path",
				req: &orchestrator.UpdateMetricRequest{
					Metric: &assessment.Metric{
						Id:          "metric-1",
						Description: "Updated description",
					},
				},
				setup: func(svc *service) {
					err := svc.db.Create(&assessment.Metric{
						Id:          "metric-1",
						Description: "Original description",
					})
					assert.NoError(t, err)
				},
				wantErr: false,
			},
			{
				name: "not found",
				req: &orchestrator.UpdateMetricRequest{
					Metric: &assessment.Metric{
						Id:          "non-existent",
						Description: "Updated description",
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

			res, err := svc.UpdateMetric(context.Background(), req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, res)
			assert.Equal(t, tt.req.Metric.Description, res.Msg.Description)
		})
	}
}

func TestService_RemoveMetric(t *testing.T) {
	var (
		tests = []struct {
			name    string
			id      string
			setup   func(*service)
			wantErr bool
		}{
			{
				name: "happy path",
				id:   "metric-1",
				setup: func(svc *service) {
					err := svc.db.Create(&assessment.Metric{
						Id:          "metric-1",
						Description: "A test metric",
					})
					assert.NoError(t, err)
				},
				wantErr: false,
			},
		}
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				svc = &service{
					db: persistencetest.NewInMemoryDB(t, types, joinTables),
				}
				req = connect.NewRequest(&orchestrator.RemoveMetricRequest{
					MetricId: tt.id,
				})
			)

			tt.setup(svc)

			res, err := svc.RemoveMetric(context.Background(), req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, res)
		})
	}
}

func TestService_GetMetricImplementation(t *testing.T) {
	var (
		tests = []struct {
			name     string
			metricId string
			setup    func(*service)
			wantErr  bool
		}{
			{
				name:     "happy path",
				metricId: "metric-1",
				setup: func(svc *service) {
					err := svc.db.Create(&assessment.MetricImplementation{
						MetricId: "metric-1",
						Lang:     assessment.MetricImplementation_LANGUAGE_REGO,
					})
					assert.NoError(t, err)
				},
				wantErr: false,
			},
			{
				name:     "not found",
				metricId: "non-existent",
				setup:    func(svc *service) {},
				wantErr:  true,
			},
		}
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				svc = &service{
					db: persistencetest.NewInMemoryDB(t, types, joinTables),
				}
				req = connect.NewRequest(&orchestrator.GetMetricImplementationRequest{
					MetricId: tt.metricId,
				})
			)

			tt.setup(svc)

			res, err := svc.GetMetricImplementation(context.Background(), req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, res)
			assert.Equal(t, tt.metricId, res.Msg.MetricId)
		})
	}
}

func TestService_UpdateMetricImplementation(t *testing.T) {
	var (
		tests = []struct {
			name    string
			req     *orchestrator.UpdateMetricImplementationRequest
			setup   func(*service)
			wantErr bool
		}{
			{
				name: "happy path",
				req: &orchestrator.UpdateMetricImplementationRequest{
					Implementation: &assessment.MetricImplementation{
						MetricId: "metric-1",
						Lang:     assessment.MetricImplementation_LANGUAGE_REGO,
						Code:     "updated code",
					},
				},
				setup: func(svc *service) {
					err := svc.db.Create(&assessment.MetricImplementation{
						MetricId: "metric-1",
						Lang:     assessment.MetricImplementation_LANGUAGE_REGO,
						Code:     "original code",
					})
					assert.NoError(t, err)
				},
				wantErr: false,
			},
			{
				name: "not found",
				req: &orchestrator.UpdateMetricImplementationRequest{
					Implementation: &assessment.MetricImplementation{
						MetricId: "non-existent",
						Lang:     assessment.MetricImplementation_LANGUAGE_REGO,
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

			res, err := svc.UpdateMetricImplementation(context.Background(), req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, res)
			assert.Equal(t, tt.req.Implementation.Code, res.Msg.Code)
		})
	}
}

func TestService_GetMetricConfiguration(t *testing.T) {
	var (
		tests = []struct {
			name    string
			toeId   string
			metricId string
			setup   func(*service)
			wantErr bool
		}{
			{
				name:     "happy path",
				toeId:    "toe-1",
				metricId: "metric-1",
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
				name:     "not found",
				toeId:    "toe-1",
				metricId: "non-existent",
				setup:    func(svc *service) {},
				wantErr:  true,
			},
		}
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				svc = &service{
					db: persistencetest.NewInMemoryDB(t, types, joinTables),
				}
				req = connect.NewRequest(&orchestrator.GetMetricConfigurationRequest{
					TargetOfEvaluationId: tt.toeId,
					MetricId:             tt.metricId,
				})
			)

			tt.setup(svc)

			res, err := svc.GetMetricConfiguration(context.Background(), req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, res)
			assert.Equal(t, tt.metricId, res.Msg.MetricId)
		})
	}
}

func TestService_ListMetricConfigurations(t *testing.T) {
	var (
		tests = []struct {
			name      string
			toeId     string
			setup     func(*service)
			wantCount int
		}{
			{
				name:  "list all for TOE",
				toeId: "toe-1",
				setup: func(svc *service) {
					err := svc.db.Create(&assessment.MetricConfiguration{
						TargetOfEvaluationId: "toe-1",
						MetricId:             "metric-1",
						IsDefault:            true,
					})
					assert.NoError(t, err)

					err = svc.db.Create(&assessment.MetricConfiguration{
						TargetOfEvaluationId: "toe-1",
						MetricId:             "metric-2",
						IsDefault:            false,
					})
					assert.NoError(t, err)

					err = svc.db.Create(&assessment.MetricConfiguration{
						TargetOfEvaluationId: "toe-2",
						MetricId:             "metric-3",
						IsDefault:            true,
					})
					assert.NoError(t, err)
				},
				wantCount: 2,
			},
		}
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				svc = &service{
					db: persistencetest.NewInMemoryDB(t, types, joinTables),
				}
				req = connect.NewRequest(&orchestrator.ListMetricConfigurationRequest{
					TargetOfEvaluationId: tt.toeId,
				})
			)

			tt.setup(svc)

			res, err := svc.ListMetricConfigurations(context.Background(), req)

			assert.NoError(t, err)
			assert.NotNil(t, res)
			assert.Equal(t, tt.wantCount, len(res.Msg.Configurations))
		})
	}
}

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
