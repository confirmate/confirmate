import { page } from '$app/state';
import type { SchemaMetric } from '$lib/api/openapi/orchestrator';

/**
 * Returns the display name of the metric with the given ID. Metric IDs are UUIDs, so pages show
 * the name instead. The metrics come from the root layout, which loads them once for all pages.
 * Falls back to the ID itself if the metric is not known, e.g. because it has been removed.
 */
export function metricName(id: string | undefined): string {
	if (!id) return '';

	const metrics = (page.data as { metrics?: Map<string, SchemaMetric> }).metrics;
	return metrics?.get(id)?.name || id;
}
