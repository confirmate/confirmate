import type { Answer, Questionnaire, RequirementGroup, ScopableRequirement } from './profile';

/** What is true for a requirement: it applies, it is not applicable, or it applies and a residual risk is accepted. */
export type Treatment = 'applies' | 'na' | 'risk';

/** A decision of a person about one requirement. */
export interface Decision {
	treatment: Treatment;
	reason: string;
}

/** The decisions of a person, by requirement ID, and the suggestions that the person has dismissed. */
export interface Decisions {
	decided: Record<string, Decision>;
	dismissed: Record<string, boolean>;
}

export interface ScopeRow {
	requirement: ScopableRequirement;
	treatment: Treatment;
	reason: string;
	/** Where the treatment comes from. */
	source: 'default' | 'rule' | 'person';
	/** A suggestion that the person has not decided yet. */
	suggestion?: { treatment: 'na' | 'risk'; reason: string };
	/** The requirement cannot be taken out. */
	locked: boolean;
	/** The treatment is not applicable or accepted risk, but there is no reason yet. */
	missingReason: boolean;
}

export interface ScopeGroup {
	group: RequirementGroup;
	rows: ScopeRow[];
}

export interface ScopeSummary {
	total: number;
	applies: number;
	na: number;
	risk: number;
	/** Suggestions without a decision. */
	suggestions: number;
	missingReasons: number;
	/** Requirements that still need a decision of a person. */
	review: number;
}

function matches(when: Partial<Record<string, Answer>>, answers: Record<string, Answer>): boolean {
	return Object.entries(when).every(([id, expected]) => answers[id] === expected);
}

/**
 * Works out the scope of every requirement from the answers of the questionnaire and the decisions of a person.
 *
 * - A mandatory requirement always applies.
 * - A requirement that only applies if the product has a capability is not applicable if the answer about that
 *   capability is an explicit no.
 * - A suggestion is only a suggestion until a person accepts or dismisses it. "Don't know" never produces one.
 * - A decision of a person always wins.
 */
export function computeScope(q: Questionnaire, answers: Record<string, Answer>, decisions: Decisions): ScopeGroup[] {
	return q.groups.map((group) => ({
		group,
		rows: group.requirements.map((requirement): ScopeRow => {
			const locked = requirement.applicability === 'mandatory';
			const decided = decisions.decided[requirement.id];

			let treatment: Treatment = 'applies';
			let reason = '';
			let source: ScopeRow['source'] = 'default';
			let suggestion: ScopeRow['suggestion'];

			if (!locked) {
				const capability = requirement.capability;
				if (capability && answers[capability.question] === 'no') {
					treatment = 'na';
					reason = `Only applies if ${capability.text}. It does not.`;
					source = 'rule';
				} else {
					const hit = q.suggestions.find((s) => s.requirements.includes(requirement.id) && matches(s.when, answers));
					if (hit) suggestion = { treatment: hit.treatment, reason: hit.reason };
				}
			}

			if (!locked && decided) {
				treatment = decided.treatment;
				reason = decided.reason;
				source = 'person';
			}
			if (decisions.dismissed[requirement.id] || (decided && decided.treatment !== 'applies') || source === 'person') {
				suggestion = undefined;
			}

			return {
				requirement,
				treatment,
				reason,
				source,
				suggestion,
				locked,
				missingReason: treatment !== 'applies' && !reason.trim()
			};
		})
	}));
}

/** Counts the requirements of a scope. */
export function summarize(groups: ScopeGroup[]): ScopeSummary {
	const rows = groups.flatMap((g) => g.rows);
	const count = (t: Treatment) => rows.filter((r) => r.treatment === t).length;
	const suggestions = rows.filter((r) => r.suggestion).length;
	const missingReasons = rows.filter((r) => r.missingReason).length;

	return {
		total: rows.length,
		applies: count('applies'),
		na: count('na'),
		risk: count('risk'),
		suggestions,
		missingReasons,
		review: rows.filter((r) => r.suggestion || r.missingReason).length
	};
}

/** The notes of a questionnaire that apply to the current answers. */
export function activeNotes(q: Questionnaire, answers: Record<string, Answer>) {
	return q.notes.filter((n) => matches(n.when, answers));
}
