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
	"context"
	"errors"

	"google.golang.org/protobuf/proto"
)

type ResultHookFunc func(ctx context.Context, result *AssessmentResult, err error)

var (
	ErrMetricConfigurationMissing            = errors.New("metric configuration in assessment result is missing")
	ErrMetricConfigurationOperatorMissing    = errors.New("operator in metric data is missing")
	ErrMetricConfigurationTargetValueMissing = errors.New("target value in metric data is missing")
)

const (
	DefaultNonCompliantMessage = "The result of the metric indicates that the resource contains properties that are not compliant with the target value."
	DefaultCompliantMessage    = "The result of the metric shows that the evidence is compliant to the target value."
	AdditionalDetailsMessage   = "Additional details can be found in the comparison below."
)

const AssessmentToolId = "Clouditor Assessment"

func (req *AssessEvidenceRequest) GetPayload() proto.Message {
	return req.Evidence
}

// GetTargetOfEvaluationId is a shortcut to implement TargetOfEvaluationRequest. It returns the target of evaluation ID of the inner
// object.
func (req *AssessEvidenceRequest) GetTargetOfEvaluationId() string {
	return req.GetEvidence().GetTargetOfEvaluationId()
}
