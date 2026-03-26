import { listTargetOfEvaluationAssessmentResults } from '$lib/api/orchestrator';

import { error } from '@sveltejs/kit';

import type { PageLoad } from './$types';

export const load = (async ({ fetch, params, url }) => {
	if (params.id == undefined) {
		error(405, 'Required parameter missing');
	}

	const filterIds = url.searchParams.get('filterIds')?.split(',') ?? [];
	const filterResourceId = url.searchParams.get('filterResourceId');
	const page = Number(url.searchParams.get('page'));
	const results = await listTargetOfEvaluationAssessmentResults(params.id, fetch);

	return {
		filterIds,
		filterResourceId,
		page,
		results
	};
}) satisfies PageLoad;
