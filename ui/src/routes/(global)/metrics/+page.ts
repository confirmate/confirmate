import type { PageLoad } from './$types';

export const load = (async ({ parent }) => {
	const { metrics } = await parent();
	return { metrics };
}) satisfies PageLoad;
