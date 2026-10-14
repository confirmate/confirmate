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
	"confirmate.io/core/persistence"
	"confirmate.io/core/persistence/persistencetest"
	"confirmate.io/core/service/orchestrator/orchestratortest"
	"confirmate.io/core/util/assert"
)

func Test_loadByIds(t *testing.T) {
	type args struct {
		db  persistence.DB
		ids []string
	}
	tests := []struct {
		name    string
		args    args
		want    assert.Want[[]*assessment.AssessmentResult]
		wantErr assert.WantErr
	}{
		{
			name: "no ids",
			args: args{
				db: persistencetest.NewInMemoryDB(t, types, nil),
			},
			want:    assert.Nil[[]*assessment.AssessmentResult],
			wantErr: assert.NoError,
		},
		{
			name: "database error",
			args: args{
				db:  persistencetest.ListErrorDB(t, persistence.ErrDatabase, types, nil),
				ids: []string{orchestratortest.MockAssessmentResult1.Id},
			},
			want: assert.Nil[[]*assessment.AssessmentResult],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.ErrorContains(t, err, persistence.ErrDatabase.Error())
			},
		},
		{
			name: "returns records in the order of the given ids, skipping unknown ones",
			args: args{
				db: persistencetest.NewInMemoryDB(t, types, nil, func(d persistence.DB) {
					assert.NoError(t, d.Create(orchestratortest.MockAssessmentResult1))
					assert.NoError(t, d.Create(orchestratortest.MockAssessmentResult2))
				}),
				ids: []string{
					orchestratortest.MockAssessmentResult2.Id,
					orchestratortest.MockNonExistentId,
					orchestratortest.MockAssessmentResult1.Id,
				},
			},
			want: func(t *testing.T, got []*assessment.AssessmentResult, args ...any) bool {
				if !assert.Equal(t, 2, len(got)) {
					return false
				}
				assert.Equal(t, orchestratortest.MockAssessmentResult2.Id, got[0].Id)
				return assert.Equal(t, orchestratortest.MockAssessmentResult1.Id, got[1].Id)
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := loadByIds[*assessment.AssessmentResult](tt.args.db, tt.args.ids)

			tt.wantErr(t, err)
			tt.want(t, got)
		})
	}
}
