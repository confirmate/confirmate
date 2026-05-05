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
	const categories = catalog?.categories ?? [];

	// Fetch all controls for this catalog (without pagination to preserve hierarchy)
	const { data: allControlsResp } = await client.GET('/v1/orchestrator/controls', {
		params: {
			query: {
				catalog_id: auditScope.catalogId,
				pageSize: 1000
			}
		}
	});
	const allControls = allControlsResp?.controls ?? [];

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
				'filter.targetOfEvaluationId': params.id,
				latestByControlId: true,
				pageSize: 1000
			}
		}
	});

	// Fetch ALL evaluation results for history
	const evalResAll = await client.GET('/v1/evaluation/results', {
		params: {
			query: {
				'filter.targetOfEvaluationId': params.id,
				pageSize: 1000
			}
		}
	});

	const evaluationResults = evalRes.data?.results ?? [];
	const allEvaluationResults = evalResAll.data?.results ?? [];

	// Index evaluation results by control ID
	const evaluationByControl: Record<string, SchemaEvaluationResult> = {};
	for (const result of evaluationResults) {
		const key = result.controlId ?? '';
		if (key) {
			evaluationByControl[key] = result;
		}
	}

	// Fetch assessment results and count by metric ID
	const assessmentRes = await client.GET('/v1/orchestrator/assessment_results', {
		params: {
			query: {
				'filter.targetOfEvaluationId': params.id,
				pageSize: 1000
			}
		}
	});

	const assessmentResults = assessmentRes.data?.results ?? [];

	// Map assessment results by ID for lookup
	const assessmentById: Record<string, { metricId?: string; compliant?: boolean; timestamp?: string }> = {};
	for (const ar of assessmentResults) {
		if (ar.id) {
			assessmentById[ar.id] = { metricId: ar.metricId, compliant: ar.compliant, timestamp: ar.timestamp };
		}
	}

	// Count assessment results by metric name - separate passing and failing
	const assessmentCountByMetric: Record<string, { passing: number; failing: number }> = {};
	for (const ar of assessmentResults) {
		const metricName = ar.metricId;
		if (metricName) {
			if (!assessmentCountByMetric[metricName]) {
				assessmentCountByMetric[metricName] = { passing: 0, failing: 0 };
			}
			if (ar.compliant) {
				assessmentCountByMetric[metricName].passing++;
			} else {
				assessmentCountByMetric[metricName].failing++;
			}
		}
	}

	return { auditScope, catalog, controlsByCategory, evaluationResults, allEvaluationResults, evaluationByControl, assessmentCountByMetric, assessmentById };
}) satisfies PageLoad;