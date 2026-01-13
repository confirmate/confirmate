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
	"confirmate.io/core/api/orchestrator/orchestratorconnect"
	"confirmate.io/core/persistence"
	"confirmate.io/core/persistence/persistencetest"
	"confirmate.io/core/server"
	"confirmate.io/core/server/servertest"
	"confirmate.io/core/service/orchestrator/orchestratortest"
	"confirmate.io/core/util"
	"confirmate.io/core/util/assert"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestService_StoreAssessmentResults(t *testing.T) {
	type fields struct {
		db          *persistence.DB
		subscribers map[int64]*subscriber
	}
	tests := []struct {
		name         string
		fields       fields
		results      []*assessment.AssessmentResult
		wantStatuses assert.Want[[]bool]
		wantErr      assert.WantErr
	}{
		{
			name: "happy path - single result",
			fields: fields{
				db:          persistencetest.NewInMemoryDB(t, types, joinTables),
				subscribers: make(map[int64]*subscriber),
			},
			results: []*assessment.AssessmentResult{
				orchestratortest.MockNewAssessmentResult,
			},
			wantStatuses: func(t *testing.T, got []bool, args ...any) bool {
				return assert.Equal(t, []bool{true}, got)
			},
			wantErr: assert.NoError,
		},
		{
			name: "happy path - multiple results",
			fields: fields{
				db:          persistencetest.NewInMemoryDB(t, types, joinTables),
				subscribers: make(map[int64]*subscriber),
			},
			results: []*assessment.AssessmentResult{
				orchestratortest.MockNewAssessmentResult,
				{
					Id:                   "00000000-0000-0000-0002-000000000004",
					CreatedAt:            timestamppb.Now(),
					MetricId:             "metric-2",
					MetricConfiguration:  orchestratortest.MockMetricConfiguration2,
					Compliant:            false,
					EvidenceId:           orchestratortest.MockEvidenceID2,
					ResourceId:           "resource-2",
					ResourceTypes:        []string{"vm"},
					ComplianceComment:    "Second resource test",
					TargetOfEvaluationId: orchestratortest.MockToeID1,
					ToolId:               util.Ref("tool-1"),
					HistoryUpdatedAt:     timestamppb.Now(),
					History: []*assessment.Record{
						{
							EvidenceId:         orchestratortest.MockEvidenceID2,
							EvidenceRecordedAt: timestamppb.Now(),
						},
					},
				},
			},
			wantStatuses: func(t *testing.T, got []bool, args ...any) bool {
				return assert.Equal(t, []bool{true, true}, got)
			},
			wantErr: assert.NoError,
		},
		{
			name: "database error - duplicate id",
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d *persistence.DB) {
					assert.NoError(t, d.Create(orchestratortest.MockNewAssessmentResult))
				}),
				subscribers: make(map[int64]*subscriber),
			},
			results: []*assessment.AssessmentResult{
				orchestratortest.MockNewAssessmentResult,
			},
			wantStatuses: func(t *testing.T, got []bool, args ...any) bool {
				// Stream should be resilient - send error response (status=false) but continue
				return assert.Equal(t, []bool{false}, got)
			},
			wantErr: assert.NoError, // Stream itself should not error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				err      error
				statuses []bool
			)

			svc := &Service{
				db:          tt.fields.db,
				subscribers: tt.fields.subscribers,
			}

			// Create test server
			_, testSrv := servertest.NewTestConnectServer(t,
				server.WithHandler(orchestratorconnect.NewOrchestratorHandler(svc)),
			)
			defer testSrv.Close()

			// Create client
			client := orchestratorconnect.NewOrchestratorClient(testSrv.Client(), testSrv.URL)

			// Start stream
			stream := client.StoreAssessmentResults(context.Background())

			// Send all results and collect responses
			for i, result := range tt.results {
				sendErr := stream.Send(&orchestrator.StoreAssessmentResultRequest{Result: result})
				if sendErr != nil {
					err = sendErr
					break
				}

				res, recvErr := stream.Receive()
				if recvErr != nil {
					err = recvErr
					break
				}
				statuses = append(statuses, res.Status)
				if !res.Status {
					t.Logf("Result %d failed with message: %s", i, res.StatusMessage)
				}
			}

			// Close stream
			_ = stream.CloseRequest()

			tt.wantStatuses(t, statuses)
			tt.wantErr(t, err)
		})
	}
}

func TestService_LoadCatalogsFunc(t *testing.T) {
	tests := []struct {
		name             string
		loadCatalogsFunc func(*Service) ([]*orchestrator.Catalog, error)
		wantCatalogCount int
		wantErr          assert.WantErr
	}{
		{
			name: "custom load function",
			loadCatalogsFunc: func(svc *Service) ([]*orchestrator.Catalog, error) {
				return []*orchestrator.Catalog{
					orchestratortest.MockCatalog1,
					orchestratortest.MockCatalog2,
				}, nil
			},
			wantCatalogCount: 2,
			wantErr:          assert.NoError,
		},
		{
			name:             "default embedded catalogs (empty folder)",
			loadCatalogsFunc: nil, // Will use default loadEmbeddedCatalogs
			wantCatalogCount: 0,   // No catalogs in empty folder
			wantErr:          assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg Config
			if tt.loadCatalogsFunc != nil {
				cfg = Config{
					LoadCatalogsFunc:                tt.loadCatalogsFunc,
					CreateDefaultTargetOfEvaluation: false,
					IgnoreDefaultMetrics:            false, // Use default metrics to avoid error
					DefaultMetricsPath:              "./policies/security-metrics/metrics",
				}
			} else {
				cfg = Config{
					CatalogsFolder:                  t.TempDir(), // Empty temp directory
					CreateDefaultTargetOfEvaluation: false,
					IgnoreDefaultMetrics:            false, // Use default metrics to avoid error
					DefaultMetricsPath:              "./policies/security-metrics/metrics",
				}
			}

			handler, err := NewService(WithConfig(cfg))
			tt.wantErr(t, err)
			if err != nil {
				return
			}

			svc := handler.(*Service)

			// List catalogs to verify they were loaded
			res, err := svc.ListCatalogs(context.Background(), connect.NewRequest(&orchestrator.ListCatalogsRequest{}))
			assert.NoError(t, err)
			assert.NotNil(t, res.Msg)
			assert.Equal(t, tt.wantCatalogCount, len(res.Msg.Catalogs))

			// If we loaded custom catalogs, verify their IDs
			if tt.loadCatalogsFunc != nil {
				assert.Equal(t, orchestratortest.MockCatalog1.Id, res.Msg.Catalogs[0].Id)
				assert.Equal(t, orchestratortest.MockCatalog2.Id, res.Msg.Catalogs[1].Id)
			}
		})
	}
}
