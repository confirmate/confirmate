<script lang="ts">
	import type { SchemaCatalog } from '$lib/api/openapi/orchestrator';

	let {
		catalogs,
		selected = $bindable(null)
	}: {
		catalogs: SchemaCatalog[];
		selected?: SchemaCatalog | null;
	} = $props();
</script>

<div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
	{#each catalogs as catalog}
		{@const isSelected = selected?.id === catalog.id}
		<button
			type="button"
			onclick={() => (selected = catalog)}
			class="{isSelected
				? 'border-confirmate bg-blue-50 ring-1 ring-confirmate'
				: 'border-gray-200 bg-white hover:border-gray-300 hover:bg-gray-50'} rounded-lg border p-4 text-left transition-colors"
		>
			<div class="font-medium text-gray-900">{catalog.name}</div>
			{#if catalog.description}
				<div class="mt-1 text-sm text-gray-500 line-clamp-2">{catalog.description}</div>
			{/if}
		</button>
	{/each}
</div>
