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
	"io"
	"testing"
	"time"

	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/api/orchestrator/orchestratorconnect"
	"confirmate.io/core/persistence"
	"confirmate.io/core/persistence/persistencetest"
	"confirmate.io/core/server"
	"confirmate.io/core/server/servertest"
	"confirmate.io/core/service/orchestrator/orchestratortest"
	"confirmate.io/core/util/assert"

	"connectrpc.com/connect"
)

// TestService_StoreAssessmentResults tests the bidirectional streaming RPC.
// This is an integration test focusing on streaming behavior:
// - Stream Send/Receive protocol cycles
// - Multiple messages in one stream
// - Error handling within the stream (status field in response)
// - Stream resilience (continues after errors)
//
// Note: Validation and database error handling for individual results
// are tested in assessment_results_test.go (StoreAssessmentResult unit tests).
func TestService_StoreAssessmentResults(t *testing.T) {
	type fields struct {
		db          persistence.DB
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
			name: "stream - single result",
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
			name: "stream - multiple results in sequence",
			fields: fields{
				db:          persistencetest.NewInMemoryDB(t, types, joinTables),
				subscribers: make(map[int64]*subscriber),
			},
			results: []*assessment.AssessmentResult{
				orchestratortest.MockNewAssessmentResult,
				orchestratortest.MockAssessmentResult2,
			},
			wantStatuses: func(t *testing.T, got []bool, args ...any) bool {
				return assert.Equal(t, []bool{true, true}, got)
			},
			wantErr: assert.NoError,
		},
		{
			name: "stream - resilience with partial failures",
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
					// Pre-create the second result to cause a duplicate error
					assert.NoError(t, d.Create(orchestratortest.MockAssessmentResultForDuplicate))
				}),
				subscribers: make(map[int64]*subscriber),
			},
			results: []*assessment.AssessmentResult{
				orchestratortest.MockNewAssessmentResult,          // Should succeed
				orchestratortest.MockAssessmentResultForDuplicate, // Duplicate - should fail
				orchestratortest.MockAssessmentResult3,            // Should succeed
			},
			wantStatuses: func(t *testing.T, got []bool, args ...any) bool {
				// Stream continues after error - verifies resilience
				return assert.Equal(t, []bool{true, false, true}, got)
			},
			wantErr: assert.NoError, // Stream itself doesn't error, just returns status=false
		},
		{
			name: "stream - empty (no messages)",
			fields: fields{
				db:          persistencetest.NewInMemoryDB(t, types, joinTables),
				subscribers: make(map[int64]*subscriber),
			},
			results: []*assessment.AssessmentResult{}, // No results
			wantStatuses: func(t *testing.T, got []bool, args ...any) bool {
				return len(got) == 0
			},
			wantErr: assert.NoError,
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

// The following tests focus on streaming protocol-specific behavior:
// - Context cancellation
// - Stream lifecycle (close and receive)
// - Concurrent streams
//
// These complement the table-driven test above which covers the basic
// Send/Receive cycles and message handling within streams.

// TestService_StoreAssessmentResults_ContextCancellation tests that stream properly handles context cancellation.
func TestService_StoreAssessmentResults_ContextCancellation(t *testing.T) {
	db := persistencetest.NewInMemoryDB(t, types, joinTables)
	svc := &Service{
		db:          db,
		subscribers: make(map[int64]*subscriber),
	}

	// Create test server
	_, testSrv := servertest.NewTestConnectServer(t,
		server.WithHandler(orchestratorconnect.NewOrchestratorHandler(svc)),
	)
	defer testSrv.Close()

	// Create client
	client := orchestratorconnect.NewOrchestratorClient(testSrv.Client(), testSrv.URL)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())

	// Start stream
	stream := client.StoreAssessmentResults(ctx)

	// Send first result successfully
	err := stream.Send(&orchestrator.StoreAssessmentResultRequest{Result: orchestratortest.MockNewAssessmentResult})
	assert.NoError(t, err)

	res, err := stream.Receive()
	assert.NoError(t, err)
	assert.True(t, res.Status)

	// Cancel context
	cancel()

	// Try to send another result - should fail due to cancelled context
	err = stream.Send(&orchestrator.StoreAssessmentResultRequest{
		Result: &assessment.AssessmentResult{
			Id:                   "00000000-0000-0000-0002-000000000999",
			MetricId:             "metric-2",
			EvidenceId:           orchestratortest.MockEvidenceID2,
			ResourceId:           "resource-2",
			TargetOfEvaluationId: orchestratortest.MockToeID1,
		},
	})

	// Should get an error related to context cancellation
	assert.Error(t, err)
	t.Logf("Got expected error after context cancellation: %v", err)

	// IMPORTANT: Close the stream to prevent server hang
	_ = stream.CloseRequest()
	_ = stream.CloseResponse()
}

// TestService_StoreAssessmentResults_CloseAndReceive tests receiving after stream is closed.
func TestService_StoreAssessmentResults_CloseAndReceive(t *testing.T) {
	db := persistencetest.NewInMemoryDB(t, types, joinTables)
	svc := &Service{
		db:          db,
		subscribers: make(map[int64]*subscriber),
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

	// Send and receive successfully
	err := stream.Send(&orchestrator.StoreAssessmentResultRequest{Result: orchestratortest.MockNewAssessmentResult})
	assert.NoError(t, err)

	res, err := stream.Receive()
	assert.NoError(t, err)
	assert.True(t, res.Status)

	// Close the request stream
	err = stream.CloseRequest()
	assert.NoError(t, err)

	// Try to receive after closing - should get EOF
	_, err = stream.Receive()
	assert.True(t, err == io.EOF || err != nil, "should get EOF or error after closing stream")
	if err == io.EOF {
		t.Log("Got expected EOF after closing stream")
	} else {
		t.Logf("Got error after closing stream: %v", err)
	}
}

// TestService_StoreAssessmentResults_ConcurrentStreams tests multiple concurrent streaming clients.
func TestService_StoreAssessmentResults_ConcurrentStreams(t *testing.T) {
	db := persistencetest.NewInMemoryDB(t, types, joinTables)
	svc := &Service{
		db:          db,
		subscribers: make(map[int64]*subscriber),
	}

	// Create test server
	_, testSrv := servertest.NewTestConnectServer(t,
		server.WithHandler(orchestratorconnect.NewOrchestratorHandler(svc)),
	)
	defer testSrv.Close()

	// Create client
	client := orchestratorconnect.NewOrchestratorClient(testSrv.Client(), testSrv.URL)

	numStreams := 3
	done := make(chan bool, numStreams)

	// Start multiple concurrent streams
	for i := 0; i < numStreams; i++ {
		go func(streamID int) {
			defer func() {
				// Ensure we always signal completion
				if r := recover(); r != nil {
					t.Errorf("Stream %d panicked: %v", streamID, r)
					done <- false
				}
			}()

			stream := client.StoreAssessmentResults(context.Background())
			defer func() {
				_ = stream.CloseRequest()
				_ = stream.CloseResponse()
			}()

			// Each stream sends a unique result using the mock factory function
			result := orchestratortest.NewMockAssessmentResultForConcurrentStream(streamID)

			err := stream.Send(&orchestrator.StoreAssessmentResultRequest{Result: result})
			if err != nil {
				t.Errorf("Stream %d: Send failed: %v", streamID, err)
				done <- false
				return
			}

			res, err := stream.Receive()
			if err != nil {
				t.Errorf("Stream %d: Receive failed: %v", streamID, err)
				done <- false
				return
			}

			if !res.Status {
				t.Errorf("Stream %d: Expected success, got failure: %s", streamID, res.StatusMessage)
				done <- false
				return
			}

			done <- true
		}(i)
	}

	// Wait for all streams to complete
	timeout := time.After(5 * time.Second)
	successCount := 0
	for i := 0; i < numStreams; i++ {
		select {
		case success := <-done:
			if success {
				successCount++
			}
		case <-timeout:
			t.Fatal("Test timed out waiting for concurrent streams")
		}
	}

	assert.Equal(t, numStreams, successCount)
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
			loadCatalogsFunc: nil, // Will not load additional catalogs
			wantCatalogCount: 0,   // No catalogs in empty folder
			wantErr:          assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{
				LoadCatalogsFunc:                tt.loadCatalogsFunc,
				DefaultCatalogsPath:             t.TempDir(), // Empty temp directory
				LoadDefaultCatalogs:             true,
				CreateDefaultTargetOfEvaluation: false,
				LoadDefaultMetrics:              true,
				DefaultMetricsPath:              "./policies/security-metrics/metrics",
				PersistenceConfig: persistence.Config{
					InMemoryDB: true,
				},
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
