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
	"confirmate.io/core/api/orchestrator/orchestratorconnect"
	"confirmate.io/core/service"
	"confirmate.io/core/service/evaluation/evaluationtest"
	"confirmate.io/core/service/orchestrator/orchestratortest"
	"confirmate.io/core/util/assert"
)

// notCompliantResult1 is a non-compliant assessment result for the metric of [orchestratortest.MockSubControl1].
var notCompliantResult1 = &assessment.AssessmentResult{
	Id:         orchestratortest.MockResultId3,
	MetricId:   orchestratortest.MockMetricId1,
	ResourceId: orchestratortest.MockResourceId1,
	Compliant:  false,
}

func Test_latestAssessments_forMetrics(t *testing.T) {
	type args struct {
		results   []*assessment.AssessmentResult
		metricIds []string
	}
	tests := []struct {
		name string
		args args
		want assert.Want[[]*assessment.AssessmentResult]
	}{
		{
			name: "no metrics",
			args: args{
				results: []*assessment.AssessmentResult{orchestratortest.MockAssessmentResult1},
			},
			want: assert.Empty[[]*assessment.AssessmentResult],
		},
		{
			name: "unknown metric",
			args: args{
				results:   []*assessment.AssessmentResult{orchestratortest.MockAssessmentResult1},
				metricIds: []string{orchestratortest.MockNonExistentId},
			},
			want: assert.Empty[[]*assessment.AssessmentResult],
		},
		{
			name: "only the requested metrics are returned",
			args: args{
				results: []*assessment.AssessmentResult{
					orchestratortest.MockAssessmentResult1,
					orchestratortest.MockAssessmentResult2,
				},
				metricIds: []string{orchestratortest.MockMetricId2},
			},
			want: func(t *testing.T, got []*assessment.AssessmentResult, args ...any) bool {
				return assert.Equal(t, 1, len(got)) &&
					assert.Equal(t, orchestratortest.MockResultId2, got[0].GetId())
			},
		},
		{
			name: "duplicate metric ids do not duplicate results",
			args: args{
				results:   []*assessment.AssessmentResult{orchestratortest.MockAssessmentResult1},
				metricIds: []string{orchestratortest.MockMetricId1, orchestratortest.MockMetricId1},
			},
			want: func(t *testing.T, got []*assessment.AssessmentResult, args ...any) bool {
				return assert.Equal(t, 1, len(got))
			},
		},
		{
			name: "results are ordered by resource id and metric id",
			args: args{
				results: []*assessment.AssessmentResult{
					orchestratortest.MockAssessmentResult2, // resource 2, metric 2
					orchestratortest.MockAssessmentResult1, // resource 1, metric 1
				},
				metricIds: []string{orchestratortest.MockMetricId2, orchestratortest.MockMetricId1},
			},
			want: func(t *testing.T, got []*assessment.AssessmentResult, args ...any) bool {
				return assert.Equal(t, 2, len(got)) &&
					assert.Equal(t, orchestratortest.MockResultId1, got[0].GetId()) &&
					assert.Equal(t, orchestratortest.MockResultId2, got[1].GetId())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := newLatestAssessments(tt.args.results)

			got := l.forMetrics(tt.args.metricIds)

			tt.want(t, got)
		})
	}
}

// TestService_evaluateSubcontrol_withLatestAssessments checks that the snapshot of the latest assessment results is
// used instead of asking the orchestrator: listing assessment results fails in the mock, so the evaluation can only
// succeed using the snapshot.
func TestService_evaluateSubcontrol_withLatestAssessments(t *testing.T) {
	type args struct {
		latest *latestAssessments
	}
	tests := []struct {
		name string
		args args
		want evaluation.EvaluationStatus
	}{
		{
			name: "compliant",
			args: args{
				latest: newLatestAssessments([]*assessment.AssessmentResult{orchestratortest.MockAssessmentResult1}),
			},
			want: evaluation.EvaluationStatus_EVALUATION_STATUS_COMPLIANT,
		},
		{
			name: "not compliant",
			args: args{
				latest: newLatestAssessments([]*assessment.AssessmentResult{notCompliantResult1}),
			},
			want: evaluation.EvaluationStatus_EVALUATION_STATUS_NOT_COMPLIANT,
		},
		{
			name: "no assessment results for the metrics of the control",
			args: args{
				latest: newLatestAssessments([]*assessment.AssessmentResult{orchestratortest.MockAssessmentResult2}),
			},
			want: evaluation.EvaluationStatus_EVALUATION_STATUS_PENDING,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				orchestratorClient: newOrchestratorClient(t, func(m *mockOrchestratorHandler) {
					m.listAssessmentResultError = errors.New("the orchestrator must not be asked for assessment results")
				}),
			}

			got, err := svc.evaluateSubcontrol(context.Background(), &orchestrator.AuditScope{
				Id:                   evaluationtest.MockAuditScopeId1,
				TargetOfEvaluationId: evaluationtest.MockToeId1,
				CatalogId:            orchestratortest.MockCatalogId1,
			}, orchestratortest.MockSubControl1, tt.args.latest)

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got.GetStatus())
		})
	}
}

func TestService_fetchLatestAssessments(t *testing.T) {
	type args struct {
		targetOfEvaluationId string
	}
	tests := []struct {
		name    string
		args    args
		opts    []func(*mockOrchestratorHandler)
		want    assert.Want[*latestAssessments]
		wantErr assert.WantErr
	}{
		{
			name: "fetches the results of the target of evaluation, indexed by metric",
			args: args{targetOfEvaluationId: evaluationtest.MockToeId1},
			opts: []func(*mockOrchestratorHandler){func(m *mockOrchestratorHandler) {
				m.assessmentResults = []*assessment.AssessmentResult{
					orchestratortest.MockAssessmentResult1,
					orchestratortest.MockAssessmentResult2,
				}
			}},
			want: func(t *testing.T, got *latestAssessments, args ...any) bool {
				return assert.Equal(t, 2, len(got.byMetric)) &&
					assert.Equal(t, 1, len(got.byMetric[orchestratortest.MockMetricId1])) &&
					assert.Equal(t, 1, len(got.byMetric[orchestratortest.MockMetricId2]))
			},
			wantErr: assert.NoError,
		},
		{
			name: "results of other targets of evaluation are not included",
			args: args{targetOfEvaluationId: orchestratortest.MockNonExistentId},
			opts: []func(*mockOrchestratorHandler){func(m *mockOrchestratorHandler) {
				m.assessmentResults = []*assessment.AssessmentResult{orchestratortest.MockAssessmentResult1}
			}},
			want: func(t *testing.T, got *latestAssessments, args ...any) bool {
				return assert.Equal(t, 0, len(got.byMetric))
			},
			wantErr: assert.NoError,
		},
		{
			name: "orchestrator error",
			args: args{targetOfEvaluationId: evaluationtest.MockToeId1},
			opts: []func(*mockOrchestratorHandler){func(m *mockOrchestratorHandler) {
				m.listAssessmentResultError = errors.New("boom")
			}},
			want: assert.Nil[*latestAssessments],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.ErrorContains(t, err, "boom")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{orchestratorClient: newOrchestratorClient(t, tt.opts...)}

			got, err := svc.fetchLatestAssessments(context.Background(), tt.args.targetOfEvaluationId)

			tt.want(t, got)
			tt.wantErr(t, err)
		})
	}
}

// TestService_fetchLatestAssessments_request checks how the orchestrator is asked: once, for the latest result per
// resource and metric of the whole target of evaluation (no metric filter), with the largest page size.
func TestService_fetchLatestAssessments_request(t *testing.T) {
	var (
		handler *mockOrchestratorHandler
		client  orchestratorconnect.OrchestratorClient
	)

	client, handler = newOrchestratorClientWithHandler(t, func(m *mockOrchestratorHandler) {
		m.assessmentResults = []*assessment.AssessmentResult{
			orchestratortest.MockAssessmentResult1,
			orchestratortest.MockAssessmentResult2,
		}
	})
	svc := &Service{orchestratorClient: client}

	_, err := svc.fetchLatestAssessments(context.Background(), evaluationtest.MockToeId1)

	assert.NoError(t, err)
	handler.mu.Lock()
	defer handler.mu.Unlock()
	if assert.Equal(t, 1, len(handler.assessmentResultCalls)) {
		assert.Equal(t, service.DefaultPaginationOpts.MaxPageSize, handler.assessmentResultCalls[0].pageSize)
		assert.True(t, handler.assessmentResultCalls[0].latestByResourceId)
		assert.Equal(t, 0, len(handler.assessmentResultCalls[0].metricIds))
	}
}
