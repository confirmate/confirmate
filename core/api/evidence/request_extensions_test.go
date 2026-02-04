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

package evidence

import (
	"testing"

	"confirmate.io/core/util/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestStoreEvidenceRequest_TargetOfEvaluationId(t *testing.T) {
	req := &StoreEvidenceRequest{
		Evidence: &Evidence{
			TargetOfEvaluationId: "toe-1",
		},
	}

	assert.Equal(t, "toe-1", req.GetTargetOfEvaluationId())
}

func TestStoreEvidenceRequest_EvidenceId(t *testing.T) {
	tests := []struct {
		name string
		req  *StoreEvidenceRequest
		want string
	}{
		{
			name: "nil evidence",
			req:  &StoreEvidenceRequest{},
			want: "",
		},
		{
			name: "happy path",
			req: &StoreEvidenceRequest{
				Evidence: &Evidence{Id: "e-1"},
			},
			want: "e-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.req.GetEvidenceId())
		})
	}
}

func TestStoreEvidenceRequest_GetPayload(t *testing.T) {
	ev := &Evidence{
		Id:                   "e-1",
		TargetOfEvaluationId: "toe-1",
		Timestamp:            timestamppb.Now(),
	}
	req := &StoreEvidenceRequest{Evidence: ev}

	payload, ok := req.GetPayload().(*Evidence)
	if !assert.True(t, ok) {
		return
	}
	assert.Equal(t, ev, payload)
}

func TestUpdateResourceRequest_TargetOfEvaluationId(t *testing.T) {
	req := &UpdateResourceRequest{
		Resource: &Resource{
			TargetOfEvaluationId: "toe-2",
		},
	}

	assert.Equal(t, "toe-2", req.GetTargetOfEvaluationId())
}
