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
)

func TestWithEmbeddedOAuth2Server_AuthorizeRoute(t *testing.T) {
	var (
		srv    *Server
		err    error
		res    *http.Response
		client *http.Client
		url    string
	)

	srv, err = NewConnectServer([]Option{
		WithEmbeddedOAuth2Server(DefaultOAuth2KeyPath, DefaultOAuth2KeyPassword, false, ""),
	})
	if err != nil {
		t.Fatalf("NewConnectServer() failed: %v", err)
	}

	ts := httptest.NewServer(srv.Handler)
	defer ts.Close()

	client = &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	url = ts.URL + "/v1/auth/authorize?client_id=cli&code_challenge=abc&code_challenge_method=S256&redirect_uri=http%3A%2F%2Flocalhost%3A10000%2Fcallback&response_type=code&state=test"
	res, err = client.Get(url)
	if err != nil {
		t.Fatalf("client.Get() failed: %v", err)
	}
	defer func() {
		_ = res.Body.Close()
	}()
	_, _ = io.Copy(io.Discard, res.Body)

	if res.StatusCode != http.StatusFound {
		t.Fatalf("/v1/auth/authorize status = %d, want %d", res.StatusCode, http.StatusFound)
	}

	if !strings.HasPrefix(res.Header.Get("Location"), "/v1/auth/login") {
		t.Fatalf("/v1/auth/authorize redirected to %q, want /v1/auth/login...", res.Header.Get("Location"))
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
			got := normalizeOAuthPublicURL(tt.publicURL, tt.port)
			if got != tt.want {
				t.Fatalf("normalizeOAuthPublicURL() = %q, want %q", got, tt.want)
			}
		})
	}
}
