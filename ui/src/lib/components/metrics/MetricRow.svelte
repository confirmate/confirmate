<script lang="ts">
	import type { SchemaMetric } from '$lib/api/openapi/orchestrator';
	import { Icon } from '@steeze-ui/svelte-icon';
	import { CheckCircle, XCircle } from '@steeze-ui/heroicons';

	let { metric }: { metric: SchemaMetric } = $props();
</script>

<div class="flex items-start gap-4 px-4 py-3">
	<div class="min-w-0 flex-1">
		<div class="flex items-center gap-2">
			<span class="text-sm font-medium text-gray-900">{metric.name}</span>
			{#if metric.deprecatedSince}
				<span class="rounded-full bg-yellow-50 px-1.5 py-0.5 text-xs font-medium text-yellow-700">
					deprecated
				</span>
			{/if}
		</div>
		{#if metric.description}
			<p class="mt-0.5 text-sm text-gray-500">{metric.description}</p>
		{/if}
		{#if metric.comments}
			<p class="mt-1 text-xs italic text-gray-400">{metric.comments}</p>
		{/if}
	</div>
	<div class="shrink-0 text-right">
		{#if metric.implementation}
			<span class="inline-flex items-center gap-1 rounded-full bg-green-50 px-2 py-0.5 text-xs font-medium text-green-700">
				<Icon src={CheckCircle} class="h-3 w-3" />
				automated
			</span>
		{:else}
			<span class="inline-flex items-center gap-1 rounded-full bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-400">
				<Icon src={XCircle} class="h-3 w-3" />
				manual
			</span>
		{/if}
		{#if metric.version}
			<div class="mt-1 font-mono text-xs text-gray-300">v{metric.version}</div>
		{/if}
	</div>
</div>
