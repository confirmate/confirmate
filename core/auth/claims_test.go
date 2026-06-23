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

	"confirmate.io/core/api/orchestrator"
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

func TestOAuthClaimsHasRole(t *testing.T) {
	tests := []struct {
		name   string
		claims *OAuthClaims
		role   orchestrator.Role
		want   assert.Want[bool]
	}{
		{
			name:   "nil claims",
			claims: nil,
			role:   orchestrator.Role_ROLE_ADMIN,
			want: func(t *testing.T, got bool, _ ...any) bool {
				return assert.False(t, got)
			},
		},
		{
			name: "unspecified role",
			claims: &OAuthClaims{
				Roles: []orchestrator.Role{orchestrator.Role_ROLE_ADMIN},
			},
			role: orchestrator.Role_ROLE_UNSPECIFIED,
			want: func(t *testing.T, got bool, _ ...any) bool {
				return assert.False(t, got)
			},
		},
		{
			name: "role present",
			claims: &OAuthClaims{
				Roles: []orchestrator.Role{orchestrator.Role_ROLE_LEAD_AUDITOR, orchestrator.Role_ROLE_ADMIN},
			},
			role: orchestrator.Role_ROLE_ADMIN,
			want: func(t *testing.T, got bool, _ ...any) bool {
				return assert.True(t, got)
			},
		},
		{
			name: "role missing",
			claims: &OAuthClaims{
				Roles: []orchestrator.Role{orchestrator.Role_ROLE_LEAD_AUDITOR},
			},
			role: orchestrator.Role_ROLE_ADMIN,
			want: func(t *testing.T, got bool, _ ...any) bool {
				return assert.False(t, got)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.claims.HasRole(tt.role)
			tt.want(t, got)
		})
	}
}

func TestOAuthClaimsIsAdmin(t *testing.T) {
	tests := []struct {
		name   string
		claims *OAuthClaims
		want   assert.Want[bool]
	}{
		{
			name:   "nil claims",
			claims: nil,
			want: func(t *testing.T, got bool, _ ...any) bool {
				return assert.False(t, got)
			},
		},
		{
			name: "cfadmin claim is admin",
			claims: &OAuthClaims{
				IsAdminToken: true,
				Roles:        []orchestrator.Role{orchestrator.Role_ROLE_LEAD_AUDITOR},
			},
			want: func(t *testing.T, got bool, _ ...any) bool {
				return assert.True(t, got)
			},
		},
		{
			name: "role admin is admin",
			claims: &OAuthClaims{
				Roles: []orchestrator.Role{orchestrator.Role_ROLE_LEAD_AUDITOR, orchestrator.Role_ROLE_ADMIN},
			},
			want: func(t *testing.T, got bool, _ ...any) bool {
				return assert.True(t, got)
			},
		},
		{
			name: "without admin claim or role is not admin",
			claims: &OAuthClaims{
				Roles: []orchestrator.Role{orchestrator.Role_ROLE_LEAD_AUDITOR},
			},
			want: func(t *testing.T, got bool, _ ...any) bool {
				return assert.False(t, got)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.claims.IsAdmin()
			tt.want(t, got)
		})
	}
}
