// Copyright 2016-2026 Fraunhofer AISEC
//
// SPDX-License-Identifier: Apache-2.0
//
// This file is part of Confirmate Core.

package server

import (
	"strings"

	"confirmate.io/core/api/orchestrator"
)

// normalizeRole translates a raw role string from a JWT into the typed
// [orchestrator.Role] enum. The translation table covers the IdP-specific
// names we know about (EMERALD/Keycloak-style: ORCHESTRATOR_ADMIN,
// "Compliance Manager", ...) plus the canonical ROLE_* enum names; everything
// else maps to Role_ROLE_UNSPECIFIED and is dropped by the caller.
//
// Matching is case-insensitive.
func normalizeRole(raw string) orchestrator.Role {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return orchestrator.Role_ROLE_UNSPECIFIED
	}

	// Canonical enum names short-circuit so a token that already uses ROLE_*
	// is a no-op.
	if v, ok := orchestrator.Role_value[strings.ToUpper(trimmed)]; ok {
		return orchestrator.Role(v)
	}

	switch strings.ToLower(trimmed) {
	case "orchestrator_admin", "admin":
		return orchestrator.Role_ROLE_ADMIN
	case "compliance manager", "compliance_manager":
		return orchestrator.Role_ROLE_COMPLIANCE_MANAGER
	case "expert compliance manager", "expert_compliance_manager":
		return orchestrator.Role_ROLE_EXPERT_COMPLIANCE_MANAGER
	case "internal control owner", "internal_control_owner":
		return orchestrator.Role_ROLE_INTERNAL_CONTROL_OWNER
	case "technical implementer", "technical_implementer":
		return orchestrator.Role_ROLE_TECHNICAL_IMPLEMENTER
	case "auditor":
		return orchestrator.Role_ROLE_AUDITOR
	case "chief information security officer", "ciso":
		return orchestrator.Role_ROLE_CHIEF_INFORMATION_SECURITY_OFFICER
	}

	return orchestrator.Role_ROLE_UNSPECIFIED
}
