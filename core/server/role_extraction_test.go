// Copyright 2016-2026 Fraunhofer AISEC
//
// SPDX-License-Identifier: Apache-2.0
//
// This file is part of Confirmate Core.

package server

import (
	"testing"

	"confirmate.io/core/auth"
	"confirmate.io/core/util/assert"

	"github.com/golang-jwt/jwt/v5"
)

func TestExtractStringListAtPath(t *testing.T) {
	tests := []struct {
		name string
		raw  jwt.MapClaims
		path string
		want []string
	}{
		{
			name: "empty map and path returns nil",
			raw:  jwt.MapClaims{},
			path: "",
			want: nil,
		},
		{
			name: "top-level []any list",
			raw:  jwt.MapClaims{"roles": []any{"a", "b"}},
			path: "roles",
			want: []string{"a", "b"},
		},
		{
			name: "top-level []string list",
			raw:  jwt.MapClaims{"roles": []string{"a", "b"}},
			path: "roles",
			want: []string{"a", "b"},
		},
		{
			name: "nested Keycloak-style path",
			raw:  jwt.MapClaims{"realm_access": map[string]any{"roles": []any{"ORCHESTRATOR_ADMIN", "Compliance Manager"}}},
			path: "realm_access.roles",
			want: []string{"ORCHESTRATOR_ADMIN", "Compliance Manager"},
		},
		{
			name: "comma-separated string leaf",
			raw:  jwt.MapClaims{"roles": "a,b, c"},
			path: "roles",
			want: []string{"a", "b", "c"},
		},
		{
			name: "space-separated string leaf",
			raw:  jwt.MapClaims{"roles": "a b c"},
			path: "roles",
			want: []string{"a", "b", "c"},
		},
		{
			name: "missing intermediate segment returns nil",
			raw:  jwt.MapClaims{"foo": map[string]any{}},
			path: "foo.bar.baz",
			want: nil,
		},
		{
			name: "non-string element types are dropped",
			raw:  jwt.MapClaims{"roles": []any{"a", 42, "b"}},
			path: "roles",
			want: []string{"a", "b"},
		},
		{
			name: "unsupported leaf type returns nil",
			raw:  jwt.MapClaims{"roles": 42},
			path: "roles",
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, extractStringListAtPath(tt.raw, tt.path))
		})
	}
}

func TestApplyRoleMapping(t *testing.T) {
	tests := []struct {
		name   string
		paths  []string
		mapper RoleMapper
		raw    jwt.MapClaims
		// preset is the value of claims.Roles before applyRoleMapping runs (as
		// already populated by JSON unmarshal from the structured "roles" field).
		preset []string
		want   []string
	}{
		{
			name:   "no paths configured leaves preset alone",
			paths:  nil,
			raw:    jwt.MapClaims{"roles": []any{"X"}},
			preset: []string{"X"},
			want:   []string{"X"},
		},
		{
			name:  "single path replaces preset",
			paths: []string{"realm_access.roles"},
			raw: jwt.MapClaims{
				"realm_access": map[string]any{"roles": []any{"ROLE_ADMIN", "ROLE_AUDITOR"}},
			},
			preset: []string{"ignored"},
			want:   []string{"ROLE_ADMIN", "ROLE_AUDITOR"},
		},
		{
			name:  "multiple paths are merged and deduplicated",
			paths: []string{"roles", "realm_access.roles"},
			raw: jwt.MapClaims{
				"roles":        []any{"ROLE_ADMIN"},
				"realm_access": map[string]any{"roles": []any{"ROLE_ADMIN", "ROLE_AUDITOR"}},
			},
			want: []string{"ROLE_ADMIN", "ROLE_AUDITOR"},
		},
		{
			name:   "mapper translates and drops empties",
			paths:  []string{"realm_access.roles"},
			mapper: KeycloakRoleMapper,
			raw: jwt.MapClaims{
				"realm_access": map[string]any{"roles": []any{"ORCHESTRATOR_ADMIN", "Compliance Manager", "AMOE_ADMIN"}},
			},
			want: []string{"ROLE_ADMIN", "ROLE_COMPLIANCE_MANAGER", "AMOE_ADMIN"},
		},
		{
			name:  "all paths empty leaves preset alone",
			paths: []string{"realm_access.roles"},
			raw: jwt.MapClaims{
				"realm_access": map[string]any{"roles": []any{}},
			},
			preset: []string{"preset"},
			want:   []string{"preset"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ai := &AuthInterceptor{cfg: &AuthConfig{
				roleClaimPaths: tt.paths,
				roleMapper:     tt.mapper,
			}}
			claims := &auth.OAuthClaims{Raw: tt.raw, Roles: tt.preset}
			ai.applyRoleMapping(claims)
			assert.Equal(t, tt.want, claims.Roles)
		})
	}
}
