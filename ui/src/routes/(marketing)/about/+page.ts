import { redirect } from '@sveltejs/kit';
import type { PageLoad } from './$types';

export const load = (async () => {
	redirect(301, '/about/start');
}) satisfies PageLoad;
