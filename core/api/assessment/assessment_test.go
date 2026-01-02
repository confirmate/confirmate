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

	"confirmate.io/core/api"
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
		name     string
		args     args
		wantResp string
		wantErr  assert.ErrorAssertionFunc
	}{
		{
			name: "Missing assessment result id",
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
			wantResp: "",
			wantErr: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorContains(t, err, "id: value is empty, which is not a valid UUID")
			},
		},
		{
			name: "Wrong length of assessment result id",
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
			wantResp: "",
			wantErr: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorContains(t, err, "id: value must be a valid UUID")
			},
		},
		{
			name: "Wrong format of assessment result id",
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
			wantResp: "",
			wantErr: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorContains(t, err, "id: value must be a valid UUID")
			},
		},
		{
			name: "Missing assessment result timestamp",
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
			wantResp: "",
			wantErr: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorContains(t, err, "value is required")
			},
		},
		{
			name: "Missing assessment result metric id",
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
			wantResp: "",
			wantErr: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorContains(t, err, "metric_id: value length must be at least 1 characters")
			},
		},
		{
			name: "Missing assessment result resource types",
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
			wantResp: "",
			wantErr: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorContains(t, err, "resource_types: value must contain at least 1 item(s)")
			},
		},
		{
			name: "Missing assessment result metric configuration",
			args: args{
				&AssessmentResult{
					Id:            testdata.MockAssessmentResultID,
					CreatedAt:     timestamppb.Now(),
					MetricId:      testdata.MockMetricID1,
					EvidenceId:    testdata.MockEvidenceID1,
					ResourceTypes: []string{"Resource"},
				},
			},
			wantResp: "",
			wantErr: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorContains(t, err, "metric_configuration: value is required ")
			},
		},
		{
			name: "Missing assessment result metric configuration operator",
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
			wantResp: "",
			wantErr: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorContains(t, err, "metric_configuration.target_of_evaluation_id: value is empty, which is not a valid UUID")
			},
		},
		{
			name: "Missing assessment result metric configuration target value",
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
			wantResp: "",
			wantErr: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorContains(t, err, "metric_configuration.target_value: value is required")
			},
		},
		{
			name: "Missing assessment result evidence id",
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
			wantResp: "",
			wantErr: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorContains(t, err, "evidence_id: value is empty, which is not a valid UUID")
			},
		},
		{
			name: "Valid assessment result",
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
			wantResp: "",
			wantErr: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := api.Validate(tt.args.AssessmentResult)
			if tt.wantErr != nil {
				tt.wantErr(t, err, tt.args)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
