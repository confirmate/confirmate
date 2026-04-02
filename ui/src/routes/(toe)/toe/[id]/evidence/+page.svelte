<script lang="ts">
	import type { PageData } from './$types';
	import type { SchemaResourceSnapshot as Resource } from '$lib/api/openapi/evidence';
	import SectionHeader from '$lib/components/ui/SectionHeader.svelte';
	import ResourceOverview from '$lib/components/evidence/ResourceOverview.svelte';
	import ResourceColumns from '$lib/components/evidence/ResourceColumns.svelte';
	import ResourceDetail from '$lib/components/evidence/ResourceDetail.svelte';
	import ResourceSummaryBar from '$lib/components/evidence/ResourceSummaryBar.svelte';
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

	// Resource selected for graph summary bar / columns
	let graphSelected = $state<Resource | null>(null);
	let columnsSelected = $state<Resource | null>(null);

	// Full evidence detail drawer — only opened explicitly via "View evidence"
	let detailResource = $state<Resource | null>(null);
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

	<div class="mt-6 flex min-h-0 flex-1 flex-col">

		<!-- LIST: table + side detail drawer -->
		{#if view === 'list'}
			<div class="flex min-h-0 gap-6">
				<div class="min-w-0 flex-1">
					<ResourceOverview
						resources={data.resources}
						results={data.results}
						onselect={(r) => (detailResource = r)}
					/>
				</div>
				{#if detailResource}
					<div class="w-96 shrink-0 overflow-hidden rounded-xl border border-gray-200 shadow-lg" style="max-height: calc(100vh - 12rem);">
						<ResourceDetail
							resource={detailResource}
							results={data.results}
							{toeId}
							onclose={() => (detailResource = null)}
						/>
					</div>
				{/if}
			</div>

		<!-- COLUMNS: miller columns above, detail below -->
		{:else if view === 'columns'}
			<div class="flex min-h-0 flex-1 flex-col overflow-hidden rounded-xl border border-gray-200">
				<div class="min-h-0 flex-1 overflow-hidden">
					<ResourceColumns
						resources={data.resources}
						edges={data.edges}
						results={data.results}
						onselect={(r) => { columnsSelected = r; detailResource = null; }}
					/>
				</div>
				{#if columnsSelected}
					<ResourceDetail
						resource={columnsSelected}
						results={data.results}
						{toeId}
						onclose={() => (columnsSelected = null)}
						horizontal
					/>
				{/if}
			</div>

		<!-- GRAPH: graph above, summary bar + optional detail below -->
		{:else}
			<div class="flex min-h-0 flex-1 flex-col overflow-hidden rounded-xl border border-gray-200">
				<div class="min-h-0 flex-1">
					<ResourceGraph
						resources={data.resources}
						edges={data.edges}
						results={data.results}
						selectedId={graphSelected?.id ?? null}
						onresourceselect={(r) => { graphSelected = r; detailResource = null; }}
					/>
				</div>
				{#if graphSelected}
					<ResourceSummaryBar
						resource={graphSelected}
						results={data.results}
						onclose={() => (graphSelected = null)}
						ondetail={() => (detailResource = graphSelected)}
					/>
				{/if}
			</div>

			<!-- Full evidence detail — opens below when "View evidence" is clicked -->
			{#if detailResource}
				<div class="mt-4 overflow-hidden rounded-xl border border-gray-200 shadow-sm" style="max-height: 28rem;">
					<ResourceDetail
						resource={detailResource}
						results={data.results}
						{toeId}
						onclose={() => (detailResource = null)}
						horizontal
					/>
				</div>
			{/if}
		{/if}

	</div>
</div>
