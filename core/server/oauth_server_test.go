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
	"testing"
)

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
