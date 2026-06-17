<script lang="ts">
	import type { SchemaResourceSnapshot as Resource } from '$lib/api/openapi/evidence';
	import type { SchemaAssessmentResult as AssessmentResult } from '$lib/api/openapi/orchestrator';
	import { XMark, CheckCircle, XCircle } from '@steeze-ui/heroicons';
	import { Icon } from '@steeze-ui/svelte-icon';

	interface Props {
		resource: Resource;
		results: AssessmentResult[];
		onclose: () => void;
	}

	let { resource, results, onclose }: Props = $props();

	let tab = $state<'results' | 'properties'>('results');

	function shortName(id: string | undefined) {
		if (!id) return '—';
		const parts = id.split('/');
		return parts[parts.length - 1];
	}

	function flatProperties(obj: Record<string, unknown>, prefix = ''): [string, unknown][] {
		const out: [string, unknown][] = [];
		for (const [k, v] of Object.entries(obj)) {
			const key = prefix ? `${prefix}.${k}` : k;
			if (k === 'raw' || k === 'labels') continue;
			if (v === null || v === undefined || v === '' || v === 0) continue;
			if (typeof v === 'object' && !Array.isArray(v)) {
				out.push(...flatProperties(v as Record<string, unknown>, key));
			} else {
				out.push([key, v]);
			}
		}
		return out;
	}

	// Extract concrete resource data from the ontology union
	let concreteResource = $derived(() => {
		if (!resource.resource) return {};
		for (const val of Object.values(resource.resource)) {
			if (val && typeof val === 'object') return val as Record<string, unknown>;
		}
		return {};
	});

	let properties = $derived(flatProperties(concreteResource()));
	let resourceType = $derived(resource.resourceType.split(',')[0].trim());
</script>

<div class="flex h-full flex-col overflow-hidden rounded-xl border border-gray-200 bg-white shadow-xl">
	<!-- Header -->
	<div class="bg-blue-600 px-4 py-5">
		<div class="flex items-start justify-between">
			<div class="min-w-0">
				<p class="truncate text-base font-semibold text-white">{shortName(resource.id)}</p>
				<p class="mt-1 text-sm text-blue-200">
					{resourceType} &mdash; {results.length} assessment result{results.length !== 1 ? 's' : ''}
				</p>
			</div>
			<button onclick={onclose} class="ml-3 text-blue-200 hover:text-white">
				<Icon src={XMark} class="h-5 w-5" />
			</button>
		</div>
	</div>

	<!-- Tabs -->
	<div class="border-b border-gray-200 px-4">
		<nav class="-mb-px flex space-x-6">
			{#each [{ id: 'results', label: 'Assessment Results', count: results.length }, { id: 'properties', label: 'Properties', count: 0 }] as t}
				<button
					onclick={() => (tab = t.id as 'results' | 'properties')}
					class="{tab === t.id
						? 'border-blue-600 text-blue-600'
						: 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'} flex items-center gap-x-1.5 border-b-2 py-3 text-sm font-medium whitespace-nowrap"
				>
					{t.label}
					{#if t.count > 0}
						<span
							class="{tab === t.id
								? 'bg-blue-100 text-blue-600'
								: 'bg-gray-100 text-gray-600'} rounded-full px-2 py-0.5 text-xs font-medium"
						>
							{t.count}
						</span>
					{/if}
				</button>
			{/each}
		</nav>
	</div>

	<!-- Content -->
	<div class="flex-1 overflow-y-auto px-4 py-4">
		{#if tab === 'results'}
			{#if results.length === 0}
				<p class="text-sm text-gray-400">No assessment results for this resource.</p>
			{:else}
				<ul class="space-y-3">
					{#each results as r}
						<li class="flex items-start gap-x-3">
							{#if r.compliant}
								<Icon src={CheckCircle} class="mt-0.5 h-5 w-5 shrink-0 text-green-600" />
							{:else}
								<Icon src={XCircle} class="mt-0.5 h-5 w-5 shrink-0 text-red-600" />
							{/if}
							<div class="min-w-0">
								<p class="text-sm font-medium text-gray-900">{r.metricId}</p>
							</div>
						</li>
					{/each}
				</ul>
			{/if}
		{:else}
			{#if properties.length === 0}
				<p class="text-sm text-gray-400">No properties.</p>
			{:else}
				<dl class="space-y-2">
					{#each properties as [k, v]}
						<div>
							<dt class="text-xs font-medium text-gray-500">{k}</dt>
							<dd class="mt-0.5 truncate text-sm text-gray-900">{v}</dd>
						</div>
					{/each}
				</dl>
			{/if}
		{/if}
	</div>
</div>
