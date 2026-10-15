import type { Condition } from './catalog';
import type { Answer, FrameworkProfile } from './profile';

/**
 * The answers that the choices of the setup suggest for the conditions of the requirements, for example that a library
 * has no user accounts. A condition that nothing mentions starts as "yes", so that its requirements stay.
 */
export function suggestedAnswers(
	profile: FrameworkProfile,
	details: Record<string, string>,
	conditions: Condition[]
): Record<string, { answer: Answer; fromChoice: boolean }> {
	const byText = new Map<string, Answer>();
	for (const field of profile.details) {
		const defaults = field.conditionDefaults?.[details[field.id]];
		for (const [text, answer] of Object.entries(defaults ?? {})) byText.set(text, answer);
	}

	return Object.fromEntries(
		conditions.map((c) => [c.id, { answer: byText.get(c.text) ?? 'yes', fromChoice: byText.has(c.text) }])
	);
}
