import { evidenceClient, orchestratorClient } from '$lib/api/client';
import { error } from '@sveltejs/kit';
import type { PageLoad } from './$types';

export const load = (async ({ params, fetch }) => {
	const client = orchestratorClient(fetch);

	const [{ data: control, response }, { data: auditScope }, { data: usersResp }] =
		await Promise.all([
			client.GET('/v1/orchestrator/controls/{controlId}', {
				params: { path: { controlId: params.controlId } }
			}),
			client.GET('/v1/orchestrator/audit_scopes/{auditScopeId}', {
				params: { path: { auditScopeId: params.auditScopeId } }
			}),
			client.GET('/v1/users', {})
		]);

	if (!control) throw error(response.status, response.statusText);

	// Find the ControlInScope record for this control in this audit scope
	const { data: cisResp } = await client.GET('/v1/orchestrator/controls_in_scope', {
		params: {
			query: {
				'filter.auditScopeId': params.auditScopeId,
				'filter.controlId': params.controlId,
				pageSize: 1
			}
		}
	});
	const cis = cisResp?.controlsInScope?.[0] ?? null;

	// Evaluation result for this control
	const { data: evalResp } = await client.GET('/v1/orchestrator/evaluation_results', {
		params: {
			query: {
				'filter.targetOfEvaluationId': params.id,
				latestByControlId: true,
				pageSize: 1000
			}
		}
	});
	const evaluation = evalResp?.results?.find((r) => r.controlId === params.controlId) ?? null;

	// Latest assessment results of the control's metrics (automatic and manual)
	const metricIds = (control.metrics ?? []).map((m) => m.id ?? '').filter(Boolean);
	const { data: assessmentResp } = metricIds.length
		? await client.GET('/v1/orchestrator/assessment_results', {
				params: {
					query: {
						'filter.targetOfEvaluationId': params.id,
						'filter.metricIds': metricIds,
						latestByResourceId: true,
						pageSize: 1000
					}
				}
			})
		: { data: null };
	const assessmentResults = assessmentResp?.results ?? [];

	// Resources of the target of evaluation, a manual metric result can be about one of them
	const { data: resourcesResp } = await evidenceClient(fetch).GET('/v1/evidence_store/resources', {
		params: {
			query: {
				'filter.targetOfEvaluationId': params.id,
				pageSize: 1500,
				orderBy: 'resource_type',
				asc: true
			}
		}
	});
	const resources = resourcesResp?.results ?? [];

	// Audit trail filtered to this control's CIS record
	const { data: auditTrailResp } = cis?.id
		? await client.GET('/v1/orchestrator/audit_trail_events', {
				params: {
					query: {
						'filter.controlInScopeId': cis.id,
						pageSize: 200,
						orderBy: 'created_at',
						asc: false
					}
				}
			})
		: { data: null };
	const auditTrailEvents = auditTrailResp?.auditTrailEvents ?? [];

	return {
		control,
		auditScope: auditScope ?? null,
		cis,
		evaluation,
		assessmentResults,
		resources,
		auditTrailEvents,
		users: usersResp?.users ?? []
	};
}) satisfies PageLoad;
