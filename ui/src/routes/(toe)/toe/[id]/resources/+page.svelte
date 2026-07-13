<script lang="ts">
	import type { PageData } from './$types';
	import type { SchemaResourceSnapshot as Resource } from '$lib/api/openapi/evidence';
	import SectionHeader from '$lib/components/ui/SectionHeader.svelte';
	import ResourceOverview from '$lib/components/evidence/ResourceOverview.svelte';
	import NodeDetail from '$lib/components/graph/NodeDetail.svelte';
	import ResourceGraph from '$lib/components/graph/ResourceGraph.svelte';
	import ResourceDetailModal from '$lib/components/graph/ResourceDetailModal.svelte';
	import { Icon } from '@steeze-ui/svelte-icon';
	import { Share, TableCells } from '@steeze-ui/heroicons';

	interface Props {
		data: PageData;
	}
	let { data }: Props = $props();

	let view = $state<'graph' | 'list'>('graph');

	// Selected resource — shown in side panel
	let selected = $state<Resource | null>(null);
	let detailResource = $state<Resource | null>(null);
</script>

<div class="flex min-h-0 flex-1 flex-col">
	<SectionHeader
		title="Resources"
		description="Discovered resources, evidence, and assessment results."
	>
		{#snippet actions()}
			<div class="inline-flex rounded-lg border border-gray-200 bg-white p-1 shadow-sm">
				<button
					onclick={() => (view = 'graph')}
					class="{view === 'graph'
						? 'bg-gray-100 text-gray-900'
						: 'text-gray-500 hover:text-gray-700'} flex items-center gap-1.5 rounded-md px-3 py-1.5 text-sm font-medium transition-colors"
				>
					<Icon src={Share} class="h-4 w-4" />
					Graph
				</button>
				<button
					onclick={() => (view = 'list')}
					class="{view === 'list'
						? 'bg-gray-100 text-gray-900'
						: 'text-gray-500 hover:text-gray-700'} flex items-center gap-1.5 rounded-md px-3 py-1.5 text-sm font-medium transition-colors"
				>
					<Icon src={TableCells} class="h-4 w-4" />
					List
				</button>
			</div>
		{/snippet}
	</SectionHeader>

	<div class="mt-4 flex min-h-0 flex-1 gap-4">
		<!-- Main content area -->
		<div class="min-w-0 flex-1">
			{#if view === 'graph'}
				<div class="flex min-h-0 flex-1 flex-col overflow-hidden rounded-xl border border-gray-200" style="height: calc(100vh - 12rem);">
					<ResourceGraph
						resources={data.resources}
						edges={data.edges}
						results={data.results}
						selectedId={selected?.id ?? null}
						onresourceselect={(r) => { selected = r; detailResource = null; }}
					/>
				</div>
			{:else}
				<ResourceOverview
					resources={data.resources}
					results={data.results}
					onselect={(r) => { selected = r; }}
				/>
			{/if}
		</div>

		<!-- Side panel — shown when a resource is selected -->
		{#if selected}
			<div class="w-80 shrink-0" style="max-height: calc(100vh - 12rem);">
				<NodeDetail
					resource={selected}
					results={data.results.filter((r) => r.resourceId === selected?.id)}
					onclose={() => { selected = null; detailResource = null; }}
					ondetail={(r) => { detailResource = r; }}
				/>
			</div>
		{/if}
	</div>
</div>

<ResourceDetailModal
	resource={detailResource}
	onclose={() => { detailResource = null; }}
/>
