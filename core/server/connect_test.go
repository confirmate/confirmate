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
	"net/http"
	"net/http/httptest"
	"testing"

	"confirmate.io/core/util/assert"
)

func TestNewConnectServer_WithHTTPHandler(t *testing.T) {
	tests := []struct {
		name         string
		pattern      string
		requestPath  string
		wantHTTPCode int
	}{
		{
			name:         "Reflection-like route is served directly",
			pattern:      "/grpc.reflection.v1.ServerReflection/",
			requestPath:  "/grpc.reflection.v1.ServerReflection/ServerReflectionInfo",
			wantHTTPCode: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, err := NewConnectServer([]Option{
				WithHTTPHandler(tt.pattern, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusNoContent)
				})),
			})
			assert.NoError(t, err)
			assert.NotNil(t, srv)
			if err != nil || srv == nil {
				return
			}

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, tt.requestPath, nil)
			srv.Handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantHTTPCode, rec.Code)
		})
	}
}
