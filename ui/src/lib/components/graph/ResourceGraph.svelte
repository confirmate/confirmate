<script lang="ts">
	import type { NodeDefinition, EdgeDefinition } from 'cytoscape';
	import type { SchemaResourceSnapshot as Resource, SchemaGraphEdge as GraphEdge } from '$lib/api/openapi/evidence';
	import type { SchemaAssessmentResult as AssessmentResult } from '$lib/api/openapi/orchestrator';
	import DiscoveryGraph from './DiscoveryGraph.svelte';
	import { Icon } from '@steeze-ui/svelte-icon';
	import { ViewfinderCircle } from '@steeze-ui/heroicons';

	interface Props {
		resources: Resource[];
		edges: GraphEdge[];
		results: AssessmentResult[];
		selectedId?: string | null;
		onresourceselect?: (resource: Resource) => void;
	}
	let { resources, edges, results, selectedId: externalSelectedId = null, onresourceselect }: Props = $props();

	const STATUS_WAITING = 0;
	const STATUS_GOOD = 1;
	const STATUS_BAD = 2;

	let internalSelectedId = $state<string | null>(null);
	let overlay = $state(false);
	let graphComponent = $state<DiscoveryGraph | null>(null);

	$effect(() => {
		if (externalSelectedId != null) internalSelectedId = externalSelectedId;
	});

	// Extract name from ontology union resource
	function resourceLabel(r: Resource): string {
		if (r.resource) {
			for (const val of Object.values(r.resource)) {
				if (val && typeof val === 'object' && 'name' in val && val.name) {
					return String(val.name);
				}
			}
		}
		return r.id.split('/').pop() ?? r.id;
	}

	let nodes = $derived<NodeDefinition[]>(
		resources.map((r) => {
			const rResults = results.filter((res) => res.resourceId === r.id);
			let status = STATUS_WAITING;
			if (rResults.length > 0) {
				status = rResults.some((res) => !res.compliant) ? STATUS_BAD : STATUS_GOOD;
			}
			const typeEntries: Record<string, boolean> = {};
			for (const t of r.resourceType.split(',')) {
				typeEntries[`type.${t.trim()}`] = true;
			}
			return {
				group: 'nodes',
				data: {
					id: r.id,
					status,
					label: resourceLabel(r),
					...typeEntries
				}
			};
		})
	);

	let cytoscapeEdges = $derived<EdgeDefinition[]>(
		edges.map((e) => ({ group: 'edges', data: { id: e.id, source: e.source, target: e.target } }))
	);

	function onselect(node: NodeDefinition | null) {
		const id = node?.data?.id ?? null;
		internalSelectedId = id;
		if (id) {
			const resource = resources.find((r) => r.id === id);
			if (resource) onresourceselect?.(resource);
		} else {
			onresourceselect?.(undefined as unknown as Resource);
		}
	}
</script>

<div class="relative overflow-hidden rounded-xl border border-gray-200">
	<div class="flex items-center justify-between border-b border-gray-100 bg-gray-50 px-4 py-2">
		<label class="flex items-center gap-x-2 text-sm text-gray-600">
			<input type="checkbox" bind:checked={overlay} class="rounded border-gray-300 text-blue-600" />
			Show compliance overlay
		</label>
		<button
			onclick={() => graphComponent?.center()}
			class="inline-flex items-center gap-1 rounded-md px-2 py-1 text-sm text-gray-500 hover:bg-gray-100"
			title="Reset view"
		>
			<Icon src={ViewfinderCircle} class="h-4 w-4" />
			Reset
		</button>
	</div>

	{#if nodes.length === 0}
		<div class="p-16 text-center">
			<p class="text-sm text-gray-400">No resources discovered yet for this target.</p>
		</div>
	{:else}
		<DiscoveryGraph
			bind:this={graphComponent}
			{nodes}
			edges={cytoscapeEdges}
			{overlay}
			{onselect}
			initialSelect={internalSelectedId}
		/>
	{/if}
</div>
