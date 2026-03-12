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

package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"confirmate.io/core/util/assert"
)

func TestWithEmbeddedOAuth2Server_AuthorizeRoute(t *testing.T) {
	type args struct {
		path string
	}

	var (
		tests = []struct {
			name               string
			args               args
			wantStatusCode     int
			wantLocationPrefix string
		}{
			{
				name: "authorize endpoint redirects to embedded login route",
				args: args{
					path: "/v1/auth/authorize?client_id=cli&code_challenge=abc&code_challenge_method=S256&redirect_uri=http%3A%2F%2Flocalhost%3A10000%2Fcallback&response_type=code&state=test",
				},
				wantStatusCode:     http.StatusFound,
				wantLocationPrefix: "/v1/auth/login",
			},
		}
		srv    *Server
		err    error
		client *http.Client
	)

	srv, err = NewConnectServer([]Option{
		WithEmbeddedOAuth2Server(DefaultOAuth2KeyPath, DefaultOAuth2KeyPassword, false, ""),
	})
	if !assert.NoError(t, err) {
		return
	}

	ts := httptest.NewServer(srv.Handler)
	defer ts.Close()

	client = &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				res *http.Response
				url string
			)

			url = ts.URL + tt.args.path
			res, err = client.Get(url)
			if !assert.NoError(t, err) {
				return
			}
			defer func() {
				_ = res.Body.Close()
			}()
			_, _ = io.Copy(io.Discard, res.Body)

			assert.Equal(t, tt.wantStatusCode, res.StatusCode)
			assert.True(t, strings.HasPrefix(res.Header.Get("Location"), tt.wantLocationPrefix))
		})
	}
}

func TestNormalizeOAuthPublicURL(t *testing.T) {
	tests := []struct {
		name      string
		publicURL string
		port      uint16
		want      string
	}{
		{
			name:      "empty url uses localhost default",
			publicURL: "",
			port:      8081,
			want:      "http://localhost:8081/v1/auth",
		},
		{
			name:      "adds auth suffix",
			publicURL: "https://example.test",
			port:      0,
			want:      "https://example.test/v1/auth",
		},
		{
			name:      "keeps existing auth suffix",
			publicURL: "https://example.test/v1/auth/",
			port:      0,
			want:      "https://example.test/v1/auth",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string

			got = normalizeOAuthPublicURL(tt.publicURL, tt.port)
			assert.Equal(t, tt.want, got)
		})
	}
}
