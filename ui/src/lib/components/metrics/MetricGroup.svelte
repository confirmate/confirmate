<script lang="ts">
	import type { SchemaMetric } from '$lib/api/openapi/orchestrator';
	import MetricRow from './MetricRow.svelte';
	import { Icon } from '@steeze-ui/svelte-icon';
	import { ChevronDown } from '@steeze-ui/heroicons';

	let {
		category,
		metrics,
		open = true,
		ontoggle
	}: {
		category: string;
		metrics: SchemaMetric[];
		open?: boolean;
		ontoggle?: () => void;
	} = $props();
</script>

<div class="overflow-hidden rounded-lg border border-gray-200 bg-white">
	<button
		type="button"
		onclick={ontoggle}
		class="flex w-full items-center gap-2.5 px-4 py-3 text-left hover:bg-gray-50"
	>
		<Icon
			src={ChevronDown}
			class="h-4 w-4 shrink-0 text-gray-400 transition-transform {open ? '' : '-rotate-90'}"
		/>
		<span class="font-medium text-gray-900">{category}</span>
		<span class="rounded-full bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-500">
			{metrics.length}
		</span>
	</button>

	{#if open}
		<div class="divide-y divide-gray-100 border-t border-gray-100">
			{#each metrics as metric}
				<MetricRow {metric} />
			{/each}
		</div>
	{/if}
</div>
