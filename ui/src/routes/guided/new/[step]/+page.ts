import { redirect } from '@sveltejs/kit';
import { resolve } from '$app/paths';
import { profiles } from '$lib/guided/profiles';
import { guidedSource } from '$lib/guided/select';
import { wizard } from '$lib/guided/wizard.svelte';
import type { PageLoad } from './$types';

export const load = (async ({ params }) => {
	// Only the CRA profile has a guided setup so far.
	const profile = profiles[0];
	const step = profile.steps.find((s) => s.id === params.step);
	if (!step) {
		redirect(307, resolve('/guided/new/'));
	}

	// A reload or a direct link starts the wizard with the defaults. The requirements come from the catalog.
	if (!wizard.started) wizard.start(profile, await guidedSource.listRequirementGroups(profile.catalogId));

	return { profile, step };
}) satisfies PageLoad;
