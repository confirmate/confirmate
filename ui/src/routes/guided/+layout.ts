import { redirect } from '@sveltejs/kit';
import { resolve } from '$app/paths';
import { guidedEnabled } from '$lib/features';
import type { LayoutLoad } from './$types';

export const load = (async ({ parent }) => {
	// Wait for the login check of the root layout.
	await parent();

	// The guided experience is a preview, existing deployments keep the current screens.
	if (!guidedEnabled()) {
		redirect(307, resolve('/dashboard/'));
	}

	return {};
}) satisfies LayoutLoad;
