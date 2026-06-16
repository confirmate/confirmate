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

// normalizeRoleString translates a raw role string from the JWT into the
// canonical ROLE_* string used elsewhere in the codebase (matching the
// orchestrator's [orchestrator.Role] enum). The translation table covers the
// IdP-specific names we know about (EMERALD/Keycloak-style: ORCHESTRATOR_ADMIN,
// "Compliance Manager", ...); unknown role strings are passed through
// unchanged so callers can still consult them via claims.HasRole for
// project-specific access decisions.
//
// Matching is case-insensitive. Pure ROLE_* strings short-circuit to
// themselves so a token that already uses canonical names is a no-op.
func normalizeRoleString(raw string) string {
	if raw == "" {
		return ""
	}

	// Strip whitespace and try exact match against the enum names first — those
	// are already canonical and need no translation.
	trimmed := strings.TrimSpace(raw)
	if _, ok := orchestrator.Role_value[strings.ToUpper(trimmed)]; ok {
		return strings.ToUpper(trimmed)
	}

	// Known Keycloak / EMERALD role aliases. Keep this table small and explicit:
	// the goal isn't to cover every conceivable IdP naming, it's to make the
	// roles the EMERALD realm currently emits land on the right enum.
	switch strings.ToLower(trimmed) {
	case "orchestrator_admin", "admin":
		return orchestrator.RoleAdmin
	case "compliance manager", "compliance_manager":
		return orchestrator.Role_ROLE_COMPLIANCE_MANAGER.String()
	case "expert compliance manager", "expert_compliance_manager":
		return orchestrator.Role_ROLE_EXPERT_COMPLIANCE_MANAGER.String()
	case "internal control owner", "internal_control_owner":
		return orchestrator.Role_ROLE_INTERNAL_CONTROL_OWNER.String()
	case "technical implementer", "technical_implementer":
		return orchestrator.Role_ROLE_TECHNICAL_IMPLEMENTER.String()
	case "auditor":
		return orchestrator.Role_ROLE_AUDITOR.String()
	case "chief information security officer", "ciso":
		return orchestrator.Role_ROLE_CHIEF_INFORMATION_SECURITY_OFFICER.String()
	}

	// Unknown role: keep the original so the caller can still use it for
	// project-specific access decisions via claims.HasRole.
	return trimmed
}
