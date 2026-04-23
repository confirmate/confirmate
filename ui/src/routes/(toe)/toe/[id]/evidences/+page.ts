import { evidenceClient } from '$lib/api/client';
import { error } from '@sveltejs/kit';
import type { PageLoad } from './$types';

export const load = (async ({ params, fetch }) => {
	const client = evidenceClient(fetch);

	const res = await client.GET('/v1/evidence_store/evidences', {
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

	return {
		evidences: res.data?.evidences ?? []
	};
}) satisfies PageLoad;