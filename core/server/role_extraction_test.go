// Copyright 2016-2026 Fraunhofer AISEC
//
// SPDX-License-Identifier: Apache-2.0
//
// This file is part of Confirmate Core.

package server

import (
	"testing"

	"confirmate.io/core/api/orchestrator"
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
	// applyRoleMapping always runs through normalizeRole — exercise it through
	// the public constructor so the test reflects the wiring callers actually
	// get.
	tests := []struct {
		name  string
		paths []string
		raw   jwt.MapClaims
		// preset is the value of claims.Roles before applyRoleMapping runs.
		preset []orchestrator.Role
		want   []orchestrator.Role
	}{
		{
			name:   "default path reads top-level roles",
			paths:  nil, // i.e. no explicit override; constructor default applies
			raw:    jwt.MapClaims{"roles": []any{"ROLE_LEAD_AUDITOR"}},
			preset: []orchestrator.Role{orchestrator.Role_ROLE_COMPLIANCE_MANAGER},
			want:   []orchestrator.Role{orchestrator.Role_ROLE_LEAD_AUDITOR},
		},
		{
			name:  "explicit path replaces preset",
			paths: []string{"realm_access.roles"},
			raw: jwt.MapClaims{
				"realm_access": map[string]any{"roles": []any{"ROLE_ADMIN", "ROLE_LEAD_AUDITOR"}},
			},
			preset: []orchestrator.Role{orchestrator.Role_ROLE_COMPLIANCE_MANAGER},
			want:   []orchestrator.Role{orchestrator.Role_ROLE_ADMIN, orchestrator.Role_ROLE_LEAD_AUDITOR},
		},
		{
			name:  "multiple paths are merged and deduplicated",
			paths: []string{"roles", "realm_access.roles"},
			raw: jwt.MapClaims{
				"roles":        []any{"ROLE_ADMIN"},
				"realm_access": map[string]any{"roles": []any{"ROLE_ADMIN", "ROLE_LEAD_AUDITOR"}},
			},
			want: []orchestrator.Role{orchestrator.Role_ROLE_ADMIN, orchestrator.Role_ROLE_LEAD_AUDITOR},
		},
		{
			name:  "normalizer translates known IdP names",
			paths: []string{"realm_access.roles"},
			raw: jwt.MapClaims{
				"realm_access": map[string]any{"roles": []any{"ORCHESTRATOR_ADMIN", "Compliance Manager"}},
			},
			want: []orchestrator.Role{orchestrator.Role_ROLE_ADMIN, orchestrator.Role_ROLE_COMPLIANCE_MANAGER},
		},
		{
			name:  "unknown role strings are dropped",
			paths: []string{"realm_access.roles"},
			raw: jwt.MapClaims{
				"realm_access": map[string]any{"roles": []any{"ROLE_ADMIN", "SOME_OTHER_ROLE"}},
			},
			want: []orchestrator.Role{orchestrator.Role_ROLE_ADMIN},
		},
		{
			name:  "all paths empty leaves preset alone",
			paths: []string{"realm_access.roles"},
			raw: jwt.MapClaims{
				"realm_access": map[string]any{"roles": []any{}},
			},
			preset: []orchestrator.Role{orchestrator.Role_ROLE_LEAD_AUDITOR},
			want:   []orchestrator.Role{orchestrator.Role_ROLE_LEAD_AUDITOR},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var opts []AuthOption
			if tt.paths != nil {
				opts = append(opts, WithRoleClaimPaths(tt.paths...))
			}
			ai := NewAuthInterceptor(opts...)
			claims := &auth.OAuthClaims{Roles: tt.preset}
			ai.applyRoleMapping(claims, tt.raw)
			assert.Equal(t, tt.want, claims.Roles)
		})
	}
}
