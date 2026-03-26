import { listEvaluationResults } from '$lib/api/evaluation';
import { listAuditScopes } from '$lib/api/orchestrator';
import { error } from '@sveltejs/kit';
import type { LayoutLoad } from './$types';

export const load = (async ({ fetch, params, parent }) => {
	if (params.id == undefined) {
		error(405, 'Required parameter missing');
	}

	const targets = (await listAuditScopes(params.id, fetch)) || [];
	const data = await parent();

	const leftOverCatalogs = data.catalogs.filter((c) => {
		return targets.find((auditScope) => auditScope.catalogId == c.id) === undefined;
	});

	const enabledItems = data.catalogs.flatMap((catalog) => {
		const auditScope = targets.find((auditScope) => auditScope.catalogId == catalog.id);
		if (auditScope === undefined) {
			return [];
		}

		return [{ catalog, auditScope }];
	});

	// Retrieve the result of each "parent" (aka the top controls), because in the
	// overview we are only interested in that
	const topControlResults = await listEvaluationResults(
		{ targetOfEvaluationId: params.id, parentsOnly: true },
		true,
		fetch
	);

	return {
		targets,
		/**
		 * This array contains all catalogs which do not have a Target of Evaluation
		 * and are possible ready to be selected for evaluation.
		 */
		leftOverCatalogs,
		/**
		 * This array contains all currently selected catalogs and their Audit Scope.
		 */
		enabledItems,
		topControlResults
	};
}) satisfies LayoutLoad;
