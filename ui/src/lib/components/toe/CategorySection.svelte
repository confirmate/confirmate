<script lang="ts">
	import type { SchemaCategory, SchemaControl } from '$lib/api/openapi/orchestrator';
	import ControlRow from './ControlRow.svelte';

	let {
		category,
		controls
	}: {
		category: SchemaCategory;
		controls: SchemaControl[];
	} = $props();

	let open = $state(true);
</script>

<div class="rounded-lg border border-gray-200 bg-white">
	<button
		type="button"
		onclick={() => (open = !open)}
		class="flex w-full items-center justify-between px-4 py-3 text-left"
	>
		<div>
			<span class="font-medium text-gray-900">{category.name}</span>
			<span class="ml-2 text-sm text-gray-400">{controls.length} controls</span>
		</div>
		<svg
			class="h-4 w-4 shrink-0 text-gray-400 transition-transform {open ? 'rotate-180' : ''}"
			fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"
		>
			<path stroke-linecap="round" stroke-linejoin="round" d="M19 9l-7 7-7-7" />
		</svg>
	</button>

	{#if open}
		<div class="divide-y divide-gray-100 border-t border-gray-100 px-4">
			{#each controls as control}
				<ControlRow {control} />
			{/each}
		</div>
	{/if}
</div>
