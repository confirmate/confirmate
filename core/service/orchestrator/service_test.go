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
	"context"
	"testing"

	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/db"
	"confirmate.io/core/db/dbtest"
	"confirmate.io/core/util/testutil/assert"
	"connectrpc.com/connect"
)

func Test_service_ListTargetsOfEvaluation(t *testing.T) {
	tests := []struct {
		name   string
		fields struct {
			storage *db.Storage
		}
		args struct {
			ctx context.Context
			req *connect.Request[orchestrator.ListTargetsOfEvaluationRequest]
		}
		want    *connect.Response[orchestrator.ListTargetsOfEvaluationResponse]
		wantErr bool
	}{
		{
			name: "happy path",
			fields: struct {
				storage *db.Storage
			}{
				storage: dbtest.NewInMemoryStorage(t, types, joinTable, func(s *db.Storage) {
					// Create a sample TargetOfEvaluation entry
					err := s.Create(&orchestrator.TargetOfEvaluation{
						Id:   "1",
						Name: "TOE1",
					})
					if err != nil {
						t.Fatalf("could not create TOE: %v", err)
					}
				}),
			},
			args: struct {
				ctx context.Context
				req *connect.Request[orchestrator.ListTargetsOfEvaluationRequest]
			}{
				ctx: context.Background(),
				req: connect.NewRequest(&orchestrator.ListTargetsOfEvaluationRequest{}),
			},
			want: connect.NewResponse(&orchestrator.ListTargetsOfEvaluationResponse{
				TargetsOfEvaluation: []*orchestrator.TargetOfEvaluation{
					{
						Id:   "1",
						Name: "TOE1",
					},
				},
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &service{
				storage: tt.fields.storage,
			}
			got, gotErr := svc.ListTargetsOfEvaluation(context.Background(), nil)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ListTargetsOfEvaluation() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ListTargetsOfEvaluation() succeeded unexpectedly")
			}
			assert.Equal(t, tt.want.Msg, got.Msg)
		})
	}
}
