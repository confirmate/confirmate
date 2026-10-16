import { mockSource } from './mock';
import { sourceMode } from './mode.svelte';
import { realSource } from './real';
import type { GuidedSource } from './source';

/**
 * The source that the screens use. It follows the mode that the user chose (real data or sample data) at the time of
 * every call, so that switching needs no reload.
 */
export const guidedSource: GuidedSource = {
	listTargets: () => (sourceMode.current === 'sample' ? mockSource : realSource).listTargets(),
	getTarget: (id) => (sourceMode.current === 'sample' ? mockSource : realSource).getTarget(id),
	getRequirement: (t, r) => (sourceMode.current === 'sample' ? mockSource : realSource).getRequirement(t, r),
	listRequirementGroups: (c) => (sourceMode.current === 'sample' ? mockSource : realSource).listRequirementGroups(c),
	createTarget: (input) => (sourceMode.current === 'sample' ? mockSource : realSource).createTarget(input),
	addStatement: (s) => (sourceMode.current === 'sample' ? mockSource : realSource).addStatement(s),
	changeScope: (c) => (sourceMode.current === 'sample' ? mockSource : realSource).changeScope(c),
	getRiskAssessment: (t) => (sourceMode.current === 'sample' ? mockSource : realSource).getRiskAssessment(t),
	listAcceptedRisks: (t) => (sourceMode.current === 'sample' ? mockSource : realSource).listAcceptedRisks(t)
};
