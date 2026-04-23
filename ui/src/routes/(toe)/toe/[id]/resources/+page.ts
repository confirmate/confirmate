import type { PageLoad } from './$types';
import { evidenceClient, orchestratorClient } from '$lib/api/client';
import type { SchemaResourceSnapshot as Resource, SchemaGraphEdge as GraphEdge } from '$lib/api/openapi/evidence';
import type { SchemaAssessmentResult as AssessmentResult } from '$lib/api/openapi/orchestrator';

export const load = (async ({ fetch, params }) => {
	const toeId = params.id;

	const [resourcesRes, edgesRes, resultsRes] = await Promise.all([
		evidenceClient(fetch).GET('/v1/evidence_store/resources', {
			params: {
				query: {
					'filter.targetOfEvaluationId': toeId,
					pageSize: 1500,
					orderBy: 'resource_type',
					asc: true
				}
			}
		}),
		evidenceClient(fetch).GET('/v1/evidence_store/resources/edges', {}),
		orchestratorClient(fetch).GET('/v1/orchestrator/assessment_results', {
			params: {
				query: {
					'filter.targetOfEvaluationId': toeId,
					latestByResourceId: true,
					pageSize: 5000
				}
			}
		})
	]);

	const resources = (resourcesRes.data?.results ?? []) as Resource[];
	const resourceIds = new Set(resources.map((r) => r.id));
	const edges = ((edgesRes.data?.edges ?? []) as GraphEdge[]).filter(
		(e) => resourceIds.has(e.source) && resourceIds.has(e.target)
	);

	return {
		resources,
		edges,
		results: (resultsRes.data?.results ?? []) as AssessmentResult[]
	};
}) satisfies PageLoad;
