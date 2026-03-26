import { error, redirect } from '@sveltejs/kit';
import type { PageLoad } from './$types';

export const load = (async ({ params }) => {
	if (params.id == undefined) {
		error(405, 'Required parameter missing');
	}

	// In case someone access the "root" of a Target of Evaluation, we redirect him to
	// the activity view
	redirect(301, '/cloud/' + params.id + '/activity');
}) satisfies PageLoad;
