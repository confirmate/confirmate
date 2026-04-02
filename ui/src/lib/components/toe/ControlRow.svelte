<script lang="ts">
	import type { SchemaControl } from '$lib/api/openapi/orchestrator';

	let { control, depth = 0 }: { control: SchemaControl; depth?: number } = $props();
</script>

<div class="{depth > 0 ? 'ml-6 border-l border-gray-100 pl-4' : ''}">
	<div class="flex items-start gap-3 py-3">
		<span class="mt-0.5 shrink-0 rounded bg-gray-100 px-1.5 py-0.5 font-mono text-xs text-gray-500">
			{control.id}
		</span>
		<div class="min-w-0 flex-1">
			<div class="text-sm font-medium text-gray-900">{control.name}</div>
			{#if control.description}
				<div class="mt-0.5 text-sm text-gray-500">{control.description}</div>
			{/if}
		</div>
	</div>

	{#if control.controls?.length}
		<div class="mb-2">
			{#each control.controls as sub}
				<svelte:self control={sub} depth={depth + 1} />
			{/each}
		</div>
	{/if}
</div>
