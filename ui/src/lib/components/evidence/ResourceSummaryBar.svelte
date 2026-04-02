<script lang="ts">
	import type { SchemaResourceSnapshot as Resource } from '$lib/api/openapi/evidence';
	import type { SchemaAssessmentResult as AssessmentResult } from '$lib/api/openapi/orchestrator';
	import { Icon } from '@steeze-ui/svelte-icon';
	import { CheckCircle, XCircle, MinusCircle, XMark } from '@steeze-ui/heroicons';

	interface Props {
		resource: Resource;
		results: AssessmentResult[];
		onclose: () => void;
		ondetail: () => void;
	}
	let { resource, results, onclose, ondetail }: Props = $props();

	let resourceResults = $derived(results.filter((r) => r.resourceId === resource.id));

	function getResourceProp(res: Resource['resource'], prop: string): unknown {
		if (!res) return undefined;
		for (const val of Object.values(res)) {
			if (val && typeof val === 'object' && prop in val) return (val as Record<string, unknown>)[prop];
		}
		return undefined;
	}

	let shortName = $derived(
		(getResourceProp(resource.resource, 'name') as string | undefined) ??
		resource.id.split('/').pop() ??
		resource.id
	);
	let resourceType = $derived(resource.resourceType.split(',')[0].trim());
</script>

<div class="flex items-center gap-4 border-t border-gray-200 bg-white px-4 py-3">
	<!-- Identity -->
	<div class="min-w-0">
		<p class="truncate text-sm font-semibold text-gray-900">{shortName}</p>
		<p class="truncate font-mono text-xs text-gray-400">{resourceType} · {resource.id}</p>
	</div>

	<!-- Assessment badges -->
	<div class="flex shrink-0 flex-wrap gap-1.5">
		{#if resourceResults.length === 0}
			<span class="inline-flex items-center gap-1 text-xs text-gray-400">
				<Icon src={MinusCircle} class="h-3.5 w-3.5" />
				No results
			</span>
		{:else}
			{#each resourceResults as r}
				<span
					class="{r.compliant
						? 'bg-green-50 text-green-700 ring-green-200'
						: 'bg-red-50 text-red-700 ring-red-200'} inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium ring-1"
				>
					{#if r.compliant}
						<Icon src={CheckCircle} class="h-3 w-3" />
					{:else}
						<Icon src={XCircle} class="h-3 w-3" />
					{/if}
					{r.metricId}
				</span>
			{/each}
		{/if}
	</div>

	<!-- Actions -->
	<div class="ml-auto flex shrink-0 items-center gap-2">
		<button
			onclick={ondetail}
			class="rounded-md bg-confirmate px-3 py-1.5 text-xs font-medium text-white hover:bg-confirmate-light"
		>
			View evidence
		</button>
		<button onclick={onclose} class="text-gray-400 hover:text-gray-600">
			<Icon src={XMark} class="h-4 w-4" />
		</button>
	</div>
</div>
