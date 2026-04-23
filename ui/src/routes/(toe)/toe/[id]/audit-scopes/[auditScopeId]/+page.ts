import { orchestratorClient } from '$lib/api/client';
import { error } from '@sveltejs/kit';
import type { SchemaControl, SchemaEvaluationResult } from '$lib/api/openapi/orchestrator';
import type { PageLoad } from './$types';

export const load = (async ({ params, fetch }) => {
	const client = orchestratorClient(fetch);

	const { data: auditScope, response } = await client.GET(
		'/v1/orchestrator/audit_scopes/{auditScopeId}',
		{ params: { path: { auditScopeId: params.auditScopeId } } }
	);

	if (!auditScope) error(response.status, response.statusText);

	const { data: catalog } = await client.GET('/v1/orchestrator/catalogs/{catalogId}', {
		params: { path: { catalogId: auditScope.catalogId } }
	});

	// Fetch controls per category in parallel
	const categories = catalog?.categories ?? [];
	const controlResults = await Promise.all(
		categories.map((cat) =>
			client.GET('/v1/orchestrator/catalogs/{catalogId}/categories/{categoryName}/controls', {
				params: { path: { catalogId: auditScope.catalogId, categoryName: cat.name } }
			})
		)
	);

	const allControls = controlResults.flatMap((r) => r.data?.controls ?? []);

	// Build hierarchy: nest sub-controls under their parents
	const controlsMap: Record<string, SchemaControl> = {};
	const topLevelControls: SchemaControl[] = [];

	for (const ctrl of allControls) {
		controlsMap[ctrl.id!] = { ...ctrl, controls: ctrl.controls ?? [] };
	}

	for (const ctrl of Object.values(controlsMap)) {
		if (ctrl.parentControlId && controlsMap[ctrl.parentControlId]) {
			controlsMap[ctrl.parentControlId].controls!.push(ctrl);
		} else if (!ctrl.parentControlId) {
			topLevelControls.push(ctrl);
		}
	}

	const controlsByCategory: Record<string, SchemaControl[]> = Object.fromEntries(
		categories.map((cat) => [
			cat.name,
			topLevelControls.filter((c) => c.categoryName === cat.name)
		])
	);

	// Fetch evaluation results (latest by control ID)
	const evalRes = await client.GET('/v1/evaluation/results', {
		params: {
			query: {
				'filter.auditScopeId': params.auditScopeId,
				latestByControlId: true,
				pageSize: 1000
			}
		}
	});

	const evaluationResults = evalRes.data?.results ?? [];

	// Index evaluation results by control ID
	const evaluationByControl: Record<string, SchemaEvaluationResult> = {};
	for (const result of evaluationResults) {
		const key = result.controlId ?? '';
		if (key) {
			evaluationByControl[key] = result;
		}
	}

	return { auditScope, catalog, controlsByCategory, evaluationResults, evaluationByControl };
}) satisfies PageLoad;