import { error } from '@sveltejs/kit';
import type { PageLoad } from './$types';

export const load = (async ({ params, parent }) => {
	const { target } = await parent();
	const requirement = target?.requirementDetails.find((r) => r.id === params.reqId);
	if (!requirement) {
		error(404, 'This requirement has no details in the preview yet.');
	}

	return { requirement };
}) satisfies PageLoad;
