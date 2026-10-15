import { guidedSource } from '$lib/guided/select';
import type { PageLoad } from './$types';

export const load = (async ({ params }) => {
	return { risks: await guidedSource.listAcceptedRisks(params.id) };
}) satisfies PageLoad;
