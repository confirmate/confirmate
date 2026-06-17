import type { PageLoad } from './$types';

export const load = (async ({ parent }) => {
	const { catalogs, toe } = await parent();
	return { catalogs, toe };
}) satisfies PageLoad;
