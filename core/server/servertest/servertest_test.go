package servertest_test

import (
	"context"
	"log/slog"
	"testing"

	orchestratorapi "confirmate.io/core/api/orchestrator"
	"confirmate.io/core/api/orchestrator/orchestratorconnect"
	"confirmate.io/core/server"
	"confirmate.io/core/server/servertest"
	"confirmate.io/core/service/orchestrator"
	"confirmate.io/core/util/assert"

	"connectrpc.com/connect"
)

func TestNewTestServer(t *testing.T) {
	// Create an orchestrator service for testing
	svc, err := orchestrator.NewService()
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
