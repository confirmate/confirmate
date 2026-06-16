// Copyright 2016-2026 Fraunhofer AISEC
//
// SPDX-License-Identifier: Apache-2.0
//
// This file is part of Confirmate Core.

package commands

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"confirmate.io/core/api"
	"confirmate.io/core/api/evidence"
	"confirmate.io/core/api/evidence/evidenceconnect"
	"confirmate.io/core/api/ontology"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/api/orchestrator/orchestratorconnect"
	"confirmate.io/core/util/assert"

	"connectrpc.com/connect"
	"golang.org/x/oauth2/clientcredentials"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TestConfirmate_AuthEnabled_EvidenceProducesAssessmentResults is a regression
// test for issue #310. With --auth-enabled, the full evidence -> assessment ->
// orchestrator pipeline must still produce assessment results: every internal
// hop has to authenticate as the service client, and the orchestrator's
// authorizer has to recognize that client as admin.
func TestConfirmate_AuthEnabled_EvidenceProducesAssessmentResults(t *testing.T) {
	// The assessment service's Rego loader resolves policy bundles relative to
	// the current working directory (./policies/security-metrics/...). Chdir
	// into the core/ root so the bundled metrics submodule is discoverable.
	_, thisFile, _, _ := runtime.Caller(0)
	t.Chdir(filepath.Join(filepath.Dir(thisFile), "..", ".."))

	port := pickFreePort(t)
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	keyPath := filepath.Join(t.TempDir(), "api.key")

	ctx, cancel := context.WithCancel(context.Background())
	serverDone := make(chan error, 1)
	go func() {
		serverDone <- ConfirmateCommand.Run(ctx, []string{
			"confirmate",
			"--auth-enabled",
			"--db-in-memory",
			"--create-default-target-of-evaluation",
			"--api-port", fmt.Sprintf("%d", port),
			"--oauth2-key-path", keyPath,
			"--oauth2-key-password", "test",
			"--log-level", "ERROR",
			// All service-to-service hops default to localhost:8080 — point them
			// at the test's random port instead.
			"--auth-jwks-url", baseURL + "/v1/auth/certs",
			"--service-oauth2-token-endpoint", baseURL + "/v1/auth/token",
			"--assessment-orchestrator-address", baseURL,
			"--evidence-assessment-address", baseURL,
		})
	}()
	t.Cleanup(func() {
		cancel()
		select {
		case <-serverDone:
		case <-time.After(10 * time.Second):
			t.Log("confirmate command did not shut down within 10s")
		}
	})

	if !waitFor(t, 15*time.Second, 100*time.Millisecond, func() bool {
		resp, err := http.Get(baseURL + "/.well-known/openid-configuration")
		if err != nil {
			return false
		}
		_ = resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}) {
		t.Fatal("confirmate server did not become ready")
	}

	// Authenticate as the service client (same flow used by the assessment and
	// evidence services to talk to the orchestrator).
	authClient := api.NewOAuthHTTPClient(http.DefaultClient,
		api.NewOAuthAuthorizerFromClientCredentials(&clientcredentials.Config{
			ClientID:     "confirmate",
			ClientSecret: "confirmate",
			TokenURL:     baseURL + "/v1/auth/token",
		}))

	_, err := evidenceconnect.NewEvidenceStoreClient(authClient, baseURL).
		StoreEvidence(ctx, connect.NewRequest(&evidence.StoreEvidenceRequest{
			Evidence: loadBalancerEvidence(),
		}))
	assert.NoError(t, err)

	orchClient := orchestratorconnect.NewOrchestratorClient(authClient, baseURL)
	var lastCount int
	if !waitFor(t, 20*time.Second, 200*time.Millisecond, func() bool {
		res, err := orchClient.ListAssessmentResults(ctx,
			connect.NewRequest(&orchestrator.ListAssessmentResultsRequest{}))
		if err != nil {
			return false
		}
		lastCount = len(res.Msg.GetResults())
		return lastCount > 0
	}) {
		t.Fatalf("expected at least one assessment result, got %d", lastCount)
	}
}

// pickFreePort reserves a random local port and immediately releases it. There
// is a small race window before the caller binds to it, but it is negligible
// for in-process tests.
func pickFreePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("could not allocate port: %v", err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	_ = l.Close()
	return port
}

func waitFor(t *testing.T, timeout, interval time.Duration, condition func() bool) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for {
		if condition() {
			return true
		}
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(interval)
	}
}

// loadBalancerEvidence builds a minimal LoadBalancer evidence — the same shape
// that exercises the bundled default metrics so the assessment service produces
// results. The target_of_evaluation_id matches the one created by
// --create-default-target-of-evaluation.
func loadBalancerEvidence() *evidence.Evidence {
	return &evidence.Evidence{
		Id:                   "50000000-0000-0000-0000-000000000000",
		Timestamp:            timestamppb.New(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)),
		TargetOfEvaluationId: "00000000-0000-0000-0000-000000000000",
		ToolId:               "manual",
		Resource: &ontology.Resource{
			Type: &ontology.Resource_LoadBalancer{
				LoadBalancer: &ontology.LoadBalancer{
					Id:           "123e4567-e89b-12d3-a456-426614174000",
					Name:         "Example Load Balancer",
					Description:  "Example Load Balancer",
					CreationTime: timestamppb.New(time.Date(2023, 5, 11, 14, 16, 9, 0, time.UTC)),
					ParentId:     new("123e4567-e89b-12d3-a456-426614174002"),
					GeoLocation:  &ontology.GeoLocation{Region: "Germany"},
					AccessRestriction: &ontology.AccessRestriction{
						Type: &ontology.AccessRestriction_WebApplicationFirewall{
							WebApplicationFirewall: &ontology.WebApplicationFirewall{Enabled: true},
						},
					},
					TransportEncryption: &ontology.TransportEncryption{
						Enabled:         true,
						Enforced:        true,
						Protocol:        "HTTPS",
						ProtocolVersion: 1.2,
					},
				},
			},
		},
	}
}

