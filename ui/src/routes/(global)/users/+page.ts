import { orchestratorClient } from '#lib/api/client.js';
import type { PageLoad } from './$types';

export const load = (async ({ fetch }) => {
	const client = orchestratorClient(fetch);
	const { data } = await client.GET('/v1/users', {});
	return { users: data?.users ?? [] };
}) satisfies PageLoad;
