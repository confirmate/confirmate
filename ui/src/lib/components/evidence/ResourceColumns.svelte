<script lang="ts">
	import type { SchemaResourceSnapshot as Resource, SchemaGraphEdge as GraphEdge } from '$lib/api/openapi/evidence';
	import type { SchemaAssessmentResult as AssessmentResult } from '$lib/api/openapi/orchestrator';
	import { Icon } from '@steeze-ui/svelte-icon';
	import { ChevronRight } from '@steeze-ui/heroicons';

	interface Props {
		resources: Resource[];
		edges: GraphEdge[];
		results: AssessmentResult[];
		onselect: (resource: Resource) => void;
	}
	let { resources, edges, results, onselect }: Props = $props();

	// Map parentId → children[]
	let childMap = $derived((() => {
		const rMap = new Map(resources.map((r) => [r.id, r]));
		const map = new Map<string, Resource[]>();
		for (const e of edges) {
			const child = rMap.get(e.source);
			if (!child) continue;
			const list = map.get(e.target) ?? [];
			list.push(child);
			map.set(e.target, list);
		}
		return map;
	})());

	// Root resources: not the source of any edge (no parent)
	let roots = $derived((() => {
		const hasParent = new Set(edges.map((e) => e.source));
		return resources.filter((r) => !hasParent.has(r.id));
	})());

	// Selected path from root to leaf
	let path = $state<Resource[]>([]);

	// One column per level: roots, then children of path[0], path[1], …
	let columns = $derived((() => {
		const cols: Resource[][] = [roots];
		for (const r of path) {
			const children = childMap.get(r.id) ?? [];
			if (children.length === 0) break;
			cols.push(children);
		}
		return cols;
	})());

	function select(colIdx: number, r: Resource) {
		path = [...path.slice(0, colIdx), r];
		onselect(r);
	}

	function resourceName(r: Resource): string {
		if (r.resource) {
			for (const val of Object.values(r.resource)) {
				if (val && typeof val === 'object' && 'name' in val && val.name) return String(val.name);
			}
		}
		return r.id.split('/').pop() ?? r.id;
	}

	function statusOf(id: string): 'pass' | 'fail' | 'none' {
		const rr = results.filter((r) => r.resourceId === id);
		if (rr.length === 0) return 'none';
		return rr.some((r) => !r.compliant) ? 'fail' : 'pass';
	}
</script>

<div class="flex min-h-0 flex-col">
	<!-- Breadcrumb -->
	{#if path.length > 0}
		<nav class="mb-3 flex items-center gap-1 text-sm text-gray-400 overflow-x-auto">
			<button onclick={() => (path = [])} class="shrink-0 hover:text-gray-700">All</button>
			{#each path as r, i}
				<Icon src={ChevronRight} class="h-3.5 w-3.5 shrink-0" />
				<button
					onclick={() => (path = path.slice(0, i + 1))}
					class="shrink-0 hover:text-gray-700 {i === path.length - 1 ? 'text-gray-900 font-medium' : ''}"
				>
					{resourceName(r)}
				</button>
			{/each}
		</nav>
	{/if}

	<!-- Columns -->
	<div class="flex min-h-0 flex-1 overflow-x-auto">
		{#each columns as col, colIdx}
			<div class="flex w-56 shrink-0 flex-col border-r border-gray-200 last:border-r-0 {colIdx > 0 ? 'bg-gray-50/50' : 'bg-white'}">
				<!-- Column header -->
				<div class="border-b border-gray-100 px-3 py-2">
					<p class="text-xs font-medium tracking-wide text-gray-400 uppercase">
						{colIdx === 0 ? 'Resources' : resourceName(path[colIdx - 1])}
					</p>
				</div>

				<!-- Items -->
				<div class="flex-1 overflow-y-auto">
					{#each col as r}
						{@const selected = path[colIdx]?.id === r.id}
						{@const status = statusOf(r.id)}
						{@const hasChildren = (childMap.get(r.id)?.length ?? 0) > 0}
						<button
							onclick={() => select(colIdx, r)}
							class="flex w-full items-center gap-2 px-3 py-2 text-left text-sm transition-colors
								{selected
									? 'bg-confirmate text-white'
									: 'text-gray-800 hover:bg-gray-100'}"
						>
							<!-- Status dot -->
							<span
								class="h-1.5 w-1.5 shrink-0 rounded-full
									{selected
										? 'bg-white/60'
										: status === 'pass'
											? 'bg-green-500'
											: status === 'fail'
												? 'bg-red-500'
												: 'bg-gray-300'}"
							></span>

							<!-- Name -->
							<span class="min-w-0 flex-1 truncate font-medium">{resourceName(r)}</span>

							<!-- Type chip -->
							<span class="shrink-0 text-xs {selected ? 'text-white/60' : 'text-gray-400'}">
								{r.resourceType.split(',')[0].trim()}
							</span>

							<!-- Chevron if has children -->
							{#if hasChildren}
								<Icon
									src={ChevronRight}
									class="h-3.5 w-3.5 shrink-0 {selected ? 'text-white/80' : 'text-gray-400'}"
								/>
							{/if}
						</button>
					{/each}
				</div>
			</div>
		{/each}
	</div>
</div>
