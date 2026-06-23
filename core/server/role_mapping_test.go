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

func TestNormalizeRole(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want orchestrator.Role
	}{
		{name: "empty input returns UNSPECIFIED", in: "", want: orchestrator.Role_ROLE_UNSPECIFIED},
		{name: "whitespace returns UNSPECIFIED", in: "   ", want: orchestrator.Role_ROLE_UNSPECIFIED},
		{name: "canonical ROLE_ADMIN passes through", in: "ROLE_ADMIN", want: orchestrator.Role_ROLE_ADMIN},
		{name: "canonical lowercase normalized to upper", in: "role_admin", want: orchestrator.Role_ROLE_ADMIN},
		{name: "EMERALD ORCHESTRATOR_ADMIN maps to ROLE_ADMIN", in: "ORCHESTRATOR_ADMIN", want: orchestrator.Role_ROLE_ADMIN},
		{name: "shorthand admin maps to ROLE_ADMIN", in: "admin", want: orchestrator.Role_ROLE_ADMIN},
		{name: "Compliance Manager (with space) maps to enum", in: "Compliance Manager", want: orchestrator.Role_ROLE_COMPLIANCE_MANAGER},
		{name: "expert_compliance_manager maps to enum", in: "expert_compliance_manager", want: orchestrator.Role_ROLE_EXPERT_COMPLIANCE_MANAGER},
		{name: "internal_control_owner maps to enum", in: "internal_control_owner", want: orchestrator.Role_ROLE_INTERNAL_CONTROL_OWNER},
		{name: "technical_implementer maps to enum", in: "technical_implementer", want: orchestrator.Role_ROLE_TECHNICAL_IMPLEMENTER},
		{name: "internal_auditor maps to enum", in: "internal_auditor", want: orchestrator.Role_ROLE_INTERNAL_AUDITOR},
		{name: "technical_auditor maps to enum", in: "technical_auditor", want: orchestrator.Role_ROLE_TECHNICAL_AUDITOR},
		{name: "compliance_manager (snake case) maps to enum", in: "compliance_manager", want: orchestrator.Role_ROLE_COMPLIANCE_MANAGER},
		{name: "lead auditor maps to enum", in: "lead_auditor", want: orchestrator.Role_ROLE_LEAD_AUDITOR},
		{name: "CISO shorthand maps to enum", in: "ciso", want: orchestrator.Role_ROLE_CHIEF_INFORMATION_SECURITY_OFFICER},
		{name: "UI-Admin maps to enum", in: "UI-Admin", want: orchestrator.Role_ROLE_UI_ADMIN},
		{name: "unknown role returns UNSPECIFIED", in: "SOME_OTHER_ROLE", want: orchestrator.Role_ROLE_UNSPECIFIED},
		{name: "whitespace is trimmed", in: "  admin  ", want: orchestrator.Role_ROLE_ADMIN},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, normalizeRole(tt.in))
		})
	}
}
