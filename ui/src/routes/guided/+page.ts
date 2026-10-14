import { guidedSource } from '$lib/guided/source';
import type { PageLoad } from './$types';

export const load = (async () => {
	return { targets: await guidedSource.listTargets() };
}) satisfies PageLoad;
