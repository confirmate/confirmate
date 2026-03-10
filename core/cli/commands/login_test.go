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

package commands

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"confirmate.io/core/util/assert"

	"golang.org/x/oauth2"
)

func TestCallbackServer_authorizationURL(t *testing.T) {
	type fields struct {
		srv *callbackServer
	}

	tests := []struct {
		name    string
		fields  fields
		want    assert.Want[string]
		wantErr assert.WantErr
	}{
		{
			name: "includes generated state and PKCE parameters",
			fields: fields{srv: &callbackServer{
				verifier: "verifier-1",
				state:    "state-123",
				config: &oauth2.Config{Endpoint: oauth2.Endpoint{
					AuthURL: "http://localhost:8080/v1/auth/authorize",
				}},
			}},
			want: func(t *testing.T, got string, _ ...any) bool {
				parsed, err := url.Parse(got)
				if !assert.NoError(t, err) {
					return false
				}

				q := parsed.Query()
				return assert.Equal(t, "state-123", q.Get("state")) &&
					assert.Equal(t, "S256", q.Get("code_challenge_method")) &&
					assert.NotEmpty(t, q.Get("code_challenge"))
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fields.srv.authorizationURL()
			assert.True(t, tt.wantErr(t, nil))
			assert.True(t, tt.want(t, got))
		})
	}
}

func TestCallbackServer_handleCallback(t *testing.T) {
	type args struct {
		query string
	}
	type fields struct {
		state string
	}
	type want struct {
		status int
		code   string
		err    assert.Want[error]
	}

	tests := []struct {
		name    string
		args    args
		fields  fields
		want    want
		wantErr assert.WantErr
	}{
		{
			name:   "accepts valid state and code",
			args:   args{query: "code=auth-code-1&state=state-123"},
			fields: fields{state: "state-123"},
			want: want{
				status: http.StatusOK,
				code:   "auth-code-1",
				err:    assert.Nil[error],
			},
			wantErr: assert.NoError,
		},
		{
			name:   "rejects mismatched state",
			args:   args{query: "code=auth-code-1&state=unexpected"},
			fields: fields{state: "state-123"},
			want: want{
				status: http.StatusBadRequest,
				err: func(t *testing.T, got error, _ ...any) bool {
					return assert.ErrorIs(t, got, ErrInvalidOAuthState)
				},
			},
			wantErr: assert.NoError,
		},
		{
			name:   "rejects missing code",
			args:   args{query: "state=state-123"},
			fields: fields{state: "state-123"},
			want: want{
				status: http.StatusBadRequest,
				err: func(t *testing.T, got error, _ ...any) bool {
					return assert.ErrorIs(t, got, ErrMissingOAuthCode)
				},
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &callbackServer{
				state:  tt.fields.state,
				result: make(chan callbackResult, 1),
			}
			req := httptest.NewRequest(http.MethodGet, "/callback?"+tt.args.query, nil)
			rec := httptest.NewRecorder()

			srv.handleCallback(rec, req)
			result := <-srv.result

			assert.True(t, tt.wantErr(t, nil))
			assert.Equal(t, tt.want.status, rec.Code)
			assert.Equal(t, tt.want.code, result.code)
			assert.True(t, tt.want.err(t, result.err))
		})
	}
}
