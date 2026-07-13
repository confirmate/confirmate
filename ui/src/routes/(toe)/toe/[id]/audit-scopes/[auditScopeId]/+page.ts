import { orchestratorClient } from '$lib/api/client';
import { error } from '@sveltejs/kit';
import type { SchemaControl, SchemaControlInScope, SchemaEvaluationResult } from '$lib/api/openapi/orchestrator';
import type { PageLoad } from './$types';

export const load = (async ({ params, fetch, depends, url }) => {
	depends('evaluation:results');

	const client = orchestratorClient(fetch);

	const tab = url.searchParams.get('tab') as 'implementation' | 'compliance' | 'auditTrail' | null;

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
				'filter.catalogId': auditScope.catalogId,
				pageSize: 1000
			}
		}
	});
	// The server returns only top-level controls (parent_control_id IS NULL)
	// with children already preloaded via Controls.Metrics. Sort them here.
	const topLevelControls: SchemaControl[] = (allControlsResp?.controls ?? [])
		.slice()
		.sort((a, b) => (a.id ?? '').localeCompare(b.id ?? ''));

	// Control no longer carries its category back-reference (categoryName was
	// removed). Build a category-by-control-id lookup from the catalog's
	// nested controls and use that to bucket the top-level controls.
	const categoryByControlId = new Map<string, string>();
	for (const cat of categories) {
		for (const c of cat.controls ?? []) {
			if (c.id) categoryByControlId.set(c.id, cat.name);
		}
	}
	const controlsByCategory: Record<string, SchemaControl[]> = Object.fromEntries(
		categories.map((cat) => [
			cat.name,
			topLevelControls.filter((c) => categoryByControlId.get(c.id ?? '') === cat.name)
		])
	);

	// Flat map of all controls (including sub-controls) by ID for audit trail lookups.
	function flatControls(controls: SchemaControl[]): SchemaControl[] {
		return controls.flatMap((c) => [c, ...flatControls(c.controls ?? [])]);
	}
	const controlById: Record<string, { shortName?: string }> = Object.fromEntries(
		flatControls(topLevelControls).filter((c) => c.id).map((c) => [c.id!, { shortName: c.shortName }])
	);

	// Fetch evaluation results (latest by control ID), filtered by audit scope
	const evalRes = await client.GET('/v1/orchestrator/evaluation_results', {
		params: {
			query: {
				'filter.targetOfEvaluationId': params.id,
				'filter.auditScopeId': params.auditScopeId,
				latestByControlId: true,
				pageSize: 1000
			}
		}
	});

	// Fetch ALL evaluation results for history, filtered by audit scope
	const evalResAll = await client.GET('/v1/orchestrator/evaluation_results', {
		params: {
			query: {
				'filter.targetOfEvaluationId': params.id,
				'filter.auditScopeId': params.auditScopeId,
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
	const assessmentById: Record<string, { metricId?: string; compliant?: boolean; createdAt?: string }> = {};
	for (const ar of assessmentResults) {
		if (ar.id) {
			assessmentById[ar.id] = { metricId: ar.metricId, compliant: ar.compliant, createdAt: ar.createdAt };
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

	// Fetch the ControlInScope records for this audit scope. They are
	// auto-created when the audit scope is created, but a user can opt
	// individual controls out (deleting the record) or back in.
	const { data: cisResp } = await client.GET('/v1/orchestrator/controls_in_scope', {
		params: {
			query: {
				'filter.auditScopeId': params.auditScopeId,
				pageSize: 1000
			}
		}
	});
	const controlInScopeByControlId: Record<string, SchemaControlInScope> = {};
	for (const cis of cisResp?.controlsInScope ?? []) {
		if (cis.controlId) controlInScopeByControlId[cis.controlId] = cis;
	}

	const { data: usersResp } = await client.GET('/v1/users', {});
	const users = usersResp?.users ?? [];

	const { data: auditTrailResp } = await client.GET('/v1/orchestrator/audit_trail_events', {
		params: {
			query: {
				'filter.auditScopeId': params.auditScopeId,
				pageSize: 500,
				orderBy: 'created_at',
				asc: false
			}
		}
	});
	const auditTrailEvents = auditTrailResp?.auditTrailEvents ?? [];

	// Fetch the current user's permission on this audit scope so the UI
	// can show/hide admin-only actions like "Set out of scope".
	const { data: permResp } = await client.GET('/v1/users/permissions/{objectType}/{objectId}', {
		params: {
			path: {
				objectType: 'OBJECT_TYPE_AUDIT_SCOPE',
				objectId: params.auditScopeId
			}
		}
	});
	const userPermissions = permResp?.userPermissions ?? [];

	return {
		auditScope,
		catalog,
		controlsByCategory,
		controlById,
		evaluationResults,
		allEvaluationResults,
		evaluationByControl,
		assessmentCountByMetric,
		assessmentById,
		controlInScopeByControlId,
		users,
		auditTrailEvents,
		userPermissions,
		initialTab: tab ?? 'implementation'
	};
}) satisfies PageLoad;