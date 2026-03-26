<script lang="ts">
	import { invalidate } from '$app/navigation';
	import {
		stopEvaluation,
		type ComplianceStatus,
		type EvaluationResult
	} from '$lib/api/evaluation';
	import { removeAuditScope, type AuditScope } from '$lib/api/orchestrator';
	import CatalogComplianceItem from '$lib/components/CatalogComplianceItem.svelte';
	import EnableCatalogButton from '$lib/components/EnableCatalogButton.svelte';
	import type { PageData } from './$types';

	interface Props {
		data: PageData;
	}

	let { data }: Props = $props();

	function buildCompliance(
		evaluations: EvaluationResult[]
	): Map<string, Map<string, ComplianceStatus>> {
		let all = new Map();
		let compliance: Map<string, ComplianceStatus>;
		for (let result of evaluations) {
			compliance = all.get(result.controlCatalogId);
			if (compliance === undefined) {
				compliance = new Map();
				all.set(result.controlCatalogId, compliance);
			}

			compliance.set(result.controlId, result.status);
		}

		return all;
	}

	async function remove(detail: { auditScope: AuditScope }) {
		let really = confirm('Do you really want to remove this Target of Evaluation?');

		if (!really) {
			return;
		}

		await stopEvaluation(detail.auditScope);
		await removeAuditScope(detail.auditScope);

		// refresh our ToEs
		invalidate((url) => url.pathname == `/v1/orchestrator/audit_scopes`);
	}

	// TODO: This should be done in the backend
	let compliance = $derived(buildCompliance(data.topControlResults));
</script>

<ul class="grid grid-cols-1 gap-x-6 gap-y-8 lg:grid-cols-3 xl:gap-x-8">
	{#each data.enabledItems as item (item.catalog.id)}
		<CatalogComplianceItem
			{...item}
			onremove={remove}
			compliance={compliance.get(item.catalog.id) ?? new Map()}
		/>
	{/each}
	{#if data.leftOverCatalogs.length > 0}
		<li>
			<a href="/cloud/{data.service.id}/compliance/new">
				<EnableCatalogButton />
			</a>
		</li>
	{/if}
</ul>
