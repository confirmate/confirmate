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

package service

import (
	"testing"

	"confirmate.io/core/api"
	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/persistence"
	"confirmate.io/core/persistence/persistencetest"
	"confirmate.io/core/util/assert"
)

func TestPaginateSlice(t *testing.T) {
	type args struct {
		req    api.PaginatedRequest
		values []int
		opts   PaginationOpts
	}
	tests := []struct {
		name     string
		args     args
		wantPage assert.Want[[]int]
		wantNbt  assert.Want[string]
		wantErr  assert.WantErr
	}{
		{
			name: "first page",
			args: args{
				req: &orchestrator.ListAssessmentResultsRequest{
					PageSize:  2,
					PageToken: "",
				},
				values: []int{1, 2, 3, 4, 5},
				opts:   PaginationOpts{10, 10},
			},
			wantPage: func(t *testing.T, got []int) bool {
				return assert.Equal(t, []int{1, 2}, got)
			},
			wantNbt: func(t *testing.T, got string) bool {
				return assert.Equal(t, "CAIQAg==", got)
			},
			wantErr: assert.Nil[error],
		},
		{
			name: "next page",
			args: args{
				req: &orchestrator.ListAssessmentResultsRequest{
					PageSize:  2,
					PageToken: "CAIQAg==",
				},
				values: []int{1, 2, 3, 4, 5},
				opts:   PaginationOpts{10, 10},
			},
			wantPage: func(t *testing.T, got []int) bool {
				return assert.Equal(t, []int{3, 4}, got)
			},
			wantNbt: func(t *testing.T, got string) bool {
				return assert.Equal(t, "CAQQAg==", got)
			},
			wantErr: assert.Nil[error],
		},
		{
			name: "last page",
			args: args{
				req: &orchestrator.ListAssessmentResultsRequest{
					PageSize:  2,
					PageToken: "CAQQAg==",
				},
				values: []int{1, 2, 3, 4, 5},
				opts:   PaginationOpts{10, 10},
			},
			wantPage: func(t *testing.T, got []int) bool {
				return assert.Equal(t, []int{5}, got)
			},
			wantNbt: func(t *testing.T, got string) bool {
				return assert.Equal(t, "", got)
			},
			wantErr: assert.Nil[error],
		},
		{
			name: "empty slice",
			args: args{
				req: &orchestrator.ListAssessmentResultsRequest{
					PageSize:  2,
					PageToken: "",
				},
				values: []int{},
				opts:   PaginationOpts{10, 10},
			},
			wantPage: func(t *testing.T, got []int) bool {
				return assert.Equal(t, []int{}, got)
			},
			wantNbt: assert.Empty[string],
			wantErr: assert.Nil[error],
		},
		{
			name: "invalid page token",
			args: args{
				req: &orchestrator.ListAssessmentResultsRequest{
					PageSize:  2,
					PageToken: "invalid-token!!!",
				},
				values: []int{1, 2, 3, 4, 5},
				opts:   PaginationOpts{10, 10},
			},
			wantPage: func(t *testing.T, got []int) bool {
				return assert.Equal(t, []int(nil), got)
			},
			wantNbt: assert.Empty[string],
			wantErr: assert.AnyValue[error],
		},
		{
			name: "page token offset beyond slice length",
			args: args{
				req: &orchestrator.ListAssessmentResultsRequest{
					PageSize:  2,
					PageToken: "CBoQAg==", // Start=26, Size=2 (beyond slice of 5 elements)
				},
				values: []int{1, 2, 3, 4, 5},
				opts:   PaginationOpts{10, 10},
			},
			wantPage: func(t *testing.T, got []int) bool {
				return assert.Equal(t, []int{}, got)
			},
			wantNbt: func(t *testing.T, got string) bool {
				return assert.Equal(t, "", got)
			},
			wantErr: assert.Nil[error],
		},
		{
			name: "zero page size uses default",
			args: args{
				req: &orchestrator.ListAssessmentResultsRequest{
					PageSize:  0,
					PageToken: "",
				},
				values: []int{1, 2, 3, 4, 5},
				opts:   PaginationOpts{DefaultPageSize: 3, MaxPageSize: 10},
			},
			wantPage: func(t *testing.T, got []int) bool {
				// Should use DefaultPageSize (3)
				return assert.Equal(t, []int{1, 2, 3}, got)
			},
			wantNbt: func(t *testing.T, got string) bool {
				return assert.Equal(t, "CAMQAw==", got) // Start=3, Size=3
			},
			wantErr: assert.Nil[error],
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPage, gotNbt, err := PaginateSlice(tt.args.req, tt.args.values, func(a int, b int) bool { return a < b }, tt.args.opts)

			tt.wantErr(t, err)
			tt.wantNbt(t, gotNbt)
			tt.wantPage(t, gotPage)
		})
	}
}

func TestPaginateStorage(t *testing.T) {
	type args struct {
		req   api.PaginatedRequest
		db    *persistence.DB
		opts  PaginationOpts
		conds []interface{}
	}
	tests := []struct {
		name     string
		args     args
		wantPage assert.Want[[]orchestrator.TargetOfEvaluation]
		wantNbt  assert.Want[string]
		wantErr  assert.WantErr
	}{
		{
			name: "first page",
			args: args{
				req: &orchestrator.ListTargetsOfEvaluationRequest{
					PageSize:  2,
					PageToken: "",
				},
				db: persistencetest.NewInMemoryDB(t, []any{orchestrator.TargetOfEvaluation{}}, nil, func(db *persistence.DB) {
					assert.NoError(t, db.Create(&orchestrator.TargetOfEvaluation{Id: "1"}))
					assert.NoError(t, db.Create(&orchestrator.TargetOfEvaluation{Id: "2"}))
					assert.NoError(t, db.Create(&orchestrator.TargetOfEvaluation{Id: "3"}))
					assert.NoError(t, db.Create(&orchestrator.TargetOfEvaluation{Id: "4"}))
					assert.NoError(t, db.Create(&orchestrator.TargetOfEvaluation{Id: "5"}))
				}),
				opts: PaginationOpts{10, 10},
			},
			wantPage: func(t *testing.T, got []orchestrator.TargetOfEvaluation) bool {
				want := []orchestrator.TargetOfEvaluation{
					{Id: "1", ConfiguredMetrics: []*assessment.Metric{}},
					{Id: "2", ConfiguredMetrics: []*assessment.Metric{}},
				}
				return assert.Equal(t, want, got)
			},
			wantNbt: func(t *testing.T, got string) bool {
				return assert.Equal(t, "CAIQAg==", got)
			},
			wantErr: assert.Nil[error],
		},
		{
			name: "next page",
			args: args{
				req: &orchestrator.ListTargetsOfEvaluationRequest{
					PageSize:  2,
					PageToken: "CAIQAg==",
				},
				db: persistencetest.NewInMemoryDB(t, []any{orchestrator.TargetOfEvaluation{}}, nil, func(db *persistence.DB) {
					assert.NoError(t, db.Create(&orchestrator.TargetOfEvaluation{Id: "1"}))
					assert.NoError(t, db.Create(&orchestrator.TargetOfEvaluation{Id: "2"}))
					assert.NoError(t, db.Create(&orchestrator.TargetOfEvaluation{Id: "3"}))
					assert.NoError(t, db.Create(&orchestrator.TargetOfEvaluation{Id: "4"}))
					assert.NoError(t, db.Create(&orchestrator.TargetOfEvaluation{Id: "5"}))
				}),
				opts: PaginationOpts{10, 10},
			},
			wantPage: func(t *testing.T, got []orchestrator.TargetOfEvaluation) bool {
				want := []orchestrator.TargetOfEvaluation{
					{Id: "3", ConfiguredMetrics: []*assessment.Metric{}},
					{Id: "4", ConfiguredMetrics: []*assessment.Metric{}},
				}
				return assert.Equal(t, want, got)
			},
			wantNbt: func(t *testing.T, got string) bool {
				return assert.Equal(t, "CAQQAg==", got)
			},
			wantErr: assert.Nil[error],
		},
		{
			name: "last page",
			args: args{
				req: &orchestrator.ListTargetsOfEvaluationRequest{
					PageSize:  2,
					PageToken: "CAQQAg==",
				},
				db: persistencetest.NewInMemoryDB(t, []any{orchestrator.TargetOfEvaluation{}}, nil, func(db *persistence.DB) {
					assert.NoError(t, db.Create(&orchestrator.TargetOfEvaluation{Id: "1"}))
					assert.NoError(t, db.Create(&orchestrator.TargetOfEvaluation{Id: "2"}))
					assert.NoError(t, db.Create(&orchestrator.TargetOfEvaluation{Id: "3"}))
					assert.NoError(t, db.Create(&orchestrator.TargetOfEvaluation{Id: "4"}))
					assert.NoError(t, db.Create(&orchestrator.TargetOfEvaluation{Id: "5"}))
				}),
				opts: PaginationOpts{10, 10},
			},
			wantPage: func(t *testing.T, got []orchestrator.TargetOfEvaluation) bool {
				want := []orchestrator.TargetOfEvaluation{{Id: "5", ConfiguredMetrics: []*assessment.Metric{}}}

				return assert.Equal(t, want, got)
			},
			wantNbt: assert.Empty[string],
			wantErr: assert.Nil[error],
		},
		{
			name: "empty database",
			args: args{
				req: &orchestrator.ListTargetsOfEvaluationRequest{
					PageSize:  2,
					PageToken: "",
				},
				db:   persistencetest.NewInMemoryDB(t, []any{orchestrator.TargetOfEvaluation{}}, nil),
				opts: PaginationOpts{10, 10},
			},
			wantPage: func(t *testing.T, got []orchestrator.TargetOfEvaluation) bool {
				return assert.Equal(t, []orchestrator.TargetOfEvaluation{}, got)
			},
			wantNbt: assert.Empty[string],
			wantErr: assert.Nil[error],
		},
		{
			name: "invalid page token",
			args: args{
				req: &orchestrator.ListTargetsOfEvaluationRequest{
					PageSize:  2,
					PageToken: "invalid-token!!!",
				},
				db:   persistencetest.NewInMemoryDB(t, []any{orchestrator.TargetOfEvaluation{}}, nil),
				opts: PaginationOpts{10, 10},
			},
			wantPage: func(t *testing.T, got []orchestrator.TargetOfEvaluation) bool {
				return assert.Equal(t, []orchestrator.TargetOfEvaluation(nil), got)
			},
			wantNbt: assert.Empty[string],
			wantErr: assert.AnyValue[error],
		},
		{
			name: "page token offset beyond available records",
			args: args{
				req: &orchestrator.ListTargetsOfEvaluationRequest{
					PageSize:  2,
					PageToken: "CBoQAg==", // Start=26, Size=2 (beyond 5 records)
				},
				db: persistencetest.NewInMemoryDB(t, []any{orchestrator.TargetOfEvaluation{}}, nil, func(db *persistence.DB) {
					assert.NoError(t, db.Create(&orchestrator.TargetOfEvaluation{Id: "1"}))
					assert.NoError(t, db.Create(&orchestrator.TargetOfEvaluation{Id: "2"}))
					assert.NoError(t, db.Create(&orchestrator.TargetOfEvaluation{Id: "3"}))
					assert.NoError(t, db.Create(&orchestrator.TargetOfEvaluation{Id: "4"}))
					assert.NoError(t, db.Create(&orchestrator.TargetOfEvaluation{Id: "5"}))
				}),
				opts: PaginationOpts{10, 10},
			},
			wantPage: func(t *testing.T, got []orchestrator.TargetOfEvaluation) bool {
				return assert.Equal(t, []orchestrator.TargetOfEvaluation{}, got)
			},
			wantNbt: assert.Empty[string],
			wantErr: assert.Nil[error],
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPage, gotNbt, err := PaginateStorage[orchestrator.TargetOfEvaluation](tt.args.req, tt.args.db,
				tt.args.opts, tt.args.conds...)

			tt.wantErr(t, err)
			tt.wantNbt(t, gotNbt)
			tt.wantPage(t, gotPage)
		})
	}
}
