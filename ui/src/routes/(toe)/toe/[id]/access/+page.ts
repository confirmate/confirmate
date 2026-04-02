import type { PageLoad } from './$types';

export const load = (async ({ parent }) => {
	const { toe } = await parent();
	return { toe };
}) satisfies PageLoad;
