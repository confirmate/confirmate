import { orchestratorClient } from '$lib/api/client';
import type { LayoutLoad } from './$types';

export const ssr = false;
export const trailingSlash = 'always';

export const load = (async ({ fetch }) => {
	const client = orchestratorClient(fetch);

	const [servicesRes, catalogsRes, metricsRes] = await Promise.all([
		client.GET('/v1/orchestrator/targets_of_evaluation', {}),
		client.GET('/v1/orchestrator/catalogs', {}),
		client.GET('/v1/orchestrator/metrics', { params: { query: { pageSize: 200 } } })
	]);

	const services = servicesRes.data?.targetsOfEvaluation ?? [];
	const catalogs = catalogsRes.data?.catalogs ?? [];
	const metricList = metricsRes.data?.metrics ?? [];
	const metrics = new Map(metricList.map((m) => [m.id, m]));

	return { services, catalogs, metrics };
}) satisfies LayoutLoad;
