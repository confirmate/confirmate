import { orchestratorClient } from '$lib/api/client';
import { error } from '@sveltejs/kit';
import type { LayoutLoad } from './$types';

export const load = (async ({ params, fetch }) => {
	const client = orchestratorClient(fetch);

	const [toeRes, auditScopesRes] = await Promise.all([
		client.GET('/v1/orchestrator/targets_of_evaluation/{targetOfEvaluationId}', {
			params: { path: { targetOfEvaluationId: params.id } }
		}),
		client.GET('/v1/orchestrator/audit_scopes', {
			params: { query: { 'filter.targetOfEvaluationId': params.id } }
		})
	]);

	if (!toeRes.data) error(toeRes.response.status, toeRes.response.statusText);

	return {
		toe: toeRes.data,
		auditScopes: auditScopesRes.data?.auditScopes ?? []
	};
}) satisfies LayoutLoad;
