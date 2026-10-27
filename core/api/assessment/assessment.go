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

	"confirmate.io/core/util/timestamp"
)

type ResultHookFunc func(ctx context.Context, result *AssessmentResult, err error)

const (
	DefaultNonCompliantMessage = "The result of the metric indicates that the resource contains properties that are not compliant with the target value."
	DefaultCompliantMessage    = "The result of the metric shows that the evidence is compliant to the target value."
	AdditionalDetailsMessage   = "Additional details can be found in the compliance details."
)

const AssessmentToolId = "Clouditor Assessment"

// NormalizeTimestampAssessmentResult normalizes the timestamp in the given assessment result to microsecond precision.
func NormalizeTimestampAssessmentResult(ar *AssessmentResult) {
	if ar.CreatedAt != nil {
		ar.CreatedAt = timestamp.New(ar.CreatedAt.AsTime())
	}
	if ar.HistoryUpdatedAt != nil {
		ar.HistoryUpdatedAt = timestamp.New(ar.HistoryUpdatedAt.AsTime())
	}
	if ar.History != nil {
		for _, h := range ar.History {
			if h.EvidenceRecordedAt != nil {
				h.EvidenceRecordedAt = timestamp.New(h.EvidenceRecordedAt.AsTime())
			}
		}
	}
	if ar.ValidUntil != nil {
		ar.ValidUntil = timestamp.New(ar.ValidUntil.AsTime())
	}
}
