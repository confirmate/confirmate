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

package service_test

import (
	"io"
	"testing"

	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/persistence"
	"confirmate.io/core/service"
	"confirmate.io/core/util/assert"
	"connectrpc.com/connect"
)

func TestHandleDatabaseError(t *testing.T) {
	type args struct {
		err          error
		notFoundErrs []error
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.WantErr
	}{
		{
			name: "happy path",
			args: args{
				err:          nil,
				notFoundErrs: []error{},
			},
			wantErr: assert.NoError,
		},
		{
			name: "not found error",
			args: args{
				err:          persistence.ErrRecordNotFound,
				notFoundErrs: []error{service.ErrNotFound("entity")},
			},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				cErr := assert.Is[*connect.Error](t, err)
				return assert.Equal(t, connect.CodeNotFound, cErr.Code())
			},
		},
		{
			name: "other error",
			args: args{
				err:          io.EOF,
				notFoundErrs: []error{},
			},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.ErrorIs(t, err, io.EOF)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := service.HandleDatabaseError(tt.args.err, tt.args.notFoundErrs...)
			tt.wantErr(t, gotErr)
		})
	}
}

func TestValidate(t *testing.T) {
	type args struct {
		req *connect.Request[orchestrator.CreateMetricRequest]
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.WantErr
	}{
		{
			name: "happy path",
			args: args{
				req: connect.NewRequest(&orchestrator.CreateMetricRequest{
					Metric: &assessment.Metric{
						Id:          "metric-1",
						Description: "Test Metric",
						Version:     "1.0.0",
						Category:    "awesome",
					},
				}),
			},
			wantErr: assert.NoError,
		},
		{
			name: "nil request message",
			args: args{
				req: connect.NewRequest[orchestrator.CreateMetricRequest](nil),
			},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				cErr := assert.Is[*connect.Error](t, err)
				return assert.Equal(t, connect.CodeInvalidArgument, cErr.Code())
			},
		},
		{
			name: "invalid request",
			args: args{
				req: connect.NewRequest(&orchestrator.CreateMetricRequest{
					Metric: &assessment.Metric{
						Id: "", // Missing required field
					},
				}),
			},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				cErr := assert.Is[*connect.Error](t, err)
				return assert.Equal(t, connect.CodeInvalidArgument, cErr.Code())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := service.Validate(tt.args.req)
			tt.wantErr(t, gotErr)
		})
	}
}
