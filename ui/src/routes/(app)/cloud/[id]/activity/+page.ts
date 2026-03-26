import { listTargetOfEvaluationAssessmentResults } from '$lib/api/orchestrator';
import { error } from '@sveltejs/kit';
import type { PageLoad } from './$types';

export const load = (async ({ fetch, params }) => {
	if (params.id == undefined) {
		error(405, 'Required parameter missing');
	}

	const results = await listTargetOfEvaluationAssessmentResults(params.id, fetch);

	return {
		results: results
	};
}) satisfies PageLoad;
