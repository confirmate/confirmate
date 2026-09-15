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
	"confirmate.io/core/service/orchestrator/orchestratortest"
	"confirmate.io/core/util/assert"

	"github.com/golang-jwt/jwt/v5"
)

// TestProvisionCurrentUser_PopulatesRoles verifies that JIT user provisioning
// carries the Role enum values already extracted on claims.Roles by the auth
// interceptor across to the persisted User record — the wire-up for #288.
func TestProvisionCurrentUser_PopulatesRoles(t *testing.T) {
	db := persistencetest.NewInMemoryDB(t, types, joinTables)
	svc := &Service{db: db}

	ctx := auth.WithClaims(context.Background(), &auth.OAuthClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "alice",
			Issuer:  "https://idp.example",
		},
		Roles: []orchestrator.Role{
			orchestrator.Role_ROLE_ADMIN,
			orchestrator.Role_ROLE_LEAD_AUDITOR,
		},
	})

	userId, err := provisionCurrentUser(ctx, svc)
	assert.NoError(t, err)
	assert.Equal(t, orchestratortest.GetConfirmateUserID("https://idp.example", "alice"), userId)

	var got orchestrator.User
	assert.NoError(t, db.Get(&got, "id = ?", userId))
	assert.Equal(t, []orchestrator.Role{
		orchestrator.Role_ROLE_ADMIN,
		orchestrator.Role_ROLE_LEAD_AUDITOR,
	}, got.Roles)
}

// TestProvisionCurrentUser_UpdatesRolesOnReprovisioning ensures that changes to
// the JWT's roles claim between requests are reflected on the persisted user
// record (e.g. an IdP role revocation propagates).
func TestProvisionCurrentUser_UpdatesRolesOnReprovisioning(t *testing.T) {
	db := persistencetest.NewInMemoryDB(t, types, joinTables, func(d persistence.DB) {
		assert.NoError(t, d.Create(&orchestrator.User{
			Id:      orchestratortest.GetConfirmateUserID("https://idp.example", "bob"),
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
		Roles: []orchestrator.Role{orchestrator.Role_ROLE_LEAD_AUDITOR},
	})

	_, err := provisionCurrentUser(ctx, svc)
	assert.NoError(t, err)

	var got orchestrator.User
	assert.NoError(t, db.Get(&got, "id = ?", orchestratortest.GetConfirmateUserID("https://idp.example", "bob")))
	assert.Equal(t, []orchestrator.Role{orchestrator.Role_ROLE_LEAD_AUDITOR}, got.Roles)
}
