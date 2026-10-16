import { guidedSource } from '$lib/guided/select';
import type { PageLoad } from './$types';

export const load = (async ({ params }) => {
	return { assessment: await guidedSource.getRiskAssessment(params.id) };
}) satisfies PageLoad;
