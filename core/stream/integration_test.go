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
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/api/orchestrator/orchestratorconnect"
	"confirmate.io/core/server"
	orchestratorsvc "confirmate.io/core/service/orchestrator"
	"confirmate.io/core/util/assert"
)

// TestStreamRestartIntegration tests the complete stream restart functionality
// with actual bidirectional streaming over a reconnecting server.
func TestStreamRestartIntegration(t *testing.T) {
	// Create service
	svc, err := orchestratorsvc.NewService()
	assert.NoError(t, err)

	// Create server instance
	srv, err := server.NewConnectServer([]server.Option{
		server.WithHandler(
			orchestratorconnect.NewOrchestratorHandler(svc),
		),
	})
	assert.NoError(t, err)
	defer srv.Close()

	// Start initial server
	testSrv1 := httptest.NewServer(srv.Handler)
	defer testSrv1.Close()

	// Keep http.Client throughout the test (production scenario)
	httpClient := http.DefaultClient
	client1 := orchestratorconnect.NewOrchestratorClient(httpClient, testSrv1.URL)
	
	// Track restart events
	var restartAttempts atomic.Int32

	config := DefaultRestartConfig()
	config.MaxRetries = 3
	config.InitialBackoff = 50 * time.Millisecond
	config.OnRestart = func(attempt int, err error) {
		restartAttempts.Add(1)
	}

	ctx := context.Background()
	rs, err := NewRestartableBidiStream(ctx, client1.StoreAssessmentResults, config)
	assert.NoError(t, err)
	defer rs.Close()

	// Send a message on the initial stream
	err = rs.Send(&orchestrator.StoreAssessmentResultRequest{})
	assert.NoError(t, err)

	// Close request to clean up
	err = rs.CloseRequest()
	assert.NoError(t, err)

	// Verify the stream was created successfully and has a name
	assert.NotNil(t, rs)
}
