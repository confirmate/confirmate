// Copyright 2016-2025 Fraunhofer AISEC
//
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"testing"

	"confirmate.io/core/api/orchestrator"
)

func TestGetResultsCount(t *testing.T) {
	tests := []struct {
		name     string
		response *orchestrator.ListTargetsOfEvaluationResponse
		want     int
	}{
		{
			name: "nil response",
			response: &orchestrator.ListTargetsOfEvaluationResponse{
				TargetsOfEvaluation: nil,
			},
			want: 0,
		},
		{
			name: "empty response",
			response: &orchestrator.ListTargetsOfEvaluationResponse{
				TargetsOfEvaluation: []*orchestrator.TargetOfEvaluation{},
			},
			want: 0,
		},
		{
			name: "single result",
			response: &orchestrator.ListTargetsOfEvaluationResponse{
				TargetsOfEvaluation: []*orchestrator.TargetOfEvaluation{
					{Id: "test1"},
				},
			},
			want: 1,
		},
		{
			name: "multiple results",
			response: &orchestrator.ListTargetsOfEvaluationResponse{
				TargetsOfEvaluation: []*orchestrator.TargetOfEvaluation{
					{Id: "test1"},
					{Id: "test2"},
					{Id: "test3"},
				},
			},
			want: 3,
		},
		{
			name:     "nil response",
			response: nil,
			want:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetResultsCount(tt.response)
			if got != tt.want {
				t.Errorf("GetResultsCount() = %v, want %v", got, tt.want)
			}
		})
	}
}
