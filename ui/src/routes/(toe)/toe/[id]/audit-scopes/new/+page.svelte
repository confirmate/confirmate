<script lang="ts">
	import { goto, invalidate } from '$app/navigation';
	import { orchestratorClient } from '$lib/api/client';
	import CatalogPicker from '$lib/components/toe/CatalogPicker.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import type { SchemaCatalog } from '$lib/api/openapi/orchestrator';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	let selectedCatalog = $state<SchemaCatalog | null>(null);
	let name = $state('');
	let saving = $state(false);

	$effect(() => {
		if (selectedCatalog && !name) {
			name = selectedCatalog.name;
		}
	});

	async function create() {
		if (!selectedCatalog) return;
		saving = true;
		const client = orchestratorClient();
		await client.POST('/v1/orchestrator/audit_scopes', {
			body: {
				id: '',
				name,
				targetOfEvaluationId: data.toe.id!,
				catalogId: selectedCatalog.id
			}
		});
		await invalidate((url) => url.pathname === '/v1/orchestrator/audit_scopes');
		goto(`/toe/${data.toe.id}/audit-scopes/`);
	}
</script>

<div class="max-w-2xl">
	<div class="mb-8">
		<a href="/toe/{data.toe.id}/audit-scopes/" class="text-sm text-gray-500 hover:text-gray-700">
			← Back to Audit Scopes
		</a>
		<h2 class="mt-4 text-lg font-semibold text-gray-900">New Audit Scope</h2>
		<p class="mt-1 text-sm text-gray-500">Select a catalog and give this audit scope a name.</p>
	</div>

	<div class="space-y-6">
		<!-- Step 1: Catalog -->
		<div>
			<label class="mb-3 block text-sm font-medium text-gray-700">Catalog</label>
			<CatalogPicker catalogs={data.catalogs} bind:selected={selectedCatalog} />
		</div>

		<!-- Step 2: Name (shown once a catalog is selected) -->
		{#if selectedCatalog}
			<div>
				<label for="scope-name" class="block text-sm font-medium text-gray-700">Name</label>
				<input
					id="scope-name"
					type="text"
					bind:value={name}
					placeholder="e.g. ISO 27001 — Annual Audit 2025"
					class="mt-2 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm shadow-xs focus:border-confirmate focus:ring-1 focus:ring-confirmate focus:outline-none"
				/>
			</div>

			<div class="flex gap-3">
				<Button onclick={create} disabled={!name || saving}>
					{saving ? 'Creating…' : 'Create Audit Scope'}
				</Button>
				<Button variant="secondary" href="/toe/{data.toe.id}/audit-scopes/">Cancel</Button>
			</div>
		{/if}
	</div>
</div>
