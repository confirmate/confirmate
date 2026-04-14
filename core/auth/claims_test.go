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

package auth

import (
	"testing"

	"confirmate.io/core/util/assert"
	"github.com/golang-jwt/jwt/v5"
)

func TestGetConfirmateUserIDFromClaims(t *testing.T) {
	tests := []struct {
		name   string // description of this test case
		claims *OAuthClaims
		want   assert.Want[string]
	}{
		{
			name: "err: claims is nil",
			want: func(t *testing.T, got string, _ ...any) bool {
				return assert.Equal(t, "", got)
			},
		},
		{
			name: "happy path",
			claims: &OAuthClaims{
				RegisteredClaims: jwt.RegisteredClaims{
					Issuer:  "testIssuer",
					Subject: "testSubject",
				},
			},
			want: func(t *testing.T, got string, _ ...any) bool {
				expected := "testIssuer|testSubject"

				return assert.Equal(t, expected, got)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetConfirmateUserIDFromClaims(tt.claims)

			tt.want(t, got)
		})
	}
}
