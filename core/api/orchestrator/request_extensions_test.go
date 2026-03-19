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
	"confirmate.io/core/util/assert"

	"google.golang.org/protobuf/proto"
)

func TestGetPayload(t *testing.T) {
	type args struct {
		get  func() proto.Message
		want proto.Message
	}

	tests := []struct {
		name string
		args args
		want assert.Want[proto.Message]
	}{
		{
			name: "create catalog",
			args: args{
				get: func() proto.Message {
					catalog := &Catalog{Id: "catalog-1"}
					return (&CreateCatalogRequest{Catalog: catalog}).GetPayload()
				},
				want: &Catalog{Id: "catalog-1"},
			},
			want: func(t *testing.T, got proto.Message, msgAndArgs ...any) bool {
				want := assert.Is[proto.Message](t, msgAndArgs[0])
				return assert.Equal(t, want, got)
			},
		},
		{
			name: "update catalog",
			args: args{
				get: func() proto.Message {
					catalog := &Catalog{Id: "catalog-2"}
					return (&UpdateCatalogRequest{Catalog: catalog}).GetPayload()
				},
				want: &Catalog{Id: "catalog-2"},
			},
			want: func(t *testing.T, got proto.Message, msgAndArgs ...any) bool {
				want := assert.Is[proto.Message](t, msgAndArgs[0])
				return assert.Equal(t, want, got)
			},
		},
		{
			name: "create compliance attestation",
			args: args{
				get: func() proto.Message {
					attestation := &ComplianceAttestation{Id: "cert-1"}
					return (&CreateComplianceAttestationRequest{ComplianceAttestation: attestation}).GetPayload()
				},
				want: &ComplianceAttestation{Id: "cert-1"},
			},
			want: func(t *testing.T, got proto.Message, msgAndArgs ...any) bool {
				want := assert.Is[proto.Message](t, msgAndArgs[0])
				return assert.Equal(t, want, got)
			},
		},
		{
			name: "update compliance attestation",
			args: args{
				get: func() proto.Message {
					attestation := &ComplianceAttestation{Id: "cert-2"}
					return (&UpdateComplianceAttestationRequest{ComplianceAttestation: attestation}).GetPayload()
				},
				want: &ComplianceAttestation{Id: "cert-2"},
			},
			want: func(t *testing.T, got proto.Message, msgAndArgs ...any) bool {
				want := assert.Is[proto.Message](t, msgAndArgs[0])
				return assert.Equal(t, want, got)
			},
		},
		{
			name: "create audit scope",
			args: args{
				get: func() proto.Message {
					scope := &AuditScope{Id: "scope-1"}
					return (&CreateAuditScopeRequest{AuditScope: scope}).GetPayload()
				},
				want: &AuditScope{Id: "scope-1"},
			},
			want: func(t *testing.T, got proto.Message, msgAndArgs ...any) bool {
				want := assert.Is[proto.Message](t, msgAndArgs[0])
				return assert.Equal(t, want, got)
			},
		},
		{
			name: "update audit scope",
			args: args{
				get: func() proto.Message {
					scope := &AuditScope{Id: "scope-2"}
					return (&UpdateAuditScopeRequest{AuditScope: scope}).GetPayload()
				},
				want: &AuditScope{Id: "scope-2"},
			},
			want: func(t *testing.T, got proto.Message, msgAndArgs ...any) bool {
				want := assert.Is[proto.Message](t, msgAndArgs[0])
				return assert.Equal(t, want, got)
			},
		},
		{
			name: "create metric",
			args: args{
				get: func() proto.Message {
					metric := &assessment.Metric{Id: "metric-1"}
					return (&CreateMetricRequest{Metric: metric}).GetPayload()
				},
				want: &assessment.Metric{Id: "metric-1"},
			},
			want: func(t *testing.T, got proto.Message, msgAndArgs ...any) bool {
				want := assert.Is[proto.Message](t, msgAndArgs[0])
				return assert.Equal(t, want, got)
			},
		},
		{
			name: "update metric",
			args: args{
				get: func() proto.Message {
					metric := &assessment.Metric{Id: "metric-2"}
					return (&UpdateMetricRequest{Metric: metric}).GetPayload()
				},
				want: &assessment.Metric{Id: "metric-2"},
			},
			want: func(t *testing.T, got proto.Message, msgAndArgs ...any) bool {
				want := assert.Is[proto.Message](t, msgAndArgs[0])
				return assert.Equal(t, want, got)
			},
		},
		{
			name: "create target of evaluation",
			args: args{
				get: func() proto.Message {
					toe := &TargetOfEvaluation{Id: "toe-1"}
					return (&CreateTargetOfEvaluationRequest{TargetOfEvaluation: toe}).GetPayload()
				},
				want: &TargetOfEvaluation{Id: "toe-1"},
			},
			want: func(t *testing.T, got proto.Message, msgAndArgs ...any) bool {
				want := assert.Is[proto.Message](t, msgAndArgs[0])
				return assert.Equal(t, want, got)
			},
		},
		{
			name: "update target of evaluation",
			args: args{
				get: func() proto.Message {
					toe := &TargetOfEvaluation{Id: "toe-2"}
					return (&UpdateTargetOfEvaluationRequest{TargetOfEvaluation: toe}).GetPayload()
				},
				want: &TargetOfEvaluation{Id: "toe-2"},
			},
			want: func(t *testing.T, got proto.Message, msgAndArgs ...any) bool {
				want := assert.Is[proto.Message](t, msgAndArgs[0])
				return assert.Equal(t, want, got)
			},
		},
		{
			name: "register assessment tool",
			args: args{
				get: func() proto.Message {
					tool := &AssessmentTool{Id: "tool-1"}
					return (&RegisterAssessmentToolRequest{Tool: tool}).GetPayload()
				},
				want: &AssessmentTool{Id: "tool-1"},
			},
			want: func(t *testing.T, got proto.Message, msgAndArgs ...any) bool {
				want := assert.Is[proto.Message](t, msgAndArgs[0])
				return assert.Equal(t, want, got)
			},
		},
		{
			name: "update assessment tool",
			args: args{
				get: func() proto.Message {
					tool := &AssessmentTool{Id: "tool-2"}
					return (&UpdateAssessmentToolRequest{Tool: tool}).GetPayload()
				},
				want: &AssessmentTool{Id: "tool-2"},
			},
			want: func(t *testing.T, got proto.Message, msgAndArgs ...any) bool {
				want := assert.Is[proto.Message](t, msgAndArgs[0])
				return assert.Equal(t, want, got)
			},
		},
		{
			name: "update metric configuration",
			args: args{
				get: func() proto.Message {
					cfg := &assessment.MetricConfiguration{MetricId: "metric-3"}
					return (&UpdateMetricConfigurationRequest{Configuration: cfg}).GetPayload()
				},
				want: &assessment.MetricConfiguration{MetricId: "metric-3"},
			},
			want: func(t *testing.T, got proto.Message, msgAndArgs ...any) bool {
				want := assert.Is[proto.Message](t, msgAndArgs[0])
				return assert.Equal(t, want, got)
			},
		},
		{
			name: "update metric implementation",
			args: args{
				get: func() proto.Message {
					impl := &assessment.MetricImplementation{MetricId: "metric-4"}
					return (&UpdateMetricImplementationRequest{Implementation: impl}).GetPayload()
				},
				want: &assessment.MetricImplementation{MetricId: "metric-4"},
			},
			want: func(t *testing.T, got proto.Message, msgAndArgs ...any) bool {
				want := assert.Is[proto.Message](t, msgAndArgs[0])
				return assert.Equal(t, want, got)
			},
		},
		{
			name: "store assessment result",
			args: args{
				get: func() proto.Message {
					result := &assessment.AssessmentResult{Id: "result-1"}
					return (&StoreAssessmentResultRequest{Result: result}).GetPayload()
				},
				want: &assessment.AssessmentResult{Id: "result-1"},
			},
			want: func(t *testing.T, got proto.Message, msgAndArgs ...any) bool {
				want := assert.Is[proto.Message](t, msgAndArgs[0])
				return assert.Equal(t, want, got)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.args.get()
			tt.want(t, got, tt.args.want)
		})
	}
}

func TestGetTargetOfEvaluationId(t *testing.T) {
	type args struct {
		get  func() string
		want string
	}

	tests := []struct {
		name string
		args args
		want assert.Want[string]
	}{
		{
			name: "list assessment results",
			args: args{
				get: func() string {
					return (&ListAssessmentResultsRequest{Filter: &ListAssessmentResultsRequest_Filter{TargetOfEvaluationId: ref("toe-7")}}).GetTargetOfEvaluationId()
				},
				want: "toe-7",
			},
			want: func(t *testing.T, got string, msgAndArgs ...any) bool {
				want := assert.Is[string](t, msgAndArgs[0])
				return assert.Equal(t, want, got)
			},
		},
		{
			name: "list audit scopes",
			args: args{
				get: func() string {
					return (&ListAuditScopesRequest{Filter: &ListAuditScopesRequest_Filter{TargetOfEvaluationId: ref("toe-8")}}).GetTargetOfEvaluationId()
				},
				want: "toe-8",
			},
			want: func(t *testing.T, got string, msgAndArgs ...any) bool {
				want := assert.Is[string](t, msgAndArgs[0])
				return assert.Equal(t, want, got)
			},
		},
		{
			name: "nil chains resolve to empty string",
			args: args{
				get: func() string {
					return (&ListAuditScopesRequest{}).GetTargetOfEvaluationId()
				},
				want: "",
			},
			want: func(t *testing.T, got string, msgAndArgs ...any) bool {
				want := assert.Is[string](t, msgAndArgs[0])
				return assert.Equal(t, want, got)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.args.get()
			tt.want(t, tt.args.want, got)
		})
	}
}

func ref(v string) *string {
	return &v
}
