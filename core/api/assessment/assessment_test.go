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

package assessment

import (
	"testing"

	"buf.build/go/protovalidate"
	"confirmate.io/core/util"
	"confirmate.io/core/util/assert"
	"confirmate.io/core/util/testdata"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func Test_ValidateAssessmentResult(t *testing.T) {
	timestamp := timestamppb.Now()

	type args struct {
		AssessmentResult *AssessmentResult
	}

	tests := []struct {
		name    string
		args    args
		wantErr assert.WantErr
	}{
		{
			name: "happy path",
			args: args{
				&AssessmentResult{
					Id:        testdata.MockAssessmentResultID,
					CreatedAt: timestamp,
					MetricId:  testdata.MockMetricID1,
					MetricConfiguration: &MetricConfiguration{
						Operator:             "==",
						MetricId:             testdata.MockMetricID1,
						TargetValue:          testdata.MockMetricConfigurationTargetValueString,
						TargetOfEvaluationId: testdata.MockTargetOfEvaluationID1,
					},
					Compliant:            false,
					ComplianceComment:    "Some comment",
					EvidenceId:           testdata.MockEvidenceID1,
					ResourceId:           "myResource",
					ResourceTypes:        []string{"Resource"},
					TargetOfEvaluationId: testdata.MockTargetOfEvaluationID1,
					ToolId:               util.Ref(AssessmentToolId),
					HistoryUpdatedAt:     timestamp,
					History: []*Record{
						{
							EvidenceRecordedAt: timestamp,
							EvidenceId:         testdata.MockEvidenceID1,
						},
					},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "missing assessment result id",
			args: args{
				&AssessmentResult{
					// Empty id
					Id:       "",
					MetricId: testdata.MockMetricID1,
					MetricConfiguration: &MetricConfiguration{
						Operator:    "==",
						TargetValue: testdata.MockMetricConfigurationTargetValueString,
					},
					EvidenceId:    testdata.MockEvidenceID1,
					ResourceTypes: []string{"Resource"},
				},
			},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsValidationError(t, err, "id")
			},
		},
		{
			name: "wrong length of assessment result id",
			args: args{
				&AssessmentResult{
					// Only 4 characters
					Id:       "1234",
					MetricId: testdata.MockMetricID1,
					MetricConfiguration: &MetricConfiguration{
						Operator:    "==",
						TargetValue: testdata.MockMetricConfigurationTargetValueString,
					},
					EvidenceId:    testdata.MockEvidenceID1,
					ResourceTypes: []string{"Resource"},
				},
			},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsValidationError(t, err, "id")
			},
		},
		{
			name: "wrong format of assessment result id",
			args: args{
				&AssessmentResult{
					// Wrong format: 'x' not allowed (no hexadecimal character)
					Id:       "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
					MetricId: testdata.MockMetricID1,
					MetricConfiguration: &MetricConfiguration{
						Operator:    "==",
						TargetValue: testdata.MockMetricConfigurationTargetValueString,
					},
					EvidenceId:    testdata.MockEvidenceID1,
					ResourceTypes: []string{"Resource"},
				},
			},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsValidationError(t, err, "id")
			},
		},
		{
			name: "missing assessment result timestamp",
			args: args{
				&AssessmentResult{
					Id:       testdata.MockAssessmentResultID,
					MetricId: testdata.MockMetricID1,
					MetricConfiguration: &MetricConfiguration{
						Operator:    ">",
						TargetValue: testdata.MockMetricConfigurationTargetValueString,
					},
					EvidenceId:    testdata.MockEvidenceID1,
					ResourceTypes: []string{"Resource"},
				},
			},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsValidationError(t, err, "created_at")
			},
		},
		{
			name: "missing assessment result metric id",
			args: args{
				&AssessmentResult{
					Id:        testdata.MockAssessmentResultID,
					CreatedAt: timestamppb.Now(),
					MetricConfiguration: &MetricConfiguration{
						Operator:    "<",
						TargetValue: testdata.MockMetricConfigurationTargetValueString,
					},
					ResourceTypes: []string{"Resource"},
					EvidenceId:    testdata.MockEvidenceID1,
				},
			},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsValidationError(t, err, "metric_id")
			},
		},
		{
			name: "missing assessment result resource types",
			args: args{
				&AssessmentResult{
					Id:        testdata.MockAssessmentResultID,
					CreatedAt: timestamppb.Now(),
					MetricId:  testdata.MockMetricID1,
					MetricConfiguration: &MetricConfiguration{
						MetricId:             testdata.MockMetricID1,
						Operator:             "==",
						TargetValue:          testdata.MockMetricConfigurationTargetValueString,
						TargetOfEvaluationId: testdata.MockTargetOfEvaluationID1,
					},
					ResourceId: "myResource",
					EvidenceId: testdata.MockEvidenceID1,
				},
			},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsValidationError(t, err, "resource_types")
			},
		},
		{
			name: "missing assessment result metric configuration",
			args: args{
				&AssessmentResult{
					Id:            testdata.MockAssessmentResultID,
					CreatedAt:     timestamppb.Now(),
					MetricId:      testdata.MockMetricID1,
					EvidenceId:    testdata.MockEvidenceID1,
					ResourceTypes: []string{"Resource"},
				},
			},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsValidationError(t, err, "metric_configuration")
			},
		},
		{
			name: "missing assessment result metric configuration operator",
			args: args{
				&AssessmentResult{
					Id:        testdata.MockAssessmentResultID,
					CreatedAt: timestamppb.Now(),
					MetricId:  testdata.MockMetricID1,
					MetricConfiguration: &MetricConfiguration{
						TargetValue: testdata.MockMetricConfigurationTargetValueString,
						MetricId:    testdata.MockMetricID1,
					},
					EvidenceId:    testdata.MockEvidenceID1,
					ResourceTypes: []string{"Resource"},
				},
			},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsValidationError(t, err, "metric_configuration.target_of_evaluation_id")
			},
		},
		{
			name: "missing assessment result metric configuration target value",
			args: args{
				&AssessmentResult{
					Id:        testdata.MockAssessmentResultID,
					CreatedAt: timestamppb.Now(),
					MetricId:  testdata.MockMetricID1,
					MetricConfiguration: &MetricConfiguration{
						Operator: "<",
					},
					EvidenceId:    testdata.MockEvidenceID1,
					ResourceTypes: []string{"Resource"},
				},
			},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsValidationError(t, err, "metric_configuration.target_value")
			},
		},
		{
			name: "missing assessment result evidence id",
			args: args{
				&AssessmentResult{
					Id:        testdata.MockAssessmentResultID,
					CreatedAt: timestamppb.Now(),
					MetricId:  testdata.MockMetricID1,
					MetricConfiguration: &MetricConfiguration{
						Operator:    ">",
						MetricId:    testdata.MockMetricID1,
						TargetValue: testdata.MockMetricConfigurationTargetValueString,
					},
					ResourceTypes: []string{"Resource"},
				},
			},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.IsValidationError(t, err, "evidence_id")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := protovalidate.Validate(tt.args.AssessmentResult)
			tt.wantErr(t, err)
		})
	}
}
