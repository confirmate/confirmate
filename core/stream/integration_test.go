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

package stream

import (
	"context"
	"fmt"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"connectrpc.com/connect"

	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/api/orchestrator/orchestratorconnect"
	"confirmate.io/core/persistence"
	"confirmate.io/core/server"
	"confirmate.io/core/server/servertest"
	orchestratorsvc "confirmate.io/core/service/orchestrator"
	"confirmate.io/core/service/orchestrator/orchestratortest"
	"confirmate.io/core/util/assert"
)

// TestStreamRestartIntegration tests the complete stream restart functionality
// with actual bidirectional streaming over a reconnecting server.
// This test verifies that a stream can be restarted after the server disconnects and
// that it can continue working when the server comes back on the same address.
func TestStreamRestartIntegration(t *testing.T) {
	// Create service
	svc, err := orchestratorsvc.NewService(
		orchestratorsvc.WithConfig(orchestratorsvc.Config{
			PersistenceConfig: persistence.Config{
				InMemoryDB: true,
			},
			CreateDefaultTargetOfEvaluation: false,
			LoadDefaultCatalogs:             false,
			LoadDefaultMetrics:              false,
		}),
	)
	assert.NoError(t, err)

	// Create an initial test server
	_, testSrv1 := servertest.NewTestConnectServer(t,
		server.WithHandler(
			orchestratorconnect.NewOrchestratorHandler(svc),
		),
	)
	serverURL := testSrv1.URL
	// Retrieve port, so we can restart the server later on the same port
	port := testSrv1.Listener.Addr().(*net.TCPAddr).Port

	// Keep http.Client throughout the test (production scenario)
	httpClient := testSrv1.Client()

	// Track restart events
	var restartAttempts atomic.Int32
	var restartSuccesses atomic.Int32

	// Create a client that connects to the same URL
	client := orchestratorconnect.NewOrchestratorClient(httpClient, serverURL)
	factory := func(ctx context.Context) *connect.BidiStreamForClient[orchestrator.StoreAssessmentResultRequest, orchestrator.StoreAssessmentResultsResponse] {
		return client.StoreAssessmentResults(ctx)
	}

	config := DefaultRestartConfig()
	config.MaxRetries = 5
	config.InitialBackoff = 100 * time.Millisecond
	config.MaxBackoff = 500 * time.Millisecond
	config.OnRestart = func(attempt int, err error) {
		restartAttempts.Add(1)
		t.Logf("Restart attempt %d: %v", attempt, err)
	}
	config.OnRestartSuccess = func(attempt int) {
		restartSuccesses.Add(1)
		t.Logf("Restart successful after %d attempts", attempt)
	}

	ctx := context.Background()
	rs, err := NewRestartableBidiStream(ctx, factory, config)
	assert.NoError(t, err)
	defer rs.Close()

	// Send a message on the initial stream
	err = rs.Send(&orchestrator.StoreAssessmentResultRequest{
		Result: orchestratortest.MockAssessmentResult1,
	})

	// Wait until the first result is stored
	_, err = rs.Receive()
	assert.NoError(t, err)

	// Force close client connections before shutting down server
	testSrv1.CloseClientConnections()
	time.Sleep(1000 * time.Millisecond)

	// Close the server to simulate connection loss (but don't close the stream)
	testSrv1.Close()
	t.Log("Server closed - simulating connection loss")
	time.Sleep(200 * time.Millisecond)

	// Restart the server on the same port
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	assert.NoError(t, err)
	_, testSrv2 := servertest.NewTestConnectServerWithListener(t, listener,
		server.WithHandler(
			orchestratorconnect.NewOrchestratorHandler(svc),
		),
	)
	defer testSrv2.Close()
	t.Log("Server restarted on same address")

	// Try to send again - this will trigger restart automatically
	err = rs.Send(&orchestrator.StoreAssessmentResultRequest{
		Result: orchestratortest.MockAssessmentResult2,
	})
	assert.NoError(t, err)

	// Wait until the second result is stored
	_, err = rs.Receive()
	assert.NoError(t, err)

	// Verify restart was attempted and succeeded
	assert.True(t, restartAttempts.Load() > 0, "Expected at least one restart attempt")
	assert.True(t, restartSuccesses.Load() > 0, "Expected at least one successful restart")
	t.Logf("Total restart attempts: %d", restartAttempts.Load())
	t.Logf("Successful restarts: %d", restartSuccesses.Load())

	res, err := svc.ListAssessmentResults(context.Background(), connect.NewRequest(&orchestrator.ListAssessmentResultsRequest{}))
	assert.NoError(t, err)
	assert.Equal(t, 2, len(res.Msg.Results))

	t.Log("Test passed: Stream restart mechanism successfully triggered and executed")
	testSrv2.CloseClientConnections()
}
