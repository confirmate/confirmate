import { guidedSource } from '$lib/guided/select';
import type { LayoutLoad } from './$types';

export const load = (async ({ params }) => {
	return { target: await guidedSource.getTarget(params.id) };
}) satisfies LayoutLoad;
