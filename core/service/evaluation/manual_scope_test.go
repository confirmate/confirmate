// Copyright 2016-2026 Fraunhofer AISEC
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

package evaluation

import (
	"context"
	"testing"

	"confirmate.io/core/api/evaluation"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/service/evaluation/evaluationtest"
	"confirmate.io/core/util/assert"
)

// TestService_evaluateCatalog_manualResultsAreScopedToTheAuditScope checks that a manual evaluation result of a control only
// applies within the audit scope it was created in: several audit scopes can share the same target of evaluation and
// catalog, and a manual result of one must not suppress the automated evaluation of the other.
func TestService_evaluateCatalog_manualResultsAreScopedToTheAuditScope(t *testing.T) {
	tests := []struct {
		name    string
		results []*evaluation.EvaluationResult
		want    assert.Want[[]*evaluation.EvaluationResult]
	}{
		{
			name:    "manual result of the same audit scope: the control is not evaluated automatically",
			results: []*evaluation.EvaluationResult{evaluationtest.MockManualEvaluationResult1}, // audit scope 1
			want: func(t *testing.T, got []*evaluation.EvaluationResult, msgAndArgs ...any) bool {
				return assert.Equal(t, 0, len(got))
			},
		},
		{
			name:    "manual result of another audit scope of the same target of evaluation: the control is evaluated",
			results: []*evaluation.EvaluationResult{evaluationtest.MockManualEvaluationResult3}, // audit scope 2
			want: func(t *testing.T, got []*evaluation.EvaluationResult, msgAndArgs ...any) bool {
				return assert.Equal(t, 1, len(got)) &&
					assert.Equal(t, evaluationtest.MockAuditScopeId1, got[0].GetAuditScopeId())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, handler := newOrchestratorClientWithHandler(t, WithEvaluationResults(tt.results))
			svc := Service{
				orchestratorClient: client,
				catalogControls: map[string]map[string]*orchestrator.Control{
					evaluationtest.MockCatalog1.Id: {
						evaluationtest.MockControl1.Id:     evaluationtest.MockControl1,
						evaluationtest.MockSubcontrol11.Id: evaluationtest.MockSubcontrol11,
						evaluationtest.MockSubcontrol12.Id: evaluationtest.MockSubcontrol12,
					},
				},
			}

			err := svc.evaluateCatalog(context.Background(), evaluationtest.MockAuditScope1, evaluationtest.MockCatalog1, 5)

			assert.NoError(t, err)

			// Collect the results for the top-level control that were created by this evaluation
			handler.mu.Lock()
			defer handler.mu.Unlock()
			var created []*evaluation.EvaluationResult
			for _, r := range handler.evaluationResults {
				if r.GetControlId() == evaluationtest.MockControl1.Id && r.GetStatus() != evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT_MANUALLY {
					created = append(created, r)
				}
			}
			tt.want(t, created)
		})
	}
}
