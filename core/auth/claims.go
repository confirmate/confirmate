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
	"confirmate.io/core/api/orchestrator"

	"github.com/golang-jwt/jwt/v5"
)

// OAuthClaims represents the expected claims in the JWT token for authentication. It extends the
// standard JWT claims with additional fields commonly used in OAuth2 and OpenID Connect tokens.
type OAuthClaims struct {
	jwt.RegisteredClaims
	Scope             string   `json:"scope,omitempty"`
	Email             string   `json:"email,omitempty"`
	PreferredUsername string   `json:"preferred_username,omitempty"`
	GivenName         string   `json:"given_name,omitempty"`
	FamilyName        string   `json:"family_name,omitempty"`
	Roles             []string `json:"roles,omitempty"`

	// IsAdminToken is a custom claim that indicates whether the token is an admin token. This can
	// be used to grant elevated permissions to tokens that are meant for administrative purposes.
	// The presence and value of this claim should be determined by the authentication provider
	// issuing the token.
	IsAdminToken bool `json:"cfadmin,omitempty"`

	// Raw holds the complete, unstructured set of JWT claims for flexible extraction.
	// It is not serialized/deserialized directly by jwt; we fill it during parsing.
	Raw jwt.MapClaims `json:"-"`
}

// ClaimOption mutates/normalizes claims after parsing.
type ClaimOption func(claims *OAuthClaims)

// Apply applies options to claims.
func (c *OAuthClaims) Apply(opts ...ClaimOption) {
	if c == nil {
		return
	}
	for _, opt := range opts {
		opt(c)
	}
}

// WithRoles ensures that all specified roles are present in the claims.
// Duplicate roles are ignored. Empty role strings are ignored.
func WithRoles(roles ...string) ClaimOption {
	return func(claims *OAuthClaims) {
		if claims == nil || len(roles) == 0 {
			return
		}

		for _, role := range roles {
			if role == "" {
				continue
			}
			// ensure role exists (idempotent)
			found := false
			for _, r := range claims.Roles {
				if r == role {
					found = true
					break
				}
			}
			if !found {
				claims.Roles = append(claims.Roles, role)
			}
		}
	}
}

// WithAdminToken sets the IsAdminToken claim to true, indicating that the token is an admin token.
func WithAdminToken() ClaimOption {
	return func(claims *OAuthClaims) {
		if claims == nil {
			return
		}
		claims.IsAdminToken = true
	}
}

// IsAdmin returns whether the claims indicate that the token is an admin token. It checks the
// IsAdminToken field as well as the presence of the ROLE_ADMIN in the roles claim.
func (claims *OAuthClaims) IsAdmin() bool {
	if claims == nil {
		return false
	}

	if claims.HasRole(orchestrator.RoleAdmin) || claims.IsAdminToken {
		return true
	}

	return false
}

// HasRole checks if the given role exists in the roles claim.
func (claims *OAuthClaims) HasRole(role string) (ok bool) {
	var candidate string

	if claims == nil || role == "" {
		return false
	}

	for _, candidate = range claims.Roles {
		if candidate == role {
			return true
		}
	}

	return false
}

// GetConfirmateUserIDFromClaims constructs a unique user ID from the claims. It combines the issuer and subject
func GetConfirmateUserIDFromClaims(claims *OAuthClaims) string {
	if claims == nil || claims.RegisteredClaims.Issuer == "" || claims.RegisteredClaims.Subject == "" {
		return ""
	}
	return claims.RegisteredClaims.Issuer + "|" + claims.RegisteredClaims.Subject
}
