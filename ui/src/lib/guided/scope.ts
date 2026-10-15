import { conditionId } from './catalog';
import type { Answer, RequirementGroup, RiskProfile, ScopableRequirement } from './profile';
import { describeScenario, scenarioReached } from './risk';

/** What is true for a requirement: it applies, it is not applicable, or it applies and a residual risk is accepted. */
export type Treatment = 'applies' | 'na' | 'risk';

/**
 * Why a requirement is not applicable. It is one of a fixed set, so that the reasons can be compared and reported. The
 * names are those of the API (NotApplicableReason).
 */
export type NotApplicableReason =
	| 'TARGET_ABSENT'
	| 'CONDITION_NOT_MET'
	| 'COMPENSATED'
	| 'CONFLICTING_REGULATION'
	| 'BELOW_RISK_LEVEL';

export const notApplicableReasons: { value: NotApplicableReason; label: string; hint: string }[] = [
	{ value: 'TARGET_ABSENT', label: 'It protects something the product does not have', hint: 'For example no stored data, no interface' },
	{ value: 'CONDITION_NOT_MET', label: 'Its condition does not hold', hint: 'The requirement only applies in a case that does not occur' },
	{ value: 'COMPENSATED', label: 'Another requirement is fulfilled instead', hint: 'It must give at least the same protection' },
	{ value: 'CONFLICTING_REGULATION', label: 'Another regulation does not allow it', hint: 'Say which one in the reason' },
	{ value: 'BELOW_RISK_LEVEL', label: 'The risk is below the level at which it applies', hint: 'According to your risk assessment' }
];

/** A decision of a person about one requirement. */
export interface Decision {
	treatment: Treatment;
	reason: string;
	/** For "not applicable": why. */
	naReason?: NotApplicableReason;
	/** For the reason "compensated": the requirement that is fulfilled instead. */
	compensatedBy?: string;
}

/** The decisions of a person, by requirement ID. */
export interface Decisions {
	decided: Record<string, Decision>;
}

export interface ScopeRow {
	requirement: ScopableRequirement;
	treatment: Treatment;
	reason: string;
	naReason?: NotApplicableReason;
	compensatedBy?: string;
	/** Where the treatment comes from. */
	source: 'default' | 'rule' | 'person';
	/** What the answers and the risk profile say about the requirement, whatever the person decided. */
	rule: { treatment: 'applies' | 'na'; reason: string; naReason?: NotApplicableReason };
	/** The requirement cannot be taken out. */
	locked: boolean;
	/** The treatment is not applicable or accepted risk, but there is no reason yet. */
	missingReason: boolean;
	/** Taking the requirement out would leave no requirement of its point, so "not applicable" is not offered. */
	naBlocked: boolean;
	/** Why the requirement is not applicable but the person has to say which requirement replaces it. */
	missingCompensation: boolean;
}

export interface ScopeGroup {
	group: RequirementGroup;
	rows: ScopeRow[];
	/** Every requirement of the point is not applicable. A point cannot drop out as a whole. */
	empty: boolean;
}

export interface ScopeSummary {
	total: number;
	applies: number;
	na: number;
	risk: number;
	/** Requirements whose treatment depends on the product: conditional or risk based. */
	depends: number;
	missingReasons: number;
	/** Points in which every requirement is not applicable. */
	emptyGroups: number;
	/** Requirements and points that still need a decision or a reason of a person. */
	review: number;
}

/** What the answers and the risk profile say about one requirement. */
function ruleOf(r: ScopableRequirement, risk: RiskProfile, answers: Record<string, Answer>): ScopeRow['rule'] {
	const a = r.applicability;

	if (a.kind === 'conditional' && a.condition && answers[conditionId(a.condition)] === 'no') {
		return {
			treatment: 'na',
			naReason: 'CONDITION_NOT_MET',
			reason: `It only applies if this is true: "${a.condition}". You answered no.`
		};
	}
	if (a.kind === 'risk' && !scenarioReached(risk, a.scenario)) {
		return {
			treatment: 'na',
			naReason: 'BELOW_RISK_LEVEL',
			reason: `${a.scenario ? describeScenario(a.scenario) : ''} Your risk profile is below that.`.trim()
		};
	}
	return { treatment: 'applies', reason: '' };
}

/**
 * Works out the scope of every requirement from the risk profile, the answers about the conditions and the decisions of
 * a person.
 *
 * - A requirement that always applies cannot be taken out.
 * - A conditional requirement is not applicable if the answer about its condition is an explicit no. "Don't know" keeps it.
 * - A risk based requirement is not applicable if the risk profile does not reach its scenario. If that would take out
 *   every requirement of a point, those requirements stay, as on the server: a point cannot drop out as a whole.
 * - A decision of a person always wins.
 */
export function computeScope(
	groups: RequirementGroup[],
	risk: RiskProfile,
	answers: Record<string, Answer>,
	decisions: Decisions
): ScopeGroup[] {
	return groups.map((group) => {
		const rules = group.requirements.map((r) => ruleOf(r, risk, answers));

		// Keep the risk based requirements of a point if the rules would take out all of its requirements.
		if (rules.length > 0 && rules.every((r) => r.treatment === 'na')) {
			rules.forEach((rule, i) => {
				if (group.requirements[i].applicability.kind === 'risk') rules[i] = { treatment: 'applies', reason: '' };
			});
		}

		const rows = group.requirements.map((requirement, i): ScopeRow => {
			const rule = rules[i];
			const locked = requirement.applicability.kind === 'always';
			const decided = locked ? undefined : decisions.decided[requirement.id];

			const treatment: Treatment = decided ? decided.treatment : rule.treatment;
			const reason = decided ? decided.reason : rule.reason;
			const naReason = decided ? decided.naReason : rule.naReason;
			const compensatedBy = decided?.compensatedBy;

			return {
				requirement,
				treatment,
				reason,
				naReason: treatment === 'na' ? naReason : undefined,
				compensatedBy,
				source: decided ? 'person' : rule.treatment === 'na' ? 'rule' : 'default',
				rule,
				locked,
				missingReason: treatment !== 'applies' && !reason.trim(),
				naBlocked: false,
				missingCompensation: treatment === 'na' && naReason === 'COMPENSATED' && !compensatedBy
			};
		});

		// A point cannot drop out as a whole: the last requirement that is not out cannot be taken out, and if the
		// answers have taken out all of them, a person has to resolve that.
		const remaining = rows.filter((r) => r.treatment !== 'na');
		for (const r of rows) r.naBlocked = remaining.length === 1 && remaining[0] === r;

		return { group, rows, empty: rows.length > 0 && remaining.length === 0 };
	});
}

/** Counts the requirements of a scope. */
export function summarize(groups: ScopeGroup[]): ScopeSummary {
	const rows = groups.flatMap((g) => g.rows);
	const count = (t: Treatment) => rows.filter((r) => r.treatment === t).length;
	const missingReasons = rows.filter((r) => r.missingReason || r.missingCompensation).length;
	const emptyGroups = groups.filter((g) => g.empty).length;

	return {
		total: rows.length,
		applies: count('applies'),
		na: count('na'),
		risk: count('risk'),
		depends: rows.filter((r) => r.requirement.applicability.kind !== 'always').length,
		missingReasons,
		emptyGroups,
		review: missingReasons + emptyGroups
	};
}
