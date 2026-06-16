// Copyright 2016-2026 Fraunhofer AISEC
//
// SPDX-License-Identifier: Apache-2.0
//
// This file is part of Confirmate Core.

package server

import (
	"testing"

	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/util/assert"
)

func TestNormalizeRoleString(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "empty input returns empty", in: "", want: ""},
		{name: "canonical ROLE_ADMIN passes through", in: "ROLE_ADMIN", want: orchestrator.RoleAdmin},
		{name: "canonical lowercase normalized to upper", in: "role_admin", want: orchestrator.RoleAdmin},
		{name: "EMERALD ORCHESTRATOR_ADMIN maps to ROLE_ADMIN", in: "ORCHESTRATOR_ADMIN", want: orchestrator.RoleAdmin},
		{name: "shorthand admin maps to ROLE_ADMIN", in: "admin", want: orchestrator.RoleAdmin},
		{name: "Compliance Manager (with space) maps to enum", in: "Compliance Manager", want: orchestrator.Role_ROLE_COMPLIANCE_MANAGER.String()},
		{name: "compliance_manager (snake case) maps to enum", in: "compliance_manager", want: orchestrator.Role_ROLE_COMPLIANCE_MANAGER.String()},
		{name: "auditor maps to enum", in: "auditor", want: orchestrator.Role_ROLE_AUDITOR.String()},
		{name: "CISO shorthand maps to enum", in: "ciso", want: orchestrator.Role_ROLE_CHIEF_INFORMATION_SECURITY_OFFICER.String()},
		{name: "unknown role is passed through", in: "SOME_OTHER_ROLE", want: "SOME_OTHER_ROLE"},
		{name: "whitespace is trimmed", in: "  admin  ", want: orchestrator.RoleAdmin},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, normalizeRoleString(tt.in))
		})
	}
}
