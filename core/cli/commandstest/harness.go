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

package commandstest

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"confirmate.io/core/api/orchestrator/orchestratorconnect"
	"confirmate.io/core/cli/commands"
	"confirmate.io/core/persistence"
	"confirmate.io/core/server"
	"confirmate.io/core/server/servertest"
	"confirmate.io/core/service/orchestrator"
	"confirmate.io/core/service/orchestrator/orchestratortest"
	"confirmate.io/core/util/assert"
)

var (
	testServerURL string
	testServer    *httptest.Server
	setupOnce     sync.Once
	setupErr      error
	cleanup       func()
)

// Cleanup stops the shared test server when tests complete.
func Cleanup() {
	if cleanup != nil {
		cleanup()
	}
}

// ensureHarness initializes the shared test harness (once) and returns any setup error.
func ensureHarness(t *testing.T) error {
	t.Helper()

	setupOnce.Do(func() {
		setupErr = func() error {
			var (
				err error
				svc orchestratorconnect.OrchestratorHandler
				srv *server.Server
			)

			svc, err = orchestrator.NewService(orchestrator.WithConfig(orchestrator.Config{
				DefaultMetricsPath:              "../../policies/security-metrics/metrics",
				LoadDefaultMetrics:              false,
				LoadDefaultCatalogs:             false,
				CreateDefaultTargetOfEvaluation: false,
				PersistenceConfig: persistence.Config{
					InMemoryDB: true,
					InitFunc: func(db persistence.DB) error {
						seedCLIData(t, db)
						return nil
					},
				},
			}))
			if err != nil {
				return err
			}

			srv, testServer = servertest.NewTestConnectServer(t,
				server.WithHandler(orchestratorconnect.NewOrchestratorHandler(svc)),
			)
			_ = srv

			testServerURL = testServer.URL
			http.DefaultClient = testServer.Client()
			cleanup = func() {
				testServer.Close()
			}

			return nil
		}()
	})

	return setupErr
}

// RunCLI executes the CLI command and returns combined stdout/stderr output.
func RunCLI(t *testing.T, args ...string) (string, error) {
	t.Helper()

	if err := ensureHarness(t); err != nil {
		return "", err
	}

	cmd := commands.NewRootCommand()
	ctx := context.Background()
	return captureOutput(t, func() error {
		return cmd.Run(ctx, append([]string{"cf", "--addr", testServerURL}, args...))
	})
}

func seedCLIData(t *testing.T, db persistence.DB) {
	assert.NoError(t, db.Create(orchestratortest.MockMetric1))
	assert.NoError(t, db.Create(orchestratortest.MockMetric2))
	assert.NoError(t, db.Create(orchestratortest.MockTargetOfEvaluation1))
	assert.NoError(t, db.Create(orchestratortest.MockTargetOfEvaluation2))
	assert.NoError(t, db.Create(orchestratortest.MockCatalog1))
	assert.NoError(t, db.Create(orchestratortest.MockCatalog2))
	assert.NoError(t, db.Create(orchestratortest.MockCertificate1))
	assert.NoError(t, db.Create(orchestratortest.MockCertificate2))
	assert.NoError(t, db.Create(orchestratortest.MockAssessmentTool1))
	assert.NoError(t, db.Create(orchestratortest.MockAssessmentTool2))
	assert.NoError(t, db.Create(orchestratortest.MockAssessmentResult1))
}

func captureOutput(t *testing.T, fn func() error) (string, error) {
	t.Helper()

	oldStdout := os.Stdout
	oldStderr := os.Stderr

	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		return "", err
	}
	stderrR, stderrW, err := os.Pipe()
	if err != nil {
		_ = stdoutR.Close()
		_ = stdoutW.Close()
		return "", err
	}

	os.Stdout = stdoutW
	os.Stderr = stderrW

	fnErr := fn()

	_ = stdoutW.Close()
	_ = stderrW.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	var stdout bytes.Buffer
	_, _ = io.Copy(&stdout, stdoutR)
	_ = stdoutR.Close()

	var stderr bytes.Buffer
	_, _ = io.Copy(&stderr, stderrR)
	_ = stderrR.Close()

	if stderr.Len() > 0 {
		stdout.WriteString(stderr.String())
	}

	return stdout.String(), fnErr
}
