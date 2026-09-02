import { orchestratorClient } from '$lib/api/client';
import { error } from '@sveltejs/kit';
import type { SchemaControl } from '$lib/api/openapi/orchestrator';
import {
	bucketControlsByCategory,
	countAssessmentsByMetric,
	indexAssessmentsById,
	indexControlsById,
	indexControlsInScope,
	indexEvaluationsByControl
} from '$lib/auditScope';
import type { PageLoad } from './$types';

export const load = (async ({ params, fetch, depends, url }) => {
	depends('evaluation:results');

	const client = orchestratorClient(fetch);

	const tab = url.searchParams.get('tab') as 'implementation' | 'compliance' | 'auditTrail' | null;

	const { data: auditScope, response } = await client.GET(
		'/v1/orchestrator/audit_scopes/{auditScopeId}',
		{ params: { path: { auditScopeId: params.auditScopeId } } }
	);

	if (!auditScope) throw error(response.status, response.statusText);

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
	// removed). Bucket the top-level controls using the catalog's nested
	// category/control structure.
	const controlsByCategory = bucketControlsByCategory(categories, topLevelControls);

	// Flat map of all controls (including sub-controls) by ID for audit trail lookups.
	const controlById = indexControlsById(topLevelControls);

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
	const evaluationByControl = indexEvaluationsByControl(evaluationResults);

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
	const assessmentById = indexAssessmentsById(assessmentResults);
	const assessmentCountByMetric = countAssessmentsByMetric(assessmentResults);

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
	const controlInScopeByControlId = indexControlsInScope(cisResp?.controlsInScope ?? []);

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