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
	"log/slog"
	"testing"

	"confirmate.io/core/api/orchestrator"
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

func TestAddPaginationAttributes(t *testing.T) {
	tests := []struct {
		name     string
		req      any
		res      any
		wantKeys map[string]bool
	}{
		{
			name: "paginated request with results",
			req: &orchestrator.ListTargetsOfEvaluationRequest{
				PageSize:  10,
				PageToken: "token123",
			},
			res: &orchestrator.ListTargetsOfEvaluationResponse{
				TargetsOfEvaluation: []*orchestrator.TargetOfEvaluation{
					{Id: "toe1"},
					{Id: "toe2"},
					{Id: "toe3"},
				},
				NextPageToken: "nextToken456",
			},
			wantKeys: map[string]bool{
				"page_size":       true,
				"page_token":      true,
				"results":         true,
				"next_page_token": true,
			},
		},
		{
			name: "paginated request without page token",
			req: &orchestrator.ListTargetsOfEvaluationRequest{
				PageSize: 5,
			},
			res: &orchestrator.ListTargetsOfEvaluationResponse{
				TargetsOfEvaluation: []*orchestrator.TargetOfEvaluation{
					{Id: "toe1"},
				},
			},
			wantKeys: map[string]bool{
				"page_size": true,
				"results":   true,
			},
		},
		{
			name: "empty results",
			req: &orchestrator.ListTargetsOfEvaluationRequest{
				PageSize: 10,
			},
			res: &orchestrator.ListTargetsOfEvaluationResponse{
				TargetsOfEvaluation: []*orchestrator.TargetOfEvaluation{},
			},
			wantKeys: map[string]bool{
				"page_size": true,
			},
		},
		{
			name:     "non-paginated request",
			req:      &orchestrator.CreateCatalogRequest{},
			res:      &orchestrator.Catalog{},
			wantKeys: map[string]bool{
				// Should not have any pagination attributes
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var attrs []slog.Attr
			li := &LoggingInterceptor{}

			li.addPaginationAttributes(&attrs, tt.req, tt.res)

			// Build a map of actual attribute keys
			actualKeys := make(map[string]bool)
			for _, attr := range attrs {
				actualKeys[attr.Key] = true
			}

			// Check that all expected keys are present
			for wantKey := range tt.wantKeys {
				if !actualKeys[wantKey] {
					t.Errorf("addPaginationAttributes() missing expected key %q, got keys: %v", wantKey, actualKeys)
				}
			}

			// Check that no unexpected keys are present
			for actualKey := range actualKeys {
				if !tt.wantKeys[actualKey] {
					t.Errorf("addPaginationAttributes() has unexpected key %q", actualKey)
				}
			}

			// Verify specific values for the first test case
			if tt.name == "paginated request with results" {
				for _, attr := range attrs {
					switch attr.Key {
					case "page_size":
						if attr.Value.Int64() != 10 {
							t.Errorf("page_size = %v, want 10", attr.Value.Int64())
						}
					case "page_token":
						if attr.Value.String() != "token123" {
							t.Errorf("page_token = %v, want token123", attr.Value.String())
						}
					case "results":
						if attr.Value.Int64() != 3 {
							t.Errorf("results = %v, want 3", attr.Value.Int64())
						}
					case "next_page_token":
						if attr.Value.String() != "nextToken456" {
							t.Errorf("next_page_token = %v, want nextToken456", attr.Value.String())
						}
					}
				}
			}
		})
	}
}
