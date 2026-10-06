import { resolve } from '$app/paths';
import { redirect } from '@sveltejs/kit';
import type { PageLoad } from './$types';

export const load = (({ params }) => {
	throw redirect(307, resolve(`/toe/${params.id}/audit-scopes/`));
}) satisfies PageLoad;
