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

package server

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	"confirmate.io/core/api"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/util/assert"
)

func TestOperationType(t *testing.T) {
	tests := []struct {
		name   string
		method string
		want   orchestrator.RequestType
	}{
		{"CreateCatalog", "CreateCatalog", orchestrator.RequestType_REQUEST_TYPE_CREATED},
		{"UpdateCatalog", "UpdateCatalog", orchestrator.RequestType_REQUEST_TYPE_UPDATED},
		{"RemoveCatalog", "RemoveCatalog", orchestrator.RequestType_REQUEST_TYPE_DELETED},
		{"StoreAssessmentResult", "StoreAssessmentResult", orchestrator.RequestType_REQUEST_TYPE_STORED},
		{"RegisterAssessmentTool", "RegisterAssessmentTool", orchestrator.RequestType_REQUEST_TYPE_REGISTERED},
		{"GetCatalog", "GetCatalog", orchestrator.RequestType_REQUEST_TYPE_UNSPECIFIED},
		{"ListCatalogs", "ListCatalogs", orchestrator.RequestType_REQUEST_TYPE_UNSPECIFIED},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := operationType(tt.method); got != tt.want {
				t.Errorf("operationType(%q) = %v, want %v", tt.method, got, tt.want)
			}
		})
	}
}

func TestMethodName(t *testing.T) {
	tests := []struct {
		name      string
		procedure string
		want      string
	}{
		{"Full path", "/confirmate.orchestrator.v1.OrchestratorService/CreateCatalog", "CreateCatalog"},
		{"Simple", "/Service/Method", "Method"},
		{"No slash", "MethodName", "MethodName"},
		{"Empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := methodName(tt.procedure); got != tt.want {
				t.Errorf("methodName(%q) = %v, want %v", tt.procedure, got, tt.want)
			}
		})
	}
}

func TestIsReadOperation(t *testing.T) {
	tests := []struct {
		name   string
		method string
		want   bool
	}{
		{"Get", "GetCatalog", true},
		{"List", "ListCatalogs", true},
		{"Query", "QueryMetrics", true},
		{"Search", "SearchTools", true},
		{"Create", "CreateCatalog", false},
		{"Update", "UpdateCatalog", false},
		{"Delete", "DeleteCatalog", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isReadOperation(tt.method); got != tt.want {
				t.Errorf("isReadOperation(%q) = %v, want %v", tt.method, got, tt.want)
			}
		})
	}
}

func TestRequestTypeToVerb(t *testing.T) {
	tests := []struct {
		name        string
		requestType orchestrator.RequestType
		want        string
	}{
		{
			name:        "Created",
			requestType: orchestrator.RequestType_REQUEST_TYPE_CREATED,
			want:        "created",
		},
		{
			name:        "Updated",
			requestType: orchestrator.RequestType_REQUEST_TYPE_UPDATED,
			want:        "updated",
		},
		{
			name:        "Deleted",
			requestType: orchestrator.RequestType_REQUEST_TYPE_DELETED,
			want:        "deleted",
		},
		{
			name:        "Registered",
			requestType: orchestrator.RequestType_REQUEST_TYPE_REGISTERED,
			want:        "registered",
		},
		{
			name:        "Stored",
			requestType: orchestrator.RequestType_REQUEST_TYPE_STORED,
			want:        "stored",
		},
		{
			name:        "Unknown",
			requestType: orchestrator.RequestType_REQUEST_TYPE_UNSPECIFIED,
			want:        "changed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := requestTypeToVerb(tt.requestType)
			if got != tt.want {
				t.Errorf("requestTypeToVerb() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLogRequest(t *testing.T) {
	tests := []struct {
		name            string
		level           slog.Level
		requestType     orchestrator.RequestType
		req             any
		attrs           []slog.Attr
		wantContains    []string
		wantNotContains []string
	}{
		{
			name:        "nil request",
			level:       slog.LevelDebug,
			requestType: orchestrator.RequestType_REQUEST_TYPE_CREATED,
			req:         nil,
			// Should not log anything
			wantNotContains: []string{"Catalog", "created"},
		},
		{
			name:        "Create catalog with ID",
			level:       slog.LevelInfo,
			requestType: orchestrator.RequestType_REQUEST_TYPE_CREATED,
			req: &orchestrator.CreateCatalogRequest{
				Catalog: &orchestrator.Catalog{
					Id:   "catalog-123",
					Name: "Test Catalog",
				},
			},
			wantContains: []string{
				"CreateCatalogRequest",
				"created",
			},
		},
		{
			name:        "Update target of evaluation",
			level:       slog.LevelDebug,
			requestType: orchestrator.RequestType_REQUEST_TYPE_UPDATED,
			req: &orchestrator.UpdateTargetOfEvaluationRequest{
				TargetOfEvaluation: &orchestrator.TargetOfEvaluation{
					Id:   "toe-456",
					Name: "Test TOE",
				},
			},
			wantContains: []string{
				"UpdateTargetOfEvaluationRequest",
				"updated",
			},
		},
		{
			name:        "Create with additional attributes",
			level:       slog.LevelInfo,
			requestType: orchestrator.RequestType_REQUEST_TYPE_CREATED,
			req: &orchestrator.CreateCatalogRequest{
				Catalog: &orchestrator.Catalog{
					Id:   "catalog-abc",
					Name: "Test Catalog",
				},
			},
			attrs: []slog.Attr{
				slog.String("extra", "info"),
				slog.Int("count", 42),
			},
			wantContains: []string{
				"CreateCatalogRequest",
				"created",
				"extra=info",
				"count=42",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				buf    bytes.Buffer
				logger *slog.Logger
			)

			// Create a logger that writes to our buffer
			logger = slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			}))

			// Set as default logger for the test
			slog.SetDefault(logger)

			// Call logRequest with the request (which is now PayloadRequest directly)
			if tt.req != nil {
				if payloadReq, ok := tt.req.(api.PayloadRequest); ok {
					li := &LoggingInterceptor{}
					li.logRequest(context.Background(), tt.level, tt.requestType, payloadReq, tt.attrs...)
				}
			}

			output := buf.String()

			// Check that expected strings are present
			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("logRequest() output does not contain %q\nGot: %s", want, output)
				}
			}

			// Check that unexpected strings are not present
			for _, notWant := range tt.wantNotContains {
				if strings.Contains(output, notWant) {
					t.Errorf("logRequest() output should not contain %q\nGot: %s", notWant, output)
				}
			}
		})
	}
}

func TestLogRequest_Integration(t *testing.T) {
	var (
		buf    bytes.Buffer
		logger *slog.Logger
		req    *orchestrator.CreateCatalogRequest
	)

	// Create a logger with JSON handler for easier parsing in production
	logger = slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	req = &orchestrator.CreateCatalogRequest{
		Catalog: &orchestrator.Catalog{
			Id:          "catalog-integration-test",
			Name:        "Integration Test Catalog",
			Description: "This is a test catalog",
		},
	}

	li := &LoggingInterceptor{}
	li.logRequest(context.Background(), slog.LevelInfo, orchestrator.RequestType_REQUEST_TYPE_CREATED, req,
		slog.String("user", "admin"),
		slog.String("source", "api"),
	)

	output := buf.String()

	// Verify the output contains expected JSON fields
	assert.True(t, strings.Contains(output, `"level":"INFO"`))
	assert.True(t, strings.Contains(output, `"msg":"CreateCatalogRequest created"`))
	assert.True(t, strings.Contains(output, `"user":"admin"`))
	assert.True(t, strings.Contains(output, `"source":"api"`))
}
