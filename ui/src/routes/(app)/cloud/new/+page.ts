import type { PageLoad } from './$types';

export const load = (async ({ url }) => {
	const step = Number.parseInt(url.searchParams.get('step') ?? '0');

	return {
		step
	};
}) satisfies PageLoad;
