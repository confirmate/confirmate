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
	"confirmate.io/core/api/orchestrator/orchestratorconnect"
	"confirmate.io/core/persistence"
	"confirmate.io/core/persistence/persistencetest"
	"confirmate.io/core/server"
	"confirmate.io/core/server/servertest"
	"confirmate.io/core/util/assert"

	"connectrpc.com/connect"
)

func Test_service_ListTargetsOfEvaluation(t *testing.T) {
	tests := []struct {
		name   string
		fields struct {
			db *persistence.DB
		}
		args struct {
			ctx context.Context
			req *connect.Request[orchestrator.ListTargetsOfEvaluationRequest]
		}
		want    *connect.Response[orchestrator.ListTargetsOfEvaluationResponse]
		wantErr bool
	}{
		{
			name: "happy path",
			fields: struct {
				db *persistence.DB
			}{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(s *persistence.DB) {
					// Create a sample TargetOfEvaluation entry
					err := s.Create(&orchestrator.TargetOfEvaluation{
						Id:   "1",
						Name: "TOE1",
					})
					if err != nil {
						t.Fatalf("could not create TOE: %v", err)
					}
				}),
			},
			args: struct {
				ctx context.Context
				req *connect.Request[orchestrator.ListTargetsOfEvaluationRequest]
			}{
				ctx: context.Background(),
				req: connect.NewRequest(&orchestrator.ListTargetsOfEvaluationRequest{}),
			},
			want: connect.NewResponse(&orchestrator.ListTargetsOfEvaluationResponse{
				TargetsOfEvaluation: []*orchestrator.TargetOfEvaluation{
					{
						Id:   "1",
						Name: "TOE1",
					},
				},
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &service{
				db: tt.fields.db,
			}
			got, gotErr := svc.ListTargetsOfEvaluation(context.Background(), nil)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ListTargetsOfEvaluation() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ListTargetsOfEvaluation() succeeded unexpectedly")
			}
			assert.Equal(t, tt.want.Msg, got.Msg)
		})
	}
}

// TestStoreAssessmentResults_Basic tests basic streaming functionality
func TestStoreAssessmentResults_Basic(t *testing.T) {
	// Create an orchestrator service for testing
	svc, err := NewService()
	assert.NoError(t, err)
	assert.NotNil(t, svc)

	// Start the orchestrator server
	srv, testSrv := servertest.NewTestConnectServer(t,
		server.WithHandler(orchestratorconnect.NewOrchestratorHandler(svc)),
	)
	defer testSrv.Close()
	assert.NotNil(t, srv)
	assert.NotNil(t, testSrv)

	// Create a client
	client := orchestratorconnect.NewOrchestratorClient(
		testSrv.Client(),
		testSrv.URL,
	)

	// Start the stream
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	stream := client.StoreAssessmentResults(ctx)

	// Send a message
	t.Log("Sending message...")
	err = stream.Send(&orchestrator.StoreAssessmentResultRequest{
		Result: &assessment.AssessmentResult{
			Id:       "test-result-1",
			MetricId: "test-metric",
		},
	})
	assert.NoError(t, err)

	// Receive response
	t.Log("Receiving response...")
	resp, err := stream.Receive()
	assert.NoError(t, err)
	assert.True(t, resp.Status)
	assert.Equal(t, "received", resp.StatusMessage)

	// Close the stream properly
	err = stream.CloseRequest()
	assert.NoError(t, err)

	t.Log("✓ Basic streaming test passed")
}

// TestStoreAssessmentResults_StreamResilience tests the streaming capabilities including
// server stop/restart scenarios.
func TestStoreAssessmentResults_StreamResilience(t *testing.T) {
	// Create an orchestrator service for testing
	svc, err := NewService()
	assert.NoError(t, err)
	assert.NotNil(t, svc)

	// Start the orchestrator server
	srv, testSrv := servertest.NewTestConnectServer(t,
		server.WithHandler(orchestratorconnect.NewOrchestratorHandler(svc)),
	)
	assert.NotNil(t, srv)
	assert.NotNil(t, testSrv)

	// Create a client
	client := orchestratorconnect.NewOrchestratorClient(
		testSrv.Client(),
		testSrv.URL,
	)

	// Start the stream
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	stream := client.StoreAssessmentResults(ctx)

	// Send a message
	t.Log("Sending first message...")
	err = stream.Send(&orchestrator.StoreAssessmentResultRequest{
		Result: &assessment.AssessmentResult{
			Id:       "test-result-1",
			MetricId: "test-metric",
		},
	})
	assert.NoError(t, err)

	// Receive response
	t.Log("Receiving first response...")
	resp, err := stream.Receive()
	assert.NoError(t, err)
	assert.True(t, resp.Status)
	assert.Equal(t, "test-result-1 received", resp.StatusMessage)
	t.Log("First message sent and received successfully")

	// Close the orchestrator server
	t.Log("Closing orchestrator server...")
	testSrv.CloseClientConnections()

	// Give it a moment to fully close
	time.Sleep(100 * time.Millisecond)

	// Try to send another message - this should fail since server is closed
	t.Log("Attempting to send message with closed server...")
	err = stream.Send(&orchestrator.StoreAssessmentResultRequest{
		Result: &assessment.AssessmentResult{
			Id:       "test-result-2",
			MetricId: "test-metric",
		},
	})

	// We expect an error here since the server is closed
	assert.True(t, err != nil, "Expected error when sending to closed server")
	t.Logf("Got expected error after server close: %v", err)
}
