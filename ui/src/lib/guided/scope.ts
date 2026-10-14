import type { Answer, Questionnaire } from './profile';

export interface ScopedRequirement {
	id: string;
	title: string;
	out: boolean;
	reason: string;
}

export interface ScopeResult {
	requirements: ScopedRequirement[];
	applyCount: number;
}

/**
 * Works out which requirements apply from the answers of a questionnaire. An effect only switches requirements out
 * if every one of its conditions is an explicit match, so "don't know" always keeps a requirement in scope.
 */
export function computeScope(q: Questionnaire, answers: Record<string, Answer>): ScopeResult {
	const reasons = new Map<string, string>();

	for (const effect of q.effects) {
		const matches = Object.entries(effect.when).every(([id, expected]) => answers[id] === expected);
		if (!matches) continue;
		for (const id of effect.out) {
			if (!reasons.has(id)) reasons.set(id, effect.reason);
		}
	}

	const requirements = q.requirements.map((r) => ({
		id: r.id,
		title: r.title,
		out: reasons.has(r.id),
		reason: reasons.get(r.id) ?? ''
	}));

	return { requirements, applyCount: requirements.filter((r) => !r.out).length };
}

/** The notes of a questionnaire that apply to the current answers. */
export function activeNotes(q: Questionnaire, answers: Record<string, Answer>) {
	return q.notes.filter((n) => Object.entries(n.when).every(([id, expected]) => answers[id] === expected));
}
