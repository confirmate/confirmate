import { listTargetOfEvaluationAssessmentResults } from '$lib/api/orchestrator';

import { error } from '@sveltejs/kit';

import type { PageLoad } from './$types';
import { listGraphEdges, listResources } from '$lib/api/evidence_store';

export const load = (async ({ fetch, params, url }) => {
	if (params.id == undefined) {
		error(405, 'Required parameter missing');
	}

	const results = await listTargetOfEvaluationAssessmentResults(params.id, fetch);
	// const edges = await listGraphEdges();
	const resource = await listResources(params.id, '', fetch);
	const page = Number(url.searchParams.get('page'));

	return {
		results,
		// edges,
		page
	};
}) satisfies PageLoad;
