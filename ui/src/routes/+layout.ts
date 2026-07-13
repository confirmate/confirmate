import { orchestratorClient } from '$lib/api/client';
import { isAuthenticated, login } from '$lib/auth';
import type { LayoutLoad } from './$types';

export const ssr = false;
export const trailingSlash = 'always';

export const load = (async ({ url, fetch }) => {
	// Skip auth check for the OAuth callback route
	if (url.pathname.startsWith('/auth/')) {
		return { services: [], catalogs: [], metrics: new Map(), currentUser: null };
	}

	if (!isAuthenticated()) {
		await login(url.pathname + url.search);
		return { services: [], catalogs: [], metrics: new Map(), currentUser: null };
	}

	const client = orchestratorClient(fetch);

	const [servicesRes, catalogsRes, metricsRes, currentUserRes] = await Promise.all([
		client.GET('/v1/orchestrator/targets_of_evaluation', {}),
		client.GET('/v1/orchestrator/catalogs', {}),
		client.GET('/v1/orchestrator/metrics', { params: { query: { pageSize: 200 } } }),
		client.GET('/v1/users/me', {})
	]);

	const services = servicesRes.data?.targetsOfEvaluation ?? [];
	const catalogs = catalogsRes.data?.catalogs ?? [];
	const metricList = metricsRes.data?.metrics ?? [];
	const metrics = new Map(metricList.map((m) => [m.id, m]));
	const currentUser = currentUserRes.data ?? null;

	return { services, catalogs, metrics, currentUser };
}) satisfies LayoutLoad;
