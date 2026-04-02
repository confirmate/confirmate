<script lang="ts">
	import type { PageData } from './$types';
	import type { SchemaResourceSnapshot as Resource } from '$lib/api/openapi/evidence';
	import SectionHeader from '$lib/components/ui/SectionHeader.svelte';
	import ResourceOverview from '$lib/components/evidence/ResourceOverview.svelte';
	import ResourceColumns from '$lib/components/evidence/ResourceColumns.svelte';
	import ResourceDetail from '$lib/components/evidence/ResourceDetail.svelte';
	import ResourceGraph from '$lib/components/graph/ResourceGraph.svelte';
	import { Icon } from '@steeze-ui/svelte-icon';
	import { Share, TableCells, Bars3BottomLeft } from '@steeze-ui/heroicons';
	import { page } from '$app/state';

	interface Props {
		data: PageData;
	}
	let { data }: Props = $props();

	const toeId = page.params.id ?? '';

	let view = $state<'list' | 'columns' | 'graph'>('graph');
	let selected = $state<Resource | null>(null);
</script>

<div class="flex min-h-0 flex-1 flex-col">
	<SectionHeader
		title="Evidence"
		description="Resources, evidence, and assessment results for this target."
	>
		{#snippet actions()}
			<div class="inline-flex rounded-lg border border-gray-200 bg-white p-1 shadow-sm">
				<button
					onclick={() => (view = 'list')}
					class="{view === 'list'
						? 'bg-gray-100 text-gray-900'
						: 'text-gray-500 hover:text-gray-700'} flex items-center gap-1.5 rounded-md px-3 py-1.5 text-sm font-medium transition-colors"
				>
					<Icon src={TableCells} class="h-4 w-4" />
					Resources
				</button>
				<button
					onclick={() => (view = 'columns')}
					class="{view === 'columns'
						? 'bg-gray-100 text-gray-900'
						: 'text-gray-500 hover:text-gray-700'} flex items-center gap-1.5 rounded-md px-3 py-1.5 text-sm font-medium transition-colors"
				>
					<Icon src={Bars3BottomLeft} class="h-4 w-4" />
					Explore
				</button>
				<button
					onclick={() => (view = 'graph')}
					class="{view === 'graph'
						? 'bg-gray-100 text-gray-900'
						: 'text-gray-500 hover:text-gray-700'} flex items-center gap-1.5 rounded-md px-3 py-1.5 text-sm font-medium transition-colors"
				>
					<Icon src={Share} class="h-4 w-4" />
					Graph
				</button>
			</div>
		{/snippet}
	</SectionHeader>

	<!-- Main area + slide-over -->
	<div class="relative mt-6 flex min-h-0 gap-6">
		<!-- Content -->
		<div class="min-w-0 flex-1 overflow-hidden {view === 'columns' ? 'flex flex-col' : ''}">
			{#if view === 'list'}
				<ResourceOverview
					resources={data.resources}
					results={data.results}
					onselect={(r) => (selected = r)}
				/>
			{:else if view === 'columns'}
				<ResourceColumns
					resources={data.resources}
					edges={data.edges}
					results={data.results}
					onselect={(r) => (selected = r)}
				/>
			{:else}
				<ResourceGraph
					resources={data.resources}
					edges={data.edges}
					results={data.results}
					selectedId={selected?.id ?? null}
					onresourceselect={(r) => (selected = r)}
				/>
			{/if}
		</div>

		<!-- Resource detail slide-over -->
		{#if selected}
			<div class="w-96 shrink-0 overflow-hidden rounded-xl border border-gray-200 shadow-lg" style="max-height: calc(100vh - 12rem);">
				<ResourceDetail
					resource={selected}
					results={data.results}
					{toeId}
					onclose={() => (selected = null)}
				/>
			</div>
		{/if}
	</div>
</div>
