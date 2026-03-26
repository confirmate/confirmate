import { listMetrics } from '$lib/api/orchestrator';

export const load = async () => {
	const metrics = await listMetrics();

	return {
		metrics: metrics
	};
};
