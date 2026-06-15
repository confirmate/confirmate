// Copyright 2016-2026 Fraunhofer AISEC
//
// SPDX-License-Identifier: Apache-2.0
//
// This file is part of Confirmate Core.

package orchestrator

import (
	"context"
	"testing"

	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/auth"
	"confirmate.io/core/persistence"
	"confirmate.io/core/persistence/persistencetest"
	"confirmate.io/core/util/assert"

	"github.com/golang-jwt/jwt/v5"
)

func TestRolesFromClaims(t *testing.T) {
	tests := []struct {
		name   string
		claims *auth.OAuthClaims
		want   []orchestrator.Role
	}{
		{name: "nil claims", claims: nil, want: nil},
		{name: "no roles", claims: &auth.OAuthClaims{}, want: nil},
		{
			name:   "single known role",
			claims: &auth.OAuthClaims{Roles: []string{"ROLE_ADMIN"}},
			want:   []orchestrator.Role{orchestrator.Role_ROLE_ADMIN},
		},
		{
			name: "multiple known roles preserve order",
			claims: &auth.OAuthClaims{Roles: []string{
				"ROLE_AUDITOR", "ROLE_COMPLIANCE_MANAGER",
			}},
			want: []orchestrator.Role{
				orchestrator.Role_ROLE_AUDITOR,
				orchestrator.Role_ROLE_COMPLIANCE_MANAGER,
			},
		},
		{
			name:   "duplicate roles are collapsed",
			claims: &auth.OAuthClaims{Roles: []string{"ROLE_ADMIN", "ROLE_ADMIN"}},
			want:   []orchestrator.Role{orchestrator.Role_ROLE_ADMIN},
		},
		{
			name:   "unknown role strings are dropped",
			claims: &auth.OAuthClaims{Roles: []string{"AMOE_ADMIN", "ROLE_ADMIN"}},
			want:   []orchestrator.Role{orchestrator.Role_ROLE_ADMIN},
		},
		{
			name:   "ROLE_UNSPECIFIED is dropped",
			claims: &auth.OAuthClaims{Roles: []string{"ROLE_UNSPECIFIED", "ROLE_ADMIN"}},
			want:   []orchestrator.Role{orchestrator.Role_ROLE_ADMIN},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, rolesFromClaims(tt.claims))
		})
	}
}

// TestProvisionCurrentUser_PopulatesRoles verifies that JIT user provisioning
// translates claims.Roles into typed Role enum values on the persisted User
// record — the wire-up for issue #288.
func TestProvisionCurrentUser_PopulatesRoles(t *testing.T) {
	db := persistencetest.NewInMemoryDB(t, types, joinTables)
	svc := &Service{db: db}

	ctx := auth.WithClaims(context.Background(), &auth.OAuthClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "alice",
			Issuer:  "https://idp.example",
		},
		Roles: []string{"ROLE_ADMIN", "AMOE_ADMIN", "ROLE_AUDITOR"},
	})

	userId, err := provisionCurrentUser(ctx, svc)
	assert.NoError(t, err)
	assert.Equal(t, "https://idp.example|alice", userId)

	var got orchestrator.User
	assert.NoError(t, db.Get(&got, "id = ?", userId))
	assert.Equal(t, []orchestrator.Role{
		orchestrator.Role_ROLE_ADMIN,
		orchestrator.Role_ROLE_AUDITOR,
	}, got.Roles)
}

// TestProvisionCurrentUser_UpdatesRolesOnReprovisioning ensures that changes to
// the JWT's roles claim between requests are reflected on the persisted user
// record (e.g. an IdP role revocation propagates).
func TestProvisionCurrentUser_UpdatesRolesOnReprovisioning(t *testing.T) {
	db := persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
		assert.NoError(t, d.Create(&orchestrator.User{
			Id:      "https://idp.example|bob",
			Enabled: true,
			Roles:   []orchestrator.Role{orchestrator.Role_ROLE_ADMIN},
		}))
	})
	svc := &Service{db: db}

	ctx := auth.WithClaims(context.Background(), &auth.OAuthClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "bob",
			Issuer:  "https://idp.example",
		},
		Roles: []string{"ROLE_AUDITOR"},
	})

	_, err := provisionCurrentUser(ctx, svc)
	assert.NoError(t, err)

	var got orchestrator.User
	assert.NoError(t, db.Get(&got, "id = ?", "https://idp.example|bob"))
	assert.Equal(t, []orchestrator.Role{orchestrator.Role_ROLE_AUDITOR}, got.Roles)
}
