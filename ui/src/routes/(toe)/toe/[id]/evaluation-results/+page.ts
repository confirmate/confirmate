import { orchestratorClient } from '$lib/api/client';
import { error } from '@sveltejs/kit';
import type { PageLoad } from './$types';

export const load = (async ({ params, fetch }) => {
	const client = orchestratorClient(fetch);

	const res = await client.GET('/v1/evaluation/results', {
		params: {
			query: {
				'filter.targetOfEvaluationId': params.id,
				pageSize: 1000
			}
		}
	});

	if (!res.response.ok) {
		error(res.response.status, res.response.statusText);
	}

	const results = res.data?.results ?? [];
	// Sort by timestamp descending (most recent first)
	results.sort((a, b) => {
		const timeA = a.timestamp ? new Date(a.timestamp).getTime() : 0;
		const timeB = b.timestamp ? new Date(b.timestamp).getTime() : 0;
		return timeB - timeA;
	});

	return {
		evaluationResults: results
	};
}) satisfies PageLoad;