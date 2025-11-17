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

package stream_test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/api/orchestrator/orchestratorconnect"
	"confirmate.io/core/server"
	"confirmate.io/core/server/servertest"
	orchestratorsvc "confirmate.io/core/service/orchestrator"
	"confirmate.io/core/stream"

	"connectrpc.com/connect"
)

// ExampleRestartableBidiStream demonstrates how to use RestartableBidiStream
// to maintain a continuous connection that automatically restarts on errors.
func ExampleRestartableBidiStream() {
	// Create a test server (in production, this would be your actual server)
	svc, err := orchestratorsvc.NewService()
	if err != nil {
		log.Fatalf("Failed to create service: %v", err)
	}

	_, testSrv := servertest.NewTestConnectServer(nil,
			server.WithHandler(
				orchestratorconnect.NewOrchestratorHandler(svc),
			),
	)
	defer testSrv.Close()

	// Create a Connect client
	client := orchestratorconnect.NewOrchestratorClient(
		http.DefaultClient,
		testSrv.URL,
	)

	// Define a factory function that creates a new stream
	factory := func(ctx context.Context) *connect.BidiStreamForClient[orchestrator.StoreAssessmentResultRequest, orchestrator.StoreAssessmentResultsResponse] {
		return client.StoreAssessmentResults(ctx)
	}

	// Configure auto-restart behavior
	config := stream.DefaultRestartConfig()
	config.MaxRetries = 5 // Try up to 5 times
	config.InitialBackoff = 100 * time.Millisecond
	config.MaxBackoff = 10 * time.Second

	// Create the restartable stream
	ctx := context.Background()
	rs, err := stream.NewRestartableBidiStream(ctx, factory, config)
	if err != nil {
		log.Fatalf("Failed to create restartable stream: %v", err)
	}
	defer rs.Close()

	// Use the stream - it will automatically restart on errors
	// Send a request (this is just an example, actual usage may vary)
	// err = rs.Send(&orchestrator.StoreAssessmentResultRequest{...})
	// if err != nil {
	//     log.Printf("Error: %v", err)
	// }

	fmt.Println("Restartable stream created successfully")
	// Output: Restartable stream created successfully
}

// ExampleRestartableBidiStream_withCallbacks demonstrates using custom callbacks
// to monitor stream restart events.
func ExampleRestartableBidiStream_withCallbacks() {
	// Create a test server
	svc, err := orchestratorsvc.NewService()
	if err != nil {
		log.Fatalf("Failed to create service: %v", err)
	}

	_, testSrv := servertest.NewTestConnectServer(nil,
			server.WithHandler(
				orchestratorconnect.NewOrchestratorHandler(svc),
			),
	)
	defer testSrv.Close()

	client := orchestratorconnect.NewOrchestratorClient(
		http.DefaultClient,
		testSrv.URL,
	)

	factory := func(ctx context.Context) *connect.BidiStreamForClient[orchestrator.StoreAssessmentResultRequest, orchestrator.StoreAssessmentResultsResponse] {
		return client.StoreAssessmentResults(ctx)
	}

	// Configure with custom callbacks
	config := stream.DefaultRestartConfig()
	config.OnRestart = func(attempt int, err error) {
		fmt.Printf("Restarting stream (attempt %d) due to error: %v\n", attempt, err)
	}
	config.OnRestartSuccess = func(attempt int) {
		fmt.Printf("Stream restart successful after %d attempts\n", attempt)
	}
	config.OnRestartFailure = func(err error) {
		fmt.Printf("Failed to restart stream: %v\n", err)
	}

	ctx := context.Background()
	rs, err := stream.NewRestartableBidiStream(ctx, factory, config)
	if err != nil {
		log.Fatalf("Failed to create restartable stream: %v", err)
	}
	defer rs.Close()

	fmt.Println("Stream with custom callbacks created")
	// Output: Stream with custom callbacks created
}

// ExampleRestartConfig demonstrates different restart configurations.
func ExampleRestartConfig() {
	// Default configuration - unlimited retries with exponential backoff
	defaultConfig := stream.DefaultRestartConfig()
	fmt.Printf("Default: MaxRetries=%d, InitialBackoff=%v, MaxBackoff=%v\n",
		defaultConfig.MaxRetries,
		defaultConfig.InitialBackoff,
		defaultConfig.MaxBackoff)

	// Custom configuration - limited retries with faster backoff
	customConfig := stream.RestartConfig{
		MaxRetries:        3,
		InitialBackoff:    50 * time.Millisecond,
		MaxBackoff:        5 * time.Second,
		BackoffMultiplier: 2.0,
		OnRestart: func(attempt int, err error) {
			fmt.Printf("Restart attempt %d\n", attempt)
		},
		OnRestartSuccess: func(attempt int) {
			fmt.Printf("Success after %d attempts\n", attempt)
		},
		OnRestartFailure: func(err error) {
			fmt.Printf("All retries failed: %v\n", err)
		},
	}
	fmt.Printf("Custom: MaxRetries=%d, InitialBackoff=%v\n",
		customConfig.MaxRetries,
		customConfig.InitialBackoff)

	// Output:
	// Default: MaxRetries=0, InitialBackoff=100ms, MaxBackoff=30s
	// Custom: MaxRetries=3, InitialBackoff=50ms
}

// ExampleRestartableBidiStream_monitoring demonstrates how to monitor
// stream health and retry statistics.
func ExampleRestartableBidiStream_monitoring() {
	// Create a test server
	svc, err := orchestratorsvc.NewService()
	if err != nil {
		log.Fatalf("Failed to create service: %v", err)
	}

	_, testSrv := servertest.NewTestConnectServer(nil,
			server.WithHandler(
				orchestratorconnect.NewOrchestratorHandler(svc),
			),
	)
	defer testSrv.Close()

	client := orchestratorconnect.NewOrchestratorClient(
		http.DefaultClient,
		testSrv.URL,
	)

	factory := func(ctx context.Context) *connect.BidiStreamForClient[orchestrator.StoreAssessmentResultRequest, orchestrator.StoreAssessmentResultsResponse] {
		return client.StoreAssessmentResults(ctx)
	}

	config := stream.DefaultRestartConfig()
	ctx := context.Background()
	rs, err := stream.NewRestartableBidiStream(ctx, factory, config)
	if err != nil {
		log.Fatalf("Failed to create restartable stream: %v", err)
	}
	defer rs.Close()

	// Check stream health
	retryCount := rs.RetryCount()
	lastError := rs.LastError()

	fmt.Printf("Retry count: %d\n", retryCount)
	if lastError != nil {
		fmt.Printf("Last error: %v\n", lastError)
	} else {
		fmt.Println("No errors yet")
	}

	// Output:
	// Retry count: 0
	// No errors yet
}
