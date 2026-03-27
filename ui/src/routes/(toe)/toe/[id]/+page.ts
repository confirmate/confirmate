import { redirect } from '@sveltejs/kit';
import type { PageLoad } from './$types';

export const load = (({ params }) => {
	redirect(307, `/toe/${params.id}/audit-scopes/`);
}) satisfies PageLoad;
