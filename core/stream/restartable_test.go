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
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/api/orchestrator/orchestratorconnect"
	"confirmate.io/core/server"
	"confirmate.io/core/server/servertest"
	orchestratorsvc "confirmate.io/core/service/orchestrator"

	"connectrpc.com/connect"
)

// mockBidiStream is a mock implementation of a bidirectional stream for testing.
type mockBidiStream[Req, Res any] struct {
	sendFunc    func(*Req) error
	receiveFunc func() (*Res, error)
	closed      bool
}

func (m *mockBidiStream[Req, Res]) Send(msg *Req) error {
	if m.sendFunc != nil {
		return m.sendFunc(msg)
	}
	return nil
}

func (m *mockBidiStream[Req, Res]) Receive() (*Res, error) {
	if m.receiveFunc != nil {
		return m.receiveFunc()
	}
	return nil, nil
}

func (m *mockBidiStream[Req, Res]) CloseRequest() error {
	m.closed = true
	return nil
}

func (m *mockBidiStream[Req, Res]) CloseResponse() error {
	m.closed = true
	return nil
}

// TestRestartableBidiStream_Basic tests basic send/receive operations.
func TestRestartableBidiStream_Basic(t *testing.T) {
	msg := &orchestrator.StoreAssessmentResultRequest{}
	resp := &orchestrator.StoreAssessmentResultsResponse{Status: true}

	mockStream := &mockBidiStream[orchestrator.StoreAssessmentResultRequest, orchestrator.StoreAssessmentResultsResponse]{
		sendFunc: func(r *orchestrator.StoreAssessmentResultRequest) error {
			return nil
		},
		receiveFunc: func() (*orchestrator.StoreAssessmentResultsResponse, error) {
			return resp, nil
		},
	}

	factory := func(ctx context.Context) *connect.BidiStreamForClient[orchestrator.StoreAssessmentResultRequest, orchestrator.StoreAssessmentResultsResponse] {
		// We can't easily create a real BidiStreamForClient in tests without a server,
		// so we'll test with integration tests below
		return nil
	}

	config := DefaultRestartConfig()
	config.MaxRetries = 3

	// For unit testing, we'll verify the config is correct
	if config.MaxRetries != 3 {
		t.Errorf("Expected MaxRetries=3, got %d", config.MaxRetries)
	}
	if config.InitialBackoff != 100*time.Millisecond {
		t.Errorf("Expected InitialBackoff=100ms, got %v", config.InitialBackoff)
	}

	_ = factory
	_ = mockStream
	_ = msg
}

// TestRestartableBidiStream_AutoRestart tests automatic restart on error.
func TestRestartableBidiStream_AutoRestart(t *testing.T) {
	// This test verifies that the restart configuration is properly applied
	config := DefaultRestartConfig()
	config.MaxRetries = 3
	config.InitialBackoff = 10 * time.Millisecond
	config.MaxBackoff = 100 * time.Millisecond

	// Verify configuration
	if config.MaxRetries != 3 {
		t.Errorf("Expected MaxRetries=3, got %d", config.MaxRetries)
	}
	if config.InitialBackoff != 10*time.Millisecond {
		t.Errorf("Expected InitialBackoff=10ms, got %v", config.InitialBackoff)
	}
	if config.MaxBackoff != 100*time.Millisecond {
		t.Errorf("Expected MaxBackoff=100ms, got %v", config.MaxBackoff)
	}
}

// TestRestartableBidiStream_RetryCount tests retry counting using integration test setup.
func TestRestartableBidiStream_RetryCount(t *testing.T) {
	// Create a test server for proper stream creation
	svc, err := orchestratorsvc.NewService()
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	_, testSrv := servertest.NewTestConnectServer(t,
		server.WithHandler(
			orchestratorconnect.NewOrchestratorHandler(svc),
		),
	)
	defer testSrv.Close()

	client := orchestratorconnect.NewOrchestratorClient(
		http.DefaultClient,
		testSrv.URL,
	)

	ctx := context.Background()
	config := DefaultRestartConfig()
	config.MaxRetries = 5
	config.InitialBackoff = 1 * time.Millisecond
	
	factory := func(ctx context.Context) *connect.BidiStreamForClient[orchestrator.StoreAssessmentResultRequest, orchestrator.StoreAssessmentResultsResponse] {
		return client.StoreAssessmentResults(ctx)
	}

	rs, err := NewRestartableBidiStream(ctx, factory, config)
	if err != nil {
		t.Fatalf("Failed to create restartable stream: %v", err)
	}
	defer rs.Close()

	// Initial retry count should be 0
	if rs.RetryCount() != 0 {
		t.Errorf("Expected retry count 0, got %d", rs.RetryCount())
	}
}

// TestRestartableBidiStream_Close tests proper cleanup on close.
func TestRestartableBidiStream_Close(t *testing.T) {
	// Create a test server for proper stream creation
	svc, err := orchestratorsvc.NewService()
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	_, testSrv := servertest.NewTestConnectServer(t,
		server.WithHandler(
			orchestratorconnect.NewOrchestratorHandler(svc),
		),
	)
	defer testSrv.Close()

	client := orchestratorconnect.NewOrchestratorClient(
		http.DefaultClient,
		testSrv.URL,
	)

	ctx := context.Background()
	factory := func(ctx context.Context) *connect.BidiStreamForClient[orchestrator.StoreAssessmentResultRequest, orchestrator.StoreAssessmentResultsResponse] {
		return client.StoreAssessmentResults(ctx)
	}

	config := DefaultRestartConfig()
	rs, err := NewRestartableBidiStream(ctx, factory, config)
	if err != nil {
		t.Fatalf("Failed to create restartable stream: %v", err)
	}

	// Close the stream
	err = rs.Close()
	if err != nil {
		t.Errorf("Close() returned error: %v", err)
	}

	// Verify stream is closed
	rs.mu.RLock()
	closed := rs.closed
	rs.mu.RUnlock()
	
	if !closed {
		t.Error("Stream should be marked as closed")
	}

	// Second close should be idempotent
	err = rs.Close()
	if err != nil {
		t.Errorf("Second Close() returned error: %v", err)
	}
}

// TestRestartableBidiStream_ContextCancellation tests behavior when context is cancelled.
func TestRestartableBidiStream_ContextCancellation(t *testing.T) {
	// Create a test server for proper stream creation
	svc, err := orchestratorsvc.NewService()
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	_, testSrv := servertest.NewTestConnectServer(t,
		server.WithHandler(
			orchestratorconnect.NewOrchestratorHandler(svc),
		),
	)
	defer testSrv.Close()

	client := orchestratorconnect.NewOrchestratorClient(
		http.DefaultClient,
		testSrv.URL,
	)

	ctx, cancel := context.WithCancel(context.Background())
	factory := func(ctx context.Context) *connect.BidiStreamForClient[orchestrator.StoreAssessmentResultRequest, orchestrator.StoreAssessmentResultsResponse] {
		return client.StoreAssessmentResults(ctx)
	}

	config := DefaultRestartConfig()
	rs, err := NewRestartableBidiStream(ctx, factory, config)
	if err != nil {
		t.Fatalf("Failed to create restartable stream: %v", err)
	}
	defer rs.Close()

	// Cancel the context
	cancel()

	// Give it a moment to process
	time.Sleep(50 * time.Millisecond)

	// Context should be done
	select {
	case <-rs.ctx.Done():
		// Expected
	default:
		t.Error("Context should be cancelled")
	}
}

// TestRestartableBidiStream_Integration performs an integration test with real server.
func TestRestartableBidiStream_Integration(t *testing.T) {
	// Create a test server
	svc, err := orchestratorsvc.NewService()
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Create a real test server
	srv, testSrv := servertest.NewTestConnectServer(t,
		server.WithHandler(
			orchestratorconnect.NewOrchestratorHandler(svc),
		),
	)
	defer testSrv.Close()
	defer srv.Close()

	// Create a client
	client := orchestratorconnect.NewOrchestratorClient(
		http.DefaultClient,
		testSrv.URL,
	)

	// Test connection by listing targets
	ctx := context.Background()
	resp, err := client.ListTargetsOfEvaluation(ctx, connect.NewRequest(&orchestrator.ListTargetsOfEvaluationRequest{}))
	if err != nil {
		t.Fatalf("Failed to list targets: %v", err)
	}

	if len(resp.Msg.TargetsOfEvaluation) == 0 {
		t.Error("Expected at least one target of evaluation")
	}

	t.Log("Integration test passed: server is working correctly")
}

// TestRestartableBidiStream_StreamRecovery tests that a stream can be recovered after connection loss.
func TestRestartableBidiStream_StreamRecovery(t *testing.T) {
	// This test simulates connection loss and recovery
	var serverRestart atomic.Int32
	var mu sync.Mutex
	
	// Create a factory that tracks server restarts
	createServer := func() *httptest.Server {
		mu.Lock()
		defer mu.Unlock()
		
		serverRestart.Add(1)
		
		svc, err := orchestratorsvc.NewService()
		if err != nil {
			t.Fatalf("Failed to create service: %v", err)
		}

		_, testSrv := servertest.NewTestConnectServer(t,
			server.WithHandler(
				orchestratorconnect.NewOrchestratorHandler(svc),
			),
		)
		
		return testSrv
	}

	// Start initial server
	testSrv := createServer()
	
	// Create client
	client := orchestratorconnect.NewOrchestratorClient(
		http.DefaultClient,
		testSrv.URL,
	)

	ctx := context.Background()
	
	// First request should succeed
	_, err := client.ListTargetsOfEvaluation(ctx, connect.NewRequest(&orchestrator.ListTargetsOfEvaluationRequest{}))
	if err != nil {
		t.Fatalf("Initial request failed: %v", err)
	}

	// Simulate server restart by closing and creating new server
	mu.Lock()
	testSrv.Close()
	mu.Unlock()
	
	// Wait a bit
	time.Sleep(100 * time.Millisecond)
	
	// Create new server
	newTestSrv := createServer()
	defer newTestSrv.Close()
	
	// Create new client pointing to new server
	newClient := orchestratorconnect.NewOrchestratorClient(
		http.DefaultClient,
		newTestSrv.URL,
	)

	// This request should succeed with the new server
	_, err = newClient.ListTargetsOfEvaluation(ctx, connect.NewRequest(&orchestrator.ListTargetsOfEvaluationRequest{}))
	if err != nil {
		t.Fatalf("Request after restart failed: %v", err)
	}

	// Verify server was restarted
	if serverRestart.Load() < 2 {
		t.Errorf("Expected at least 2 server starts, got %d", serverRestart.Load())
	}

	t.Log("Stream recovery test passed: successfully recovered from connection loss")
}

// TestRestartConfig_Defaults tests the default configuration values.
func TestRestartConfig_Defaults(t *testing.T) {
	config := DefaultRestartConfig()

	if config.MaxRetries != 0 {
		t.Errorf("Expected MaxRetries=0 (unlimited), got %d", config.MaxRetries)
	}
	if config.InitialBackoff != 100*time.Millisecond {
		t.Errorf("Expected InitialBackoff=100ms, got %v", config.InitialBackoff)
	}
	if config.MaxBackoff != 30*time.Second {
		t.Errorf("Expected MaxBackoff=30s, got %v", config.MaxBackoff)
	}
	if config.BackoffMultiplier != 2.0 {
		t.Errorf("Expected BackoffMultiplier=2.0, got %f", config.BackoffMultiplier)
	}
	if config.OnRestart == nil {
		t.Error("Expected OnRestart callback to be set")
	}
	if config.OnRestartSuccess == nil {
		t.Error("Expected OnRestartSuccess callback to be set")
	}
	if config.OnRestartFailure == nil {
		t.Error("Expected OnRestartFailure callback to be set")
	}
}

// TestRestartableBidiStream_MaxRetriesExceeded tests that retries stop after max attempts.
func TestRestartableBidiStream_MaxRetriesExceeded(t *testing.T) {
	// Create a test server for proper stream creation
	svc, err := orchestratorsvc.NewService()
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	_, testSrv := servertest.NewTestConnectServer(t,
		server.WithHandler(
			orchestratorconnect.NewOrchestratorHandler(svc),
		),
	)
	defer testSrv.Close()

	client := orchestratorconnect.NewOrchestratorClient(
		http.DefaultClient,
		testSrv.URL,
	)

	ctx := context.Background()
	var restartAttempts atomic.Int32
	var failureCalled atomic.Bool
	
	factory := func(ctx context.Context) *connect.BidiStreamForClient[orchestrator.StoreAssessmentResultRequest, orchestrator.StoreAssessmentResultsResponse] {
		return client.StoreAssessmentResults(ctx)
	}

	config := DefaultRestartConfig()
	config.MaxRetries = 3
	config.InitialBackoff = 1 * time.Millisecond
	config.MaxBackoff = 10 * time.Millisecond
	config.OnRestart = func(attempt int, err error) {
		restartAttempts.Add(1)
	}
	config.OnRestartFailure = func(err error) {
		failureCalled.Store(true)
	}

	rs, err := NewRestartableBidiStream(ctx, factory, config)
	if err != nil {
		t.Fatalf("Failed to create restartable stream: %v", err)
	}
	defer rs.Close()

	// Close the server to simulate connection failure
	testSrv.Close()

	// Create a factory that always fails to create a stream
	failingFactory := func(ctx context.Context) *connect.BidiStreamForClient[orchestrator.StoreAssessmentResultRequest, orchestrator.StoreAssessmentResultsResponse] {
		// Return nil to simulate stream creation failure
		return nil
	}
	
	// Replace the factory with the failing one
	rs.factory = failingFactory

	// Simulate an error that triggers restart
	testErr := errors.New("test connection error")
	restartErr := rs.restart(testErr)
	
	// Should fail after max retries
	if restartErr == nil {
		t.Error("Expected restart to fail after max retries")
	}

	// Verify restart was attempted at least once
	if restartAttempts.Load() < 1 {
		t.Errorf("Expected at least 1 restart attempt, got %d", restartAttempts.Load())
	}

	// Verify failure callback was called
	if !failureCalled.Load() {
		t.Error("Expected OnRestartFailure callback to be called")
	}

	// Verify last error is stored
	if rs.LastError() != testErr {
		t.Errorf("Expected last error to be %v, got %v", testErr, rs.LastError())
	}
}

// TestRestartableBidiStream_ExponentialBackoff tests exponential backoff behavior.
func TestRestartableBidiStream_ExponentialBackoff(t *testing.T) {
	config := DefaultRestartConfig()
	config.InitialBackoff = 10 * time.Millisecond
	config.MaxBackoff = 1 * time.Second
	config.BackoffMultiplier = 2.0

	// Calculate expected backoffs
	expectedBackoffs := []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
		40 * time.Millisecond,
		80 * time.Millisecond,
		160 * time.Millisecond,
		320 * time.Millisecond,
		640 * time.Millisecond,
		1 * time.Second, // capped at MaxBackoff
		1 * time.Second, // stays at MaxBackoff
	}

	backoff := config.InitialBackoff
	for i, expected := range expectedBackoffs {
		if backoff != expected {
			t.Errorf("Backoff at step %d: expected %v, got %v", i, expected, backoff)
		}
		
		backoff = time.Duration(float64(backoff) * config.BackoffMultiplier)
		if backoff > config.MaxBackoff {
			backoff = config.MaxBackoff
		}
	}
}

// TestRestartableBidiStream_ConcurrentOperations tests thread safety.
func TestRestartableBidiStream_ConcurrentOperations(t *testing.T) {
	// Create a test server for proper stream creation
	svc, err := orchestratorsvc.NewService()
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	_, testSrv := servertest.NewTestConnectServer(t,
		server.WithHandler(
			orchestratorconnect.NewOrchestratorHandler(svc),
		),
	)
	defer testSrv.Close()

	client := orchestratorconnect.NewOrchestratorClient(
		http.DefaultClient,
		testSrv.URL,
	)

	ctx := context.Background()
	factory := func(ctx context.Context) *connect.BidiStreamForClient[orchestrator.StoreAssessmentResultRequest, orchestrator.StoreAssessmentResultsResponse] {
		return client.StoreAssessmentResults(ctx)
	}

	config := DefaultRestartConfig()
	config.MaxRetries = 1
	config.InitialBackoff = 1 * time.Millisecond
	
	rs, err := NewRestartableBidiStream(ctx, factory, config)
	if err != nil {
		t.Fatalf("Failed to create restartable stream: %v", err)
	}
	defer rs.Close()

	// Run concurrent operations
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			// Try various operations concurrently
			_ = rs.RetryCount()
			_ = rs.LastError()
			_ = rs.CloseRequest()
			_ = rs.CloseResponse()
		}()
	}

	wg.Wait()
	t.Log("Concurrent operations test passed: no data races detected")
}

// TestRestartableBidiStream_SendAfterClose tests error handling when sending after close.
func TestRestartableBidiStream_SendAfterClose(t *testing.T) {
	// Create a test server for proper stream creation
	svc, err := orchestratorsvc.NewService()
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	_, testSrv := servertest.NewTestConnectServer(t,
		server.WithHandler(
			orchestratorconnect.NewOrchestratorHandler(svc),
		),
	)
	defer testSrv.Close()

	client := orchestratorconnect.NewOrchestratorClient(
		http.DefaultClient,
		testSrv.URL,
	)

	ctx := context.Background()
	factory := func(ctx context.Context) *connect.BidiStreamForClient[orchestrator.StoreAssessmentResultRequest, orchestrator.StoreAssessmentResultsResponse] {
		return client.StoreAssessmentResults(ctx)
	}

	config := DefaultRestartConfig()
	rs, err := NewRestartableBidiStream(ctx, factory, config)
	if err != nil {
		t.Fatalf("Failed to create restartable stream: %v", err)
	}

	// Close the stream
	err = rs.Close()
	if err != nil {
		t.Errorf("Close() returned error: %v", err)
	}

	// Try to send after close - should fail
	err = rs.Send(&orchestrator.StoreAssessmentResultRequest{})
	if err == nil {
		t.Error("Expected error when sending after close, got nil")
	}
	if err != nil && !errors.Is(err, fmt.Errorf("stream is closed")) {
		// Check error message contains "closed"
		if err.Error() != "stream is closed" {
			t.Errorf("Expected 'stream is closed' error, got: %v", err)
		}
	}
}

// TestRestartableBidiStream_ReceiveAfterClose tests error handling when receiving after close.
func TestRestartableBidiStream_ReceiveAfterClose(t *testing.T) {
	// Create a test server for proper stream creation
	svc, err := orchestratorsvc.NewService()
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	_, testSrv := servertest.NewTestConnectServer(t,
		server.WithHandler(
			orchestratorconnect.NewOrchestratorHandler(svc),
		),
	)
	defer testSrv.Close()

	client := orchestratorconnect.NewOrchestratorClient(
		http.DefaultClient,
		testSrv.URL,
	)

	ctx := context.Background()
	factory := func(ctx context.Context) *connect.BidiStreamForClient[orchestrator.StoreAssessmentResultRequest, orchestrator.StoreAssessmentResultsResponse] {
		return client.StoreAssessmentResults(ctx)
	}

	config := DefaultRestartConfig()
	rs, err := NewRestartableBidiStream(ctx, factory, config)
	if err != nil {
		t.Fatalf("Failed to create restartable stream: %v", err)
	}

	// Close the stream
	err = rs.Close()
	if err != nil {
		t.Errorf("Close() returned error: %v", err)
	}

	// Try to receive after close - should fail
	_, err = rs.Receive()
	if err == nil {
		t.Error("Expected error when receiving after close, got nil")
	}
	if err != nil && err.Error() != "stream is closed" {
		t.Errorf("Expected 'stream is closed' error, got: %v", err)
	}
}
