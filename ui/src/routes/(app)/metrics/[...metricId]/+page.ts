import { getMetric, getMetricConfiguration } from '$lib/api/orchestrator';

import { error } from '@sveltejs/kit';

import type { PageLoad } from './$types';

export const load = (async ({ params }) => {
	console.log(params);
	if (params.metricId.split('/').length < 2) {
		error(405, 'Required parameter missing');
	}

	// slugs[0] is the serviceId
	// slugs[1] is the metricId

	const slugs = params.metricId.split('/');
	const metric = await getMetric(slugs[1]);
	const metricConfiguration = await getMetricConfiguration(slugs[0], slugs[1]);

	return {
		metric,
		metricConfiguration
	};
}) satisfies PageLoad;
