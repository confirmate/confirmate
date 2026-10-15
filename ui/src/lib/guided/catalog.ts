import type {
	Applicability,
	ApplicabilityKind,
	Connectivity,
	OperatorSkill,
	PhysicalAccess,
	Reference,
	RequirementGroup,
	RequirementType,
	RiskScenario,
	ScopableRequirement
} from './profile';

/**
 * A control as the API returns it (camelCase) or as a catalog file stores it (snake_case). Both are read, so that the
 * sample data can use the catalog file of example-data directly.
 */
type Raw = Record<string, unknown>;

/** A category of a catalog with the top-level controls that belong to it. */
export interface RawCategory {
	name: string;
	controls: Raw[];
}

function get(o: Raw | undefined, camel: string, snake: string): unknown {
	return o ? (o[camel] ?? o[snake]) : undefined;
}

function str(v: unknown): string {
	return typeof v === 'string' ? v : '';
}

function num(v: unknown): number {
	return typeof v === 'number' ? v : 0;
}

function list(v: unknown): Raw[] {
	return Array.isArray(v) ? (v as Raw[]) : [];
}

/** Turns an enum name such as APPLICABILITY_KIND_RISK_BASED into its last part after the prefix. */
function enumValue(v: unknown, prefix: string): string {
	return str(v).replace(prefix, '');
}

function enumList<T extends string>(v: unknown, prefix: string): T[] {
	return (Array.isArray(v) ? v : []).map((x) => enumValue(x, prefix) as T);
}

const kinds: Record<string, ApplicabilityKind> = { ALWAYS: 'always', CONDITIONAL: 'conditional', RISK_BASED: 'risk' };
const types: Record<string, RequirementType> = { TECHNICAL: 'technical', ACTIVITY: 'activity', DOCUMENTATION: 'documentation' };

function scenarioOf(raw: Raw | undefined): RiskScenario | undefined {
	if (!raw) return undefined;
	return {
		minConfidentiality: num(get(raw, 'minConfidentiality', 'min_confidentiality')),
		minIntegrity: num(get(raw, 'minIntegrity', 'min_integrity')),
		minAvailability: num(get(raw, 'minAvailability', 'min_availability')),
		physicalAccess: enumList<PhysicalAccess>(get(raw, 'physicalAccess', 'physical_access'), 'PHYSICAL_ACCESS_'),
		connectivity: enumList<Connectivity>(get(raw, 'connectivity', 'connectivity'), 'CONNECTIVITY_'),
		operatorSkill: enumList<OperatorSkill>(get(raw, 'operatorSkill', 'operator_skill'), 'OPERATOR_SKILL_')
	};
}

function applicabilityOf(raw: Raw | undefined): Applicability {
	// A requirement without an applicability always applies.
	if (!raw) return { kind: 'always', compensatedBy: [] };
	const kind = kinds[enumValue(raw.kind, 'APPLICABILITY_KIND_')] ?? 'always';
	return {
		kind,
		condition: str(raw.condition) || undefined,
		scenario: kind === 'risk' ? (scenarioOf(raw.scenario as Raw | undefined) ?? scenarioOf({})) : undefined,
		compensatedBy: (Array.isArray(get(raw, 'compensatedBy', 'compensated_by')) ? (get(raw, 'compensatedBy', 'compensated_by') as string[]) : [])
	};
}

function referencesOf(raw: unknown): Reference[] {
	return list(raw).map((r) => ({
		framework: str(r.framework),
		identifier: str(r.identifier),
		relationship: str(r.relationship),
		url: str(r.url) || undefined
	}));
}

/** The requirement behind a leaf control. */
export function requirementOf(control: Raw): ScopableRequirement {
	return {
		id: str(get(control, 'shortName', 'short_name')) || str(control.id),
		title: str(control.name),
		text: str(control.description),
		checks: list(control.metrics).length,
		type: types[enumValue(control.type, 'CONTROL_TYPE_')] ?? 'technical',
		applicability: applicabilityOf(control.applicability as Raw | undefined),
		references: referencesOf(control.references)
	};
}

/** Where the points of a legal text are cited, for example "Annex I, Part I, (2)(e)". */
function referenceOf(requirements: ScopableRequirement[], fallback: string): string {
	const legal = requirements.flatMap((r) => r.references).find((r) => r.relationship === 'legal-basis');
	return legal?.identifier ?? fallback;
}

const rank: Record<ApplicabilityKind, number> = { always: 0, risk: 1, conditional: 2 };

/**
 * Puts the baseline first, then the requirements that depend on the risk, then those that depend on a condition. A
 * requirement that can replace another one (it is named in its compensatedBy) comes before it: the generic requirement
 * before its refinement. Otherwise the order of the catalog stays.
 */
export function ordered(requirements: ScopableRequirement[]): ScopableRequirement[] {
	const replacing = new Set(requirements.flatMap((r) => r.applicability.compensatedBy));
	const key = (r: ScopableRequirement) => rank[r.applicability.kind] * 2 + (replacing.has(r.id) ? 0 : 1);
	return requirements.map((r, i) => ({ r, i })).sort((a, b) => key(a.r) - key(b.r) || a.i - b.i).map(({ r }) => r);
}

/**
 * Groups the controls of a catalog for the setup: every top-level control with sub-controls is one point of the law or
 * standard, and its sub-controls are the requirements behind it.
 */
export function groupsOf(categories: RawCategory[]): RequirementGroup[] {
	return categories.flatMap((category) =>
		category.controls.map((control): RequirementGroup => {
			const children = list(control.controls);
			const requirements = ordered((children.length > 0 ? children : [control]).map(requirementOf));
			return {
				id: str(get(control, 'shortName', 'short_name')) || str(control.id),
				title: str(control.name),
				reference: referenceOf(requirements, category.name),
				part: category.name,
				requirements
			};
		})
	);
}

/** A statement that the person has to confirm or deny, because requirements depend on it. */
export interface Condition {
	id: string;
	text: string;
	/** The requirements that only apply if the statement is true. */
	requirements: string[];
}

export function conditionId(text: string): string {
	return text
		.toLowerCase()
		.replace(/[^a-z0-9]+/g, '-')
		.replace(/^-|-$/g, '');
}

/** The distinct conditions of the requirements, in the order in which they first appear. */
export function conditionsOf(groups: RequirementGroup[]): Condition[] {
	const byId = new Map<string, Condition>();
	for (const r of groups.flatMap((g) => g.requirements)) {
		const text = r.applicability.condition;
		if (r.applicability.kind !== 'conditional' || !text) continue;
		const id = conditionId(text);
		const existing = byId.get(id);
		if (existing) existing.requirements.push(r.id);
		else byId.set(id, { id, text, requirements: [r.id] });
	}
	return [...byId.values()];
}
