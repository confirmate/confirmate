import { orchestratorClient } from '$lib/api/client';
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

	if (!control) error(response.status, response.statusText);

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
		auditTrailEvents,
		users: usersResp?.users ?? []
	};
}) satisfies PageLoad;
