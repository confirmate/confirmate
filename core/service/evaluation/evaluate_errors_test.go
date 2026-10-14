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
	"errors"
	"testing"

	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/evaluation"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/service/evaluation/evaluationtest"
	"confirmate.io/core/service/orchestrator/orchestratortest"
	"confirmate.io/core/util/assert"
)

// TestService_evaluateSubcontrol_withoutLatestAssessments covers the fallback of [Service.evaluateSubcontrol], which
// asks the orchestrator itself for the latest assessment results of the control if no snapshot is available.
func TestService_evaluateSubcontrol_withoutLatestAssessments(t *testing.T) {
	tests := []struct {
		name string
		opts []func(*mockOrchestratorHandler)
		want evaluation.EvaluationStatus
	}{
		{
			name: "compliant",
			opts: []func(*mockOrchestratorHandler){func(m *mockOrchestratorHandler) {
				m.assessmentResults = []*assessment.AssessmentResult{{
					Id:                   orchestratortest.MockResultId1,
					MetricId:             orchestratortest.MockMetricId1,
					ResourceId:           orchestratortest.MockResourceId1,
					Compliant:            true,
					TargetOfEvaluationId: evaluationtest.MockToeId1,
				}}
			}},
			want: evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT,
		},
		{
			name: "no assessment results are pending",
			want: evaluation.EvaluationStatus_EVALUATION_STATUS_PENDING,
		},
		{
			name: "an orchestrator error does not fail the evaluation, it stays pending",
			opts: []func(*mockOrchestratorHandler){func(m *mockOrchestratorHandler) {
				m.listAssessmentResultError = errors.New("boom")
			}},
			want: evaluation.EvaluationStatus_EVALUATION_STATUS_PENDING,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{orchestratorClient: newOrchestratorClient(t, tt.opts...)}

			got, err := svc.evaluateSubcontrol(context.Background(), &orchestrator.AuditScope{
				Id:                   evaluationtest.MockAuditScopeId1,
				TargetOfEvaluationId: evaluationtest.MockToeId1,
				CatalogId:            orchestratortest.MockCatalogId1,
			}, orchestratortest.MockSubControl1, nil)

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got.GetStatus())
		})
	}
}

// TestService_evaluateCatalog_errors covers how [Service.evaluateCatalog] reacts to failing calls to the orchestrator.
func TestService_evaluateCatalog_errors(t *testing.T) {
	tests := []struct {
		name    string
		opts    []func(*mockOrchestratorHandler)
		wantErr assert.WantErr
	}{
		{
			name: "controls in scope cannot be listed: all controls are evaluated",
			opts: []func(*mockOrchestratorHandler){func(m *mockOrchestratorHandler) {
				m.listControlsInScopeError = errors.New("boom")
			}},
			wantErr: assert.NoError,
		},
		{
			name: "manual evaluation results cannot be listed: the evaluation fails",
			opts: []func(*mockOrchestratorHandler){func(m *mockOrchestratorHandler) {
				m.listEvalError = errors.New("boom")
			}},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.ErrorContains(t, err, "could not retrieve existing manual evaluation results")
			},
		},
		{
			name: "evaluation results cannot be stored: the evaluation fails",
			opts: []func(*mockOrchestratorHandler){func(m *mockOrchestratorHandler) {
				m.storeEvalError = errors.New("boom")
			}},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.ErrorContains(t, err, "failed to send evaluation result to orchestrator")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := Service{
				orchestratorClient: newOrchestratorClient(t, tt.opts...),
				catalogControls: map[string]map[string]*orchestrator.Control{
					evaluationtest.MockCatalog1.Id: {
						evaluationtest.MockControl1.Id:     evaluationtest.MockControl1,
						evaluationtest.MockSubcontrol11.Id: evaluationtest.MockSubcontrol11,
						evaluationtest.MockSubcontrol12.Id: evaluationtest.MockSubcontrol12,
					},
				},
			}

			err := svc.evaluateCatalog(context.Background(), evaluationtest.MockAuditScope1, evaluationtest.MockCatalog1, 5)

			tt.wantErr(t, err)
		})
	}
}
