<script lang="ts">
	import type { SchemaCategory, SchemaControl, SchemaEvaluationResult } from '$lib/api/openapi/orchestrator';
	import ControlRow from './ControlRow.svelte';
	import { ChevronDown } from '@steeze-ui/heroicons';
	import { Icon } from '@steeze-ui/svelte-icon';

	let {
		category,
		controls,
		evaluationByControl = {},
		assessmentCountByMetric = {}
	}: {
		category: SchemaCategory;
		controls: SchemaControl[];
		evaluationByControl?: Record<string, SchemaEvaluationResult>;
		assessmentCountByMetric?: Record<string, number>;
	} = $props();

	let open = $state(true);
</script>

<div class="overflow-hidden rounded-lg border border-gray-200 bg-white">
	<button
		type="button"
		onclick={() => (open = !open)}
		class="flex w-full items-center justify-between px-4 py-3 text-left hover:bg-gray-50"
	>
		<div class="flex items-center gap-2.5">
			<Icon
				src={ChevronDown}
				class="h-4 w-4 shrink-0 text-gray-400 transition-transform {open ? '' : '-rotate-90'}"
			/>
			<span class="font-medium text-gray-900">{category.name}</span>
			<span class="rounded-full bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-500">
				{controls.length}
			</span>
		</div>
	</button>

	{#if open}
		<div class="divide-y divide-gray-100 border-t border-gray-100 px-4">
			{#each controls as control}
				<ControlRow {control} evaluation={evaluationByControl[control.id ?? '']} {evaluationByControl} {assessmentCountByMetric} />
			{/each}
		</div>
	{/if}
</div>