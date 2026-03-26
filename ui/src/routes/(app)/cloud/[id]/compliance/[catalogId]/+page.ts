import { listEvaluationResults } from '$lib/api/evaluation';
import { getCatalog, listControls } from '$lib/api/orchestrator';

import { error } from '@sveltejs/kit';

import type { PageLoad } from './$types';

export const load = (async ({ fetch, params, url }) => {
	if (params.id == undefined) {
		error(405, 'Required parameter missing');
	}

	const catalog = await getCatalog(params.catalogId, fetch);
	const evaluations = await listEvaluationResults(
		{
			targetOfEvaluationId: params.id,
			catalogId: params.catalogId
		},
		true,
		fetch
	);

	const controls = new Map(
		(await listControls(params.catalogId, undefined, fetch)).map((c) => [c.id, c])
	);
	const filterStatus = url.searchParams.getAll('status');

	return {
		evaluations,
		catalog,
		controls,
		filterStatus
	};
}) satisfies PageLoad;
