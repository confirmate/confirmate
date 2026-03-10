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

package orchestratorconnect

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/service/orchestrator/orchestratortest"
	"confirmate.io/core/util/assert"

	"connectrpc.com/connect"
)

type mockOrchestratorHandler struct {
	UnimplementedOrchestratorHandler
}

func (mockOrchestratorHandler) RegisterAssessmentTool(
	ctx context.Context,
	req *connect.Request[orchestrator.RegisterAssessmentToolRequest],
) (res *connect.Response[orchestrator.AssessmentTool], err error) {
	res = connect.NewResponse(orchestratortest.MockAssessmentTool1)
	return res, nil
}

func TestNewOrchestratorCompatHandler(t *testing.T) {
	type args struct {
		procedure string
	}

	tests := []struct {
		name    string
		args    args
		want    assert.Want[*connect.Response[orchestrator.AssessmentTool]]
		wantErr assert.WantErr
	}{
		{
			name: "canonical service path",
			args: args{
				procedure: OrchestratorRegisterAssessmentToolProcedure,
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.AssessmentTool], msgAndArgs ...any) bool {
				assert.NotNil(t, got)
				assert.NotNil(t, got.Msg)
				return assert.Equal(t, orchestratortest.MockAssessmentTool1, got.Msg)
			},
			wantErr: assert.NoError,
		},
		{
			name: "clouditor compatibility service path",
			args: args{
				procedure: "/" + OrchestratorCompatName + "/RegisterAssessmentTool",
			},
			want: func(t *testing.T, got *connect.Response[orchestrator.AssessmentTool], msgAndArgs ...any) bool {
				assert.NotNil(t, got)
				assert.NotNil(t, got.Msg)
				return assert.Equal(t, orchestratortest.MockAssessmentTool1, got.Msg)
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				mux           *http.ServeMux
				server        *httptest.Server
				canonicalPath string
				compatPath    string
				handler       mockOrchestratorHandler
				canonical     http.Handler
				compat        http.Handler
				client        *connect.Client[orchestrator.RegisterAssessmentToolRequest, orchestrator.AssessmentTool]
				res           *connect.Response[orchestrator.AssessmentTool]
				err           error
			)

			canonicalPath, canonical = NewOrchestratorHandler(handler)
			compatPath, compat = NewOrchestratorCompatHandler(handler)

			mux = http.NewServeMux()
			mux.Handle(canonicalPath, canonical)
			mux.Handle(compatPath, compat)

			server = httptest.NewServer(mux)
			defer server.Close()

			client = connect.NewClient[orchestrator.RegisterAssessmentToolRequest, orchestrator.AssessmentTool](
				http.DefaultClient,
				server.URL+tt.args.procedure,
			)

			res, err = client.CallUnary(
				context.Background(),
				connect.NewRequest(&orchestrator.RegisterAssessmentToolRequest{Tool: orchestratortest.MockAssessmentTool1}),
			)

			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}
