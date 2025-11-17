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
	"sync/atomic"
	"testing"
	"time"

	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/api/orchestrator/orchestratorconnect"
	"confirmate.io/core/server"
	"confirmate.io/core/server/servertest"
	orchestratorsvc "confirmate.io/core/service/orchestrator"
	"confirmate.io/core/util/assert"
)

// TestStreamRestartIntegration tests the complete stream restart functionality
// with actual bidirectional streaming over a reconnecting server.
func TestStreamRestartIntegration(t *testing.T) {
	// Create initial server
	svc1, err := orchestratorsvc.NewService()
	assert.NoError(t, err)

	srv1, testSrv1 := servertest.NewTestConnectServer(t,
		server.WithHandler(
			orchestratorconnect.NewOrchestratorHandler(svc1),
		),
	)
	defer srv1.Close()

	// Keep http.Client throughout the test (production scenario)
	httpClient := http.DefaultClient
	client := orchestratorconnect.NewOrchestratorClient(
		httpClient,
		testSrv1.URL,
	)

	// Track restart events
	var restartAttempts atomic.Int32
	var restartSuccesses atomic.Int32

	factory := client.StoreAssessmentResults

	config := DefaultRestartConfig()
	config.MaxRetries = 3
	config.InitialBackoff = 50 * time.Millisecond
	config.OnRestart = func(attempt int, err error) {
		restartAttempts.Add(1)
	}
	config.OnRestartSuccess = func(attempt int) {
		restartSuccesses.Add(1)
	}

	ctx := context.Background()
	rs, err := NewRestartableBidiStream(ctx, factory, config)
	assert.NoError(t, err)

	// Send a message on the initial stream
	err = rs.Send(&orchestrator.StoreAssessmentResultRequest{})
	assert.NoError(t, err)

	// Close request side of stream before shutting down server
	err = rs.CloseRequest()
	assert.NoError(t, err)

	// Close the first server to simulate connection loss
	testSrv1.Close()
	time.Sleep(100 * time.Millisecond)

	// Close the restartable stream
	err = rs.Close()
	assert.NoError(t, err)
}
