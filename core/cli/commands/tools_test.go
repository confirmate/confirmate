package commands

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/urfave/cli/v3"

	"confirmate.io/core/api/orchestrator/orchestratorconnect"
	"confirmate.io/core/persistence"
	"confirmate.io/core/server"
	"confirmate.io/core/service/orchestrator"
	"confirmate.io/core/util/assert"
)

func TestToolsListCommand(t *testing.T) {
	// We need to start an orchestrator service
	svc, err := orchestrator.NewService(orchestrator.WithConfig(orchestrator.Config{
		DefaultMetricsPath: "../../policies/security-metrics/metrics",
		LoadDefaultMetrics: true,
		PersistenceConfig: persistence.Config{
			InMemoryDB: true,
		},
	}))
	assert.NoError(t, err)

	srv, err := server.NewConnectServer([]server.Option{
		server.WithHandler(orchestratorconnect.NewOrchestratorHandler(svc)),
	})
	assert.NoError(t, err)

	ts := httptest.NewServer(srv.Handler)
	defer ts.Close()

	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "addr",
				Value: ts.URL,
			},
		},
		Commands: []*cli.Command{
			{
				Name: "tools",
				Commands: []*cli.Command{
					ToolsListCommand(),
				},
			},
		},
	}

	ctx := context.Background()
	err = cmd.Run(ctx, []string{"cf", "--addr", ts.URL, "tools", "list"})
	assert.NoError(t, err)
}
