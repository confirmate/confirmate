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

	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/evaluation"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/service/evaluation/evaluationtest"
	"confirmate.io/core/service/orchestrator/orchestratortest"
	"confirmate.io/core/util/assert"
)

// TestService_evaluateControl_leafControl covers controls without sub-controls in the catalog: their own metrics are
// evaluated, and they result in exactly one evaluation result.
func TestService_evaluateControl_leafControl(t *testing.T) {
	var (
		compliantResult = &assessment.AssessmentResult{
			Id:                   orchestratortest.MockResultId1,
			MetricId:             evaluationtest.MockMetricId1,
			ResourceId:           orchestratortest.MockResourceId1,
			Compliant:            true,
			TargetOfEvaluationId: evaluationtest.MockToeId1,
		}
		nonCompliantResult = &assessment.AssessmentResult{
			Id:                   orchestratortest.MockResultId2,
			MetricId:             evaluationtest.MockMetricId1,
			ResourceId:           orchestratortest.MockResourceId2,
			Compliant:            false,
			TargetOfEvaluationId: evaluationtest.MockToeId1,
		}
		leafWithMetrics = &orchestrator.Control{
			Id:      evaluationtest.MockControlId1,
			Metrics: []*assessment.Metric{evaluationtest.MockMetric1},
		}
		leafWithoutMetrics = &orchestrator.Control{
			Id: evaluationtest.MockControlId1,
		}
		// A control whose only sub-control is out of scope must not be evaluated by its own metrics. The sub-control
		// requires a higher assurance level than the one of the audit scope.
		parentWithoutRelevantSubcontrols = &orchestrator.Control{
			Id:      evaluationtest.MockControlId1,
			Metrics: []*assessment.Metric{evaluationtest.MockMetric1},
			Controls: []*orchestrator.Control{{
				Id:              evaluationtest.MockControl1SubcontrolId11,
				ParentControlId: new(evaluationtest.MockControlId1),
				Metrics:         []*assessment.Metric{evaluationtest.MockMetric1},
				AssuranceLevel:  new("high"),
			}},
		}
		basicAuditScope = &orchestrator.AuditScope{
			Id:                   evaluationtest.MockAuditScopeId1,
			TargetOfEvaluationId: evaluationtest.MockToeId1,
			CatalogId:            evaluationtest.MockCatalogId1,
			AssuranceLevel:       new("basic"),
		}
		leveledCatalog = &orchestrator.Catalog{
			Id:              evaluationtest.MockCatalogId1,
			AssuranceLevels: []string{"basic", "high"},
		}
	)

	type args struct {
		// auditScope and catalog default to the mocks if they are not set
		auditScope *orchestrator.AuditScope
		catalog    *orchestrator.Catalog
		control    *orchestrator.Control
		results    []*assessment.AssessmentResult
		// withoutSnapshot evaluates without the snapshot of the latest assessment results, i.e. every control asks the
		// orchestrator for its assessment results itself
		withoutSnapshot bool
	}
	tests := []struct {
		name string
		args args
		want assert.Want[[]*evaluation.EvaluationResult]
	}{
		{
			name: "compliant",
			args: args{control: leafWithMetrics, results: []*assessment.AssessmentResult{compliantResult}},
			want: func(t *testing.T, got []*evaluation.EvaluationResult, msgAndArgs ...any) bool {
				return assert.Equal(t, 1, len(got)) &&
					assert.Equal(t, evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT, got[0].GetStatus()) &&
					assert.Equal(t, []string{compliantResult.Id}, got[0].GetAssessmentResultIds()) &&
					assert.Nil(t, got[0].ParentControlId)
			},
		},
		{
			name: "one non-compliant assessment result makes the control not compliant",
			args: args{control: leafWithMetrics, results: []*assessment.AssessmentResult{compliantResult, nonCompliantResult}},
			want: func(t *testing.T, got []*evaluation.EvaluationResult, msgAndArgs ...any) bool {
				return assert.Equal(t, 1, len(got)) &&
					assert.Equal(t, evaluation.EvaluationStatus_EVALUATION_STATUS_NOT_COMPLIANT, got[0].GetStatus()) &&
					assert.Equal(t, []string{compliantResult.Id, nonCompliantResult.Id}, got[0].GetAssessmentResultIds())
			},
		},
		{
			name: "without the snapshot of the assessment results the orchestrator is asked: not compliant",
			args: args{control: leafWithMetrics, results: []*assessment.AssessmentResult{compliantResult, nonCompliantResult}, withoutSnapshot: true},
			want: func(t *testing.T, got []*evaluation.EvaluationResult, msgAndArgs ...any) bool {
				return assert.Equal(t, 1, len(got)) &&
					assert.Equal(t, evaluation.EvaluationStatus_EVALUATION_STATUS_NOT_COMPLIANT, got[0].GetStatus()) &&
					assert.Equal(t, []string{compliantResult.Id, nonCompliantResult.Id}, got[0].GetAssessmentResultIds())
			},
		},
		{
			name: "no assessment results: pending",
			args: args{control: leafWithMetrics},
			want: func(t *testing.T, got []*evaluation.EvaluationResult, msgAndArgs ...any) bool {
				return assert.Equal(t, 1, len(got)) &&
					assert.Equal(t, evaluation.EvaluationStatus_EVALUATION_STATUS_PENDING, got[0].GetStatus())
			},
		},
		{
			name: "no metrics: pending",
			args: args{control: leafWithoutMetrics, results: []*assessment.AssessmentResult{compliantResult}},
			want: func(t *testing.T, got []*evaluation.EvaluationResult, msgAndArgs ...any) bool {
				return assert.Equal(t, 1, len(got)) &&
					assert.Equal(t, evaluation.EvaluationStatus_EVALUATION_STATUS_PENDING, got[0].GetStatus())
			},
		},
		{
			name: "sub-controls out of scope: the metrics of the control itself are not used",
			args: args{auditScope: basicAuditScope, catalog: leveledCatalog, control: parentWithoutRelevantSubcontrols, results: []*assessment.AssessmentResult{compliantResult}},
			want: func(t *testing.T, got []*evaluation.EvaluationResult, msgAndArgs ...any) bool {
				return assert.Equal(t, 1, len(got)) &&
					assert.Equal(t, evaluation.EvaluationStatus_EVALUATION_STATUS_PENDING, got[0].GetStatus())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, handler := newOrchestratorClientWithHandler(t, WithAssessmentResults(tt.args.results))
			svc := Service{orchestratorClient: client}
			latest := newLatestAssessments(tt.args.results)
			if tt.args.withoutSnapshot {
				latest = nil
			}

			auditScope, catalog := tt.args.auditScope, tt.args.catalog
			if auditScope == nil {
				auditScope = evaluationtest.MockAuditScope1
			}
			if catalog == nil {
				catalog = evaluationtest.MockCatalog1
			}

			err := svc.evaluateControl(context.Background(), auditScope, catalog, tt.args.control, nil, latest)

			assert.NoError(t, err)
			handler.mu.Lock()
			defer handler.mu.Unlock()
			tt.want(t, handler.evaluationResults)
		})
	}
}
