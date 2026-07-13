import type { PageLoad } from './$types';

export const load = (async ({ parent }) => {
	const { catalogs, auditScopes, toe } = await parent();
	return { catalogs, auditScopes, toe };
}) satisfies PageLoad;
