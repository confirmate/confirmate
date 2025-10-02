import type { PageLoad } from './$types';
import createClient from 'openapi-fetch';
import type { paths } from '$lib/api/orchestrator.ts';

export const load: PageLoad = async ({ fetch }) => {
	const client = createClient<paths>({ baseUrl: 'http://localhost:8080' });

	const {
		data, // only present if 2XX response
		error // only present if 4XX or 5XX response
	} = await client.GET('/v1/orchestrator/targets_of_evaluation', { fetch });

	return { toes: data?.targetsOfEvaluation ?? [], error };
};
