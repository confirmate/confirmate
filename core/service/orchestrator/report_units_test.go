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

package orchestrator

import (
	"testing"

	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/util/assert"

	"google.golang.org/protobuf/types/known/structpb"
)

func TestUserDisplayName(t *testing.T) {
	assert.Equal(t, "Jane Doe", userDisplayName(&orchestrator.User{Id: "u1", FirstName: new("Jane"), LastName: new("Doe")}))
	assert.Equal(t, "jdoe", userDisplayName(&orchestrator.User{Id: "u1", Username: new("jdoe")}))
	assert.Equal(t, "u1", userDisplayName(&orchestrator.User{Id: "u1"}))
}

func TestTruncate(t *testing.T) {
	assert.Equal(t, "hello", truncate("hello", 10))
	assert.Equal(t, "hell…", truncate("hello world", 5))
}

func TestFormatStructValue(t *testing.T) {
	assert.Equal(t, "—", formatStructValue(nil))
	assert.Equal(t, "true", formatStructValue(structpb.NewBoolValue(true)))
}

func TestFormatAssessmentCondition(t *testing.T) {
	assert.Equal(t, "prop == true (Observed: false)", formatAssessmentCondition(&assessment.AssessmentResult{
		ComplianceDetails: []*assessment.ComparisonResult{{
			Property:    "prop",
			Operator:    "==",
			TargetValue: structpb.NewBoolValue(true),
			Value:       structpb.NewBoolValue(false),
		}},
	}))
	assert.Equal(t, "Expected == true", formatAssessmentCondition(&assessment.AssessmentResult{
		MetricConfiguration: &assessment.MetricConfiguration{
			Operator:    "==",
			TargetValue: structpb.NewBoolValue(true),
		},
	}))
	assert.Equal(t, "—", formatAssessmentCondition(&assessment.AssessmentResult{}))
}

func TestFlattenControls(t *testing.T) {
	sub := &orchestrator.Control{Id: "sub"}
	top := &orchestrator.Control{Id: "top", Controls: []*orchestrator.Control{sub}}
	flat := flattenControls([]*orchestrator.Control{top})
	assert.Equal(t, 2, len(flat))
	assert.Equal(t, "top", flat[0].GetId())
	assert.Equal(t, "sub", flat[1].GetId())
}

func TestReportControlRow_StatusLabel(t *testing.T) {
	assert.Equal(t, "No Metrics", reportControlRow{}.statusLabel())
	assert.Equal(t, "Not Evaluated", reportControlRow{metrics: []reportMetricRow{{evaluated: false}}}.statusLabel())
	assert.Equal(t, "Passed", reportControlRow{metrics: []reportMetricRow{{evaluated: true, compliant: true}}}.statusLabel())
	assert.Equal(t, "Action Required", reportControlRow{metrics: []reportMetricRow{
		{evaluated: true, compliant: true},
		{evaluated: true, compliant: false},
	}}.statusLabel())
}

func TestReportControlRow_StatusWithCounts(t *testing.T) {
	tests := []struct {
		name string
		row  reportControlRow
		want string
	}{
		{
			name: "no metrics renders just the label",
			row:  reportControlRow{},
			want: "No Metrics",
		},
		{
			name: "not evaluated counts all metrics as requiring action",
			row: reportControlRow{metrics: []reportMetricRow{
				{evaluated: false},
				{evaluated: false},
			}},
			want: "Not Evaluated (2/2)",
		},
		{
			name: "passed counts compliant metrics",
			row: reportControlRow{metrics: []reportMetricRow{
				{evaluated: true, compliant: true},
				{evaluated: true, compliant: true},
				{evaluated: true, compliant: true},
				{evaluated: true, compliant: true},
				{evaluated: true, compliant: true},
			}},
			want: "Passed (5/5)",
		},
		{
			name: "action required counts non-compliant and non-evaluated metrics",
			row: reportControlRow{metrics: []reportMetricRow{
				{evaluated: true, compliant: true},
				{evaluated: true, compliant: false},
			}},
			want: "Action Required (1/2)",
		},
		{
			name: "action required includes not-yet-evaluated metrics in the action count",
			row: reportControlRow{metrics: []reportMetricRow{
				{evaluated: true, compliant: true},
				{evaluated: true, compliant: false},
				{evaluated: false},
			}},
			want: "Action Required (2/3)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.row.statusWithCounts())
		})
	}
}

func TestSummarizeReportControls(t *testing.T) {
	tests := []struct {
		name     string
		controls []reportControlRow
		want     reportSummary
	}{
		{
			name:     "no controls",
			controls: nil,
			want:     reportSummary{},
		},
		{
			name: "action required counts controls, not non-compliant metrics",
			controls: []reportControlRow{
				// One control with three non-compliant metrics: still just one "Action Required".
				{metrics: []reportMetricRow{
					{name: "m1", evaluated: true, compliant: false},
					{name: "m2", evaluated: true, compliant: false},
					{name: "m3", evaluated: true, compliant: false},
				}},
				// Passed: two compliant metrics.
				{metrics: []reportMetricRow{
					{name: "m4", evaluated: true, compliant: true},
					{name: "m5", evaluated: true, compliant: true},
				}},
				// Not evaluated: doesn't count as action required.
				{metrics: []reportMetricRow{
					{name: "m6", evaluated: false},
				}},
				// No metrics at all.
				{},
			},
			want: reportSummary{
				controls:         4,
				totalMetrics:     6,
				compliantMetrics: 2,
				evaluatedMetrics: 5,
				actionRequired:   1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// reportSummary contains unexported fields, so the comparison needs to export them.
			assert.Equal(t, tt.want, summarizeReportControls(tt.controls), assert.CompareAllUnexported())
		})
	}
}

func TestControlStatusChip(t *testing.T) {
	assert.Equal(t, pdfGreen, controlStatusChip("Passed").text)
	assert.Equal(t, pdfAmber, controlStatusChip("Action Required").text)
	assert.Equal(t, pdfGray, controlStatusChip("Not Evaluated").text)
	assert.Equal(t, pdfGray, controlStatusChip("No Metrics").text)
}

func TestImplementationChip(t *testing.T) {
	cases := []orchestrator.ControlInScopeState{
		orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_OPEN,
		orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_IN_PROGRESS,
		orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_IMPLEMENTED,
		orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_READY_FOR_REVIEW,
		orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_ACCEPTED,
		orchestrator.ControlInScopeState_CONTROL_IN_SCOPE_STATE_UNSPECIFIED,
	}
	for _, c := range cases {
		chip := implementationChip(c)
		assert.NotNil(t, chip.text)
		assert.NotNil(t, chip.bg)
	}
}

func TestCompliancePercent(t *testing.T) {
	assert.Equal(t, float64(0), reportSummary{}.compliancePercent())
	assert.Equal(t, float64(50), reportSummary{evaluatedMetrics: 2, compliantMetrics: 1}.compliancePercent())
	assert.Equal(t, float64(100), reportSummary{evaluatedMetrics: 2, compliantMetrics: 2}.compliancePercent())
}

func TestBuildReportMetricRow(t *testing.T) {
	m := &assessment.Metric{Id: "metric1", Name: "TestMetric"}

	// No results: metric hasn't been evaluated for any resource yet.
	row := buildReportMetricRow(m, nil)
	assert.Equal(t, false, row.evaluated)
	assert.Equal(t, "—", row.targetComponent)
	assert.Equal(t, "—", row.condition)

	// Multiple resources, mixed compliance: overall compliant is false, the non-compliant
	// resource's condition/comment is surfaced, and the resource count is appended.
	row = buildReportMetricRow(m, []*assessment.AssessmentResult{
		{Compliant: true, ResourceId: "res1"},
		{Compliant: false, ResourceId: "res2", ResourceTypes: []string{"VirtualMachine"}, ComplianceComment: "disk not encrypted"},
	})
	assert.Equal(t, true, row.evaluated)
	assert.Equal(t, false, row.compliant)
	assert.Equal(t, "disk not encrypted", row.complianceComment)
	assert.Equal(t, "VirtualMachine (×2 resources)", row.targetComponent)

	// Single resource, fully compliant: no resource-count suffix.
	row = buildReportMetricRow(m, []*assessment.AssessmentResult{
		{Compliant: true, ResourceId: "res1"},
	})
	assert.Equal(t, true, row.compliant)
	assert.Equal(t, "res1", row.targetComponent)
}

func TestPdfHeaderRow(t *testing.T) {
	scope := &orchestrator.AuditScope{Name: "Test Scope"}

	// Below the pass threshold: renders the amber "ACTION REQUIRED" pill.
	row := pdfHeaderRow(scope, reportSummary{evaluatedMetrics: 2, compliantMetrics: 1})
	assert.NotNil(t, row)

	// At/above the pass threshold: renders the green "PASS" pill.
	row = pdfHeaderRow(scope, reportSummary{evaluatedMetrics: 10, compliantMetrics: 10})
	assert.NotNil(t, row)
}
