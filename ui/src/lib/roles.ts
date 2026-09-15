const ROLE_LABELS: Record<string, string> = {
	ROLE_ADMIN: 'Admin',
	ROLE_COMPLIANCE_MANAGER: 'Compliance Manager',
	ROLE_EXPERT_COMPLIANCE_MANAGER: 'Expert Compliance Manager',
	ROLE_INTERNAL_CONTROL_OWNER: 'Control Owner',
	ROLE_TECHNICAL_IMPLEMENTER: 'Technical Implementer',
	ROLE_AUDITOR: 'Auditor',
	ROLE_CHIEF_INFORMATION_SECURITY_OFFICER: 'CISO'
};

export function roleLabel(role: string): string {
	return ROLE_LABELS[role] ?? role.replace('ROLE_', '').replaceAll('_', ' ');
}
