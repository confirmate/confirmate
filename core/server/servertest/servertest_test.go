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

package servertest_test

import (
	"context"
	"log/slog"
	"testing"

	orchestratorapi "confirmate.io/core/api/orchestrator"
	"confirmate.io/core/api/orchestrator/orchestratorconnect"
	"confirmate.io/core/persistence"
	"confirmate.io/core/server"
	"confirmate.io/core/server/servertest"
	"confirmate.io/core/service/orchestrator"
	"confirmate.io/core/util/assert"

	"connectrpc.com/connect"
)

func TestNewTestServer(t *testing.T) {
	// Create an orchestrator service for testing
	svc, err := orchestrator.NewService(
		orchestrator.WithConfig(orchestrator.Config{
			PersistenceConfig: persistence.Config{
				InMemoryDB: true,
			},
			CreateDefaultTargetOfEvaluation: false,
			LoadDefaultCatalogs:             false,
			LoadDefaultMetrics:              false,
		}),
	)
	assert.NoError(t, err)
	assert.NotNil(t, svc)

	srv, testSrv := servertest.NewTestConnectServer(t,
		server.WithHandler(orchestratorconnect.NewOrchestratorHandler(svc)),
	)
	defer testSrv.Close()

	assert.NotNil(t, srv)
	assert.NotNil(t, testSrv)

	client := orchestratorconnect.NewOrchestratorClient(
		testSrv.Client(),
		testSrv.URL,
	)

	resp, err := client.ListTargetsOfEvaluation(
		context.Background(),
		connect.NewRequest(&orchestratorapi.ListTargetsOfEvaluationRequest{}),
	)

	assert.NoError(t, err)
	assert.NotNil(t, resp)

	slog.Info("Got response", slog.Any("resp.msg", resp.Msg))

}
