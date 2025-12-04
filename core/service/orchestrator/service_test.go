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
	"confirmate.io/core/persistence"
	"confirmate.io/core/persistence/persistencetest"
	"confirmate.io/core/service/orchestrator/orchestratortest"
	"confirmate.io/core/util/assert"

	"connectrpc.com/connect"
)

func TestService_ListTargetsOfEvaluation(t *testing.T) {
	type args struct {
		req *orchestrator.ListTargetsOfEvaluationRequest
	}
	type fields struct {
		db *persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*orchestrator.ListTargetsOfEvaluationResponse]
		wantErr assert.WantErr[*connect.Error]
	}{
		{
			name: "happy path",
			args: args{
				req: &orchestrator.ListTargetsOfEvaluationRequest{},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d *persistence.DB) {
					err := d.Create(orchestratortest.MockTargetOfEvaluation1)
					assert.NoError(t, err)
				}),
			},
			want: func(t *testing.T, got *orchestrator.ListTargetsOfEvaluationResponse, args ...any) bool {
				return assert.NotNil(t, got) &&
					assert.Equal(t, 1, len(got.TargetsOfEvaluation)) &&
					assert.Equal(t, orchestratortest.MockTargetOfEvaluation1.Id, got.TargetsOfEvaluation[0].Id)
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &service{
				db: tt.fields.db,
			}
			res, err := svc.ListTargetsOfEvaluation(context.Background(), connect.NewRequest(tt.args.req))
			assert.WantResponse(t, res, err, tt.want, tt.wantErr)
		})
	}
}
