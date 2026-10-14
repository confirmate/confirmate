import { error } from '@sveltejs/kit';
import { guidedSource } from '$lib/guided/select';
import type { PageLoad } from './$types';

export const load = (async ({ params }) => {
	const requirement = await guidedSource.getRequirement(params.id, params.reqId);
	if (!requirement) {
		error(404, 'This requirement has no details yet.');
	}

	return { requirement };
}) satisfies PageLoad;
