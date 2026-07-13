import { orchestratorClient } from '$lib/api/client';
import { error } from '@sveltejs/kit';
import type { PageLoad } from './$types';

export const load = (async ({ params, url, fetch, depends }) => {
	// Re-run this load whenever the ?metric= query param changes — SvelteKit
	// reruns on path changes by default, not on search params.
	depends('app:assessment-results');

	const client = orchestratorClient(fetch);
	const metricFilter = url.searchParams.get('metric') ?? undefined;

	const res = await client.GET('/v1/orchestrator/assessment_results', {
		params: {
			query: {
				'filter.targetOfEvaluationId': params.id,
				...(metricFilter ? { 'filter.metricId': metricFilter } : {}),
				pageSize: 1000
			}
		}
	});

	if (!res.response.ok) {
		error(res.response.status, res.response.statusText);
	}

	return {
		assessmentResults: res.data?.results ?? [],
		metricFilter
	};
}) satisfies PageLoad;