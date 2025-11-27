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

func TestService_StoreAssessmentResult(t *testing.T) {
	var (
		tests = []struct {
			name    string
			req     *orchestrator.StoreAssessmentResultRequest
			wantErr bool
		}{
			{
				name: "happy path",
				req: &orchestrator.StoreAssessmentResultRequest{
					Result: &assessment.AssessmentResult{
						MetricId:      "metric-1",
						EvidenceId:    "evidence-1",
						ResourceId:    "resource-1",
						ResourceTypes: []string{"vm"},
						Compliant:     true,
					},
				},
				wantErr: false,
			},
			{
				name: "with existing ID",
				req: &orchestrator.StoreAssessmentResultRequest{
					Result: &assessment.AssessmentResult{
						Id:            "existing-id",
						MetricId:      "metric-2",
						EvidenceId:    "evidence-2",
						ResourceId:    "resource-2",
						ResourceTypes: []string{"storage"},
						Compliant:     false,
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

			res, err := svc.StoreAssessmentResult(context.Background(), req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, res)
			assert.NotEmpty(t, res.Msg)
		})
	}
}

func TestService_GetAssessmentResult(t *testing.T) {
	var (
		tests = []struct {
			name    string
			id      string
			setup   func(*service)
			wantErr bool
		}{
			{
				name: "happy path",
				id:   "result-1",
				setup: func(svc *service) {
					err := svc.db.Create(&assessment.AssessmentResult{
						Id:         "result-1",
						MetricId:   "metric-1",
						Compliant:  true,
						ResourceId: "resource-1",
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
				req = connect.NewRequest(&orchestrator.GetAssessmentResultRequest{
					Id: tt.id,
				})
			)

			tt.setup(svc)

			res, err := svc.GetAssessmentResult(context.Background(), req)

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

func TestService_ListAssessmentResults(t *testing.T) {
	var (
		tests = []struct {
			name      string
			filter    *orchestrator.ListAssessmentResultsRequest_Filter
			setup     func(*service)
			wantCount int
		}{
			{
				name:   "list all",
				filter: nil,
				setup: func(svc *service) {
					err := svc.db.Create(&assessment.AssessmentResult{
						Id:         "result-1",
						MetricId:   "metric-1",
						Compliant:  true,
						ResourceId: "resource-1",
					})
					assert.NoError(t, err)

					err = svc.db.Create(&assessment.AssessmentResult{
						Id:         "result-2",
						MetricId:   "metric-2",
						Compliant:  false,
						ResourceId: "resource-2",
					})
					assert.NoError(t, err)
				},
				wantCount: 2,
			},
			{
				name: "filter by metric ID",
				filter: &orchestrator.ListAssessmentResultsRequest_Filter{
					MetricId: stringPtr("metric-1"),
				},
				setup: func(svc *service) {
					err := svc.db.Create(&assessment.AssessmentResult{
						Id:         "result-1",
						MetricId:   "metric-1",
						Compliant:  true,
						ResourceId: "resource-1",
					})
					assert.NoError(t, err)

					err = svc.db.Create(&assessment.AssessmentResult{
						Id:         "result-2",
						MetricId:   "metric-2",
						Compliant:  false,
						ResourceId: "resource-2",
					})
					assert.NoError(t, err)
				},
				wantCount: 1,
			},
			{
				name: "filter by compliant",
				filter: &orchestrator.ListAssessmentResultsRequest_Filter{
					Compliant: boolPtr(true),
				},
				setup: func(svc *service) {
					err := svc.db.Create(&assessment.AssessmentResult{
						Id:         "result-1",
						MetricId:   "metric-1",
						Compliant:  true,
						ResourceId: "resource-1",
					})
					assert.NoError(t, err)

					err = svc.db.Create(&assessment.AssessmentResult{
						Id:         "result-2",
						MetricId:   "metric-2",
						Compliant:  false,
						ResourceId: "resource-2",
					})
					assert.NoError(t, err)
				},
				wantCount: 1,
			},
		}
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				svc = &service{
					db: persistencetest.NewInMemoryDB(t, types, joinTables),
				}
				req = connect.NewRequest(&orchestrator.ListAssessmentResultsRequest{
					Filter: tt.filter,
				})
			)

			tt.setup(svc)

			res, err := svc.ListAssessmentResults(context.Background(), req)

			assert.NoError(t, err)
			assert.NotNil(t, res)
			assert.Equal(t, tt.wantCount, len(res.Msg.Results))
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
