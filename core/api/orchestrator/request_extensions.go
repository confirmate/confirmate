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

package orchestrator

import (
	"google.golang.org/protobuf/proto"
)

// GetPayload returns the embedded catalog from [CreateCatalogRequest].
func (r *CreateCatalogRequest) GetPayload() proto.Message {
	return r.Catalog
}

// GetPayload returns the embedded catalog from [UpdateCatalogRequest].
func (r *UpdateCatalogRequest) GetPayload() proto.Message {
	return r.Catalog
}

// GetPayload returns the embedded certificate from [CreateCertificateRequest].
func (r *CreateCertificateRequest) GetPayload() proto.Message {
	return r.Certificate
}

// GetPayload returns the embedded certificate from [UpdateCertificateRequest].
func (r *UpdateCertificateRequest) GetPayload() proto.Message {
	return r.Certificate
}

// GetPayload returns the embedded audit_scope from [CreateAuditScopeRequest].
func (r *CreateAuditScopeRequest) GetPayload() proto.Message {
	return r.AuditScope
}

// GetPayload returns the embedded audit_scope from [UpdateAuditScopeRequest].
func (r *UpdateAuditScopeRequest) GetPayload() proto.Message {
	return r.AuditScope
}

// GetPayload returns the embedded metric from [CreateMetricRequest].
func (r *CreateMetricRequest) GetPayload() proto.Message {
	return r.Metric
}

// GetPayload returns the embedded metric from [UpdateMetricRequest].
func (r *UpdateMetricRequest) GetPayload() proto.Message {
	return r.Metric
}

// GetPayload returns the embedded target_of_evaluation from [CreateTargetOfEvaluationRequest].
func (r *CreateTargetOfEvaluationRequest) GetPayload() proto.Message {
	return r.TargetOfEvaluation
}

// GetPayload returns the embedded target_of_evaluation from [UpdateTargetOfEvaluationRequest].
func (r *UpdateTargetOfEvaluationRequest) GetPayload() proto.Message {
	return r.TargetOfEvaluation
}

// GetPayload returns the embedded tool from [UpdateAssessmentToolRequest].
func (r *UpdateAssessmentToolRequest) GetPayload() proto.Message {
	return r.Tool
}

// GetPayload returns the embedded configuration from [UpdateMetricConfigurationRequest].
func (r *UpdateMetricConfigurationRequest) GetPayload() proto.Message {
	return r.Configuration
}

// GetPayload returns the embedded implementation from [UpdateMetricImplementationRequest].
func (r *UpdateMetricImplementationRequest) GetPayload() proto.Message {
	return r.Implementation
}
