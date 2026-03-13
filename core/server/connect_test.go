// Copyright 2016-2026 Fraunhofer AISEC
//
// SPDX-License-Identifier: Apache-2.0
//
//                                 /$$$$$$  /$$                                     /$$
//                               /$$__  $$|__/                                    | $$
//  /$$_____/ /$$__  $$| $$__  $$| $$$$    | $$ /$$__  $$| $$_  $$_  $$ |____  $$|_  $$_/   /$$__  $$
// | $$      | $$  \ $$| $$  \ $$| $$_/    | $$| $$  \__/| $$ \ $$ \ $$  /$$$$$$$  | $$    | $$$$$$$$
// | $$      | $$  | $$| $$  | $$| $$      | $$| $$      | $$ | $$ | $$ /$$__  $$  | $$ /$$| $$_____/
// |  $$$$$$$|  $$$$$$/| $$  | $$| $$      | $$| $$      | $$ | $$ | $$|  $$$$$$$  |  $$$$/|  $$$$$$$
// \_______/ \______/ |__/  |__/|__/      |__/|__/      |__/ |__/ |__/ \_______/   \___/   \_______/
//
// This file is part of Confirmate Core.

package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"confirmate.io/core/api/assessment/assessmentconnect"
	"confirmate.io/core/api/evidence/evidenceconnect"
	"confirmate.io/core/api/orchestrator/orchestratorconnect"
	"confirmate.io/core/util/assert"
)

func TestServiceNamesFromHandlerPaths(t *testing.T) {
	var (
		assessmentPath      string
		assessmentHandler   http.Handler
		orchestratorPath    string
		orchestratorHandler http.Handler
		evidencePath        string
		evidenceHandler     http.Handler
	)

	assessmentPath, assessmentHandler = assessmentconnect.NewAssessmentHandler(
		assessmentconnect.UnimplementedAssessmentHandler{},
	)
	orchestratorPath, orchestratorHandler = orchestratorconnect.NewOrchestratorHandler(
		orchestratorconnect.UnimplementedOrchestratorHandler{},
	)
	evidencePath, evidenceHandler = evidenceconnect.NewEvidenceStoreHandler(
		evidenceconnect.UnimplementedEvidenceStoreHandler{},
	)

	tests := []struct {
		name     string
		handlers map[string]http.Handler
		want     []string
	}{
		{
			name: "Extracts service names from generated connect handler paths",
			handlers: map[string]http.Handler{
				orchestratorPath: orchestratorHandler,
				assessmentPath:   assessmentHandler,
				evidencePath:     evidenceHandler,
				"/":              http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}),
			},
			want: []string{
				assessmentconnect.AssessmentName,
				evidenceconnect.EvidenceStoreName,
				orchestratorconnect.OrchestratorName,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := serviceNamesFromHandlerPaths(tt.handlers)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewConnectServer_WithReflection(t *testing.T) {
	tests := []struct {
		name         string
		requestPath  string
		wantHTTPCode int
	}{
		{
			name:         "Reflection-like route is served directly",
			requestPath:  "/grpc.reflection.v1.ServerReflection/ServerReflectionInfo",
			wantHTTPCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, err := NewConnectServer([]Option{
				WithReflection(),
			})
			assert.NoError(t, err)
			assert.NotNil(t, srv)
			if err != nil || srv == nil {
				return
			}

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, tt.requestPath, nil)
			req.Proto = "HTTP/2.0"
			req.ProtoMajor = 2
			req.ProtoMinor = 0
			req.Header.Set("Content-Type", "application/connect+proto")
			req.Header.Set("Connect-Protocol-Version", "1")
			srv.Handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantHTTPCode, rec.Code)
		})
	}
}
