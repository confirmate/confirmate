import { orchestratorClient } from '$lib/api/client';
import { error } from '@sveltejs/kit';
import type { PageLoad } from './$types';

export const load = (async ({ params, fetch }) => {
	const client = orchestratorClient(fetch);

	const res = await client.GET('/v1/orchestrator/assessment_results/{id}', {
		params: {
			path: { id: params.resultId }
		}
	});

	if (!res.response.ok) {
		error(res.response.status, res.response.statusText);
	}

	return {
		result: res.data!
	};
}) satisfies PageLoad;