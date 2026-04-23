import { evidenceClient } from '$lib/api/client';
import { error } from '@sveltejs/kit';
import type { PageLoad } from './$types';

export const load = (async ({ params, fetch }) => {
	const client = evidenceClient(fetch);

	const res = await client.GET('/v1/evidence_store/evidences/{evidenceId}', {
		params: {
			path: { evidenceId: params.id2 }
		}
	});

	if (res.response.status === 404) {
		return {
			evidence: null,
			notFound: true,
			evidenceId: params.id2
		};
	}

	if (!res.response.ok) {
		error(res.response.status, res.response.statusText);
	}

	return {
		evidence: res.data!,
		notFound: false
	};
}) satisfies PageLoad;