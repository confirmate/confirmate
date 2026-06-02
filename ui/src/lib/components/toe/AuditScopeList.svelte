<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { orchestratorClient } from '$lib/api/client';
	import type { SchemaAuditScope, SchemaCatalog } from '$lib/api/openapi/orchestrator';

	let {
		toeId,
		auditScopes,
		catalogs
	}: {
		toeId: string;
		auditScopes: SchemaAuditScope[];
		catalogs: SchemaCatalog[];
	} = $props();

	const enabledCatalogIds = $derived(new Set(auditScopes.map((s) => s.catalogId)));

	async function enable(catalog: SchemaCatalog) {
		const client = orchestratorClient();
		await client.POST('/v1/orchestrator/audit_scopes', {
			body: {
				id: '',
				name: catalog.name,
				targetOfEvaluationId: toeId,
				catalogId: catalog.id
			}
		});
		await invalidateAll();
	}

	async function disable(scope: SchemaAuditScope) {
		const client = orchestratorClient();
		await client.DELETE('/v1/orchestrator/audit_scopes/{auditScopeId}', {
			params: { path: { auditScopeId: scope.id } }
		});
		await invalidateAll();
	}
</script>

<div class="space-y-3">
	{#each catalogs as catalog}
		{@const scope = auditScopes.find((s) => s.catalogId === catalog.id)}
		{@const enabled = enabledCatalogIds.has(catalog.id)}
		<div class="flex items-center justify-between rounded-lg border border-gray-200 bg-white p-4">
			<div>
				<div class="font-medium text-gray-900">{catalog.name}</div>
				{#if catalog.description}
					<div class="mt-0.5 text-sm text-gray-500">{catalog.description}</div>
				{/if}
			</div>
			<div class="ml-4 flex shrink-0 items-center gap-3">
				{#if enabled}
					<span class="inline-flex items-center rounded-full bg-green-50 px-2.5 py-0.5 text-xs font-medium text-green-700">
						Enabled
					</span>
					<button
						onclick={() => disable(scope!)}
						class="rounded-md bg-red-50 px-3 py-1.5 text-sm font-medium text-red-700 hover:bg-red-100"
					>
						Disable
					</button>
				{:else}
					<span class="inline-flex items-center rounded-full bg-gray-100 px-2.5 py-0.5 text-xs font-medium text-gray-600">
						Disabled
					</span>
					<button
						onclick={() => enable(catalog)}
						class="rounded-md bg-confirmate px-3 py-1.5 text-sm font-medium text-white hover:bg-confirmate-light"
					>
						Enable
					</button>
				{/if}
			</div>
		</div>
	{/each}

	{#if catalogs.length === 0}
		<div class="rounded-lg border border-dashed border-gray-300 p-12 text-center">
			<p class="text-gray-500">No catalogs available.</p>
		</div>
	{/if}
</div>
