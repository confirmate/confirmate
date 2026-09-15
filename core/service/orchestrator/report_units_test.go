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

func TestPdfHeaderRow(t *testing.T) {
	scope := &orchestrator.AuditScope{Name: "Test Scope"}

	// Below the pass threshold: renders the amber "ACTION REQUIRED" pill.
	row := pdfHeaderRow(scope, reportSummary{evaluatedMetrics: 2, compliantMetrics: 1})
	assert.NotNil(t, row)

	// At/above the pass threshold: renders the green "PASS" pill.
	row = pdfHeaderRow(scope, reportSummary{evaluatedMetrics: 10, compliantMetrics: 10})
	assert.NotNil(t, row)
}
