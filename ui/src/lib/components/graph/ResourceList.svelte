<script lang="ts">
	import type { SchemaResourceSnapshot as Resource } from '$lib/api/openapi/evidence';
	import type { SchemaAssessmentResult as AssessmentResult } from '$lib/api/openapi/orchestrator';
	import { Icon } from '@steeze-ui/svelte-icon';
	import { MagnifyingGlass, Funnel } from '@steeze-ui/heroicons';

	interface Props {
		resources: Resource[];
		results: AssessmentResult[];
		ongraphselect?: (id: string) => void;
	}
	let { resources, results, ongraphselect }: Props = $props();

	const PAGE_SIZE = 10;
	let currentPage = $state(1);
	let search = $state('');
	let typeFilter = $state<string[]>([]);
	let filterOpen = $state(false);

	// Extract name from ontology union resource
	function resourceName(r: Resource): string {
		if (r.resource) {
			for (const val of Object.values(r.resource)) {
				if (val && typeof val === 'object' && 'name' in val && val.name) {
					return String(val.name);
				}
			}
		}
		return r.id.split('/').pop() ?? r.id;
	}

	let allTypes = $derived(
		[...new Set(resources.map((r) => (r.resourceType ?? '').split(',')[0].trim()))].sort()
	);

	let filtered = $derived(
		resources.filter((r) => {
			const name = resourceName(r).toLowerCase();
			const matchSearch = name.includes(search.toLowerCase());
			const matchType =
				typeFilter.length === 0 ||
				typeFilter.includes((r.resourceType ?? '').split(',')[0].trim());
			return matchSearch && matchType;
		})
	);

	let totalPages = $derived(Math.max(1, Math.ceil(filtered.length / PAGE_SIZE)));
	let pageItems = $derived(filtered.slice((currentPage - 1) * PAGE_SIZE, currentPage * PAGE_SIZE));

	$effect(() => {
		void search;
		void typeFilter;
		currentPage = 1;
	});

	function resourceResults(id: string) {
		return results.filter((r) => r.resourceId === id);
	}
</script>

<div>
	<!-- Search + filter bar -->
	<div class="mb-4 flex items-center gap-x-3">
		<div class="relative max-w-xs flex-1">
			<div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
				<Icon src={MagnifyingGlass} class="h-4 w-4 text-gray-400" />
			</div>
			<input
				type="text"
				bind:value={search}
				placeholder="Search resources..."
				class="block w-full rounded-lg border border-gray-300 py-2 pr-3 pl-9 text-sm text-gray-900 placeholder:text-gray-400 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 focus:outline-none"
			/>
		</div>

		<div class="relative">
			<button
				onclick={() => (filterOpen = !filterOpen)}
				class="inline-flex items-center gap-1.5 rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
			>
				<Icon src={Funnel} class="h-4 w-4" />
				Type
				{#if typeFilter.length > 0}
					<span class="rounded-full bg-blue-100 px-1.5 py-0.5 text-xs font-semibold text-blue-700">
						{typeFilter.length}
					</span>
				{/if}
			</button>
			{#if filterOpen}
				<div
					class="absolute right-0 z-10 mt-1 w-56 origin-top-right rounded-xl bg-white p-3 shadow-lg ring-1 ring-black/5"
				>
					{#each allTypes as t}
						<label
							class="flex cursor-pointer items-center gap-2 rounded-md px-2 py-1.5 hover:bg-gray-50"
						>
							<input
								type="checkbox"
								value={t}
								bind:group={typeFilter}
								class="rounded border-gray-300 text-blue-600"
							/>
							<span class="text-sm text-gray-700">{t}</span>
						</label>
					{/each}
				</div>
			{/if}
		</div>
	</div>

	{#if filtered.length === 0}
		<div class="rounded-xl border border-dashed border-gray-300 p-12 text-center">
			<p class="text-sm text-gray-400">No resources match your filters.</p>
		</div>
	{:else}
		<div class="overflow-hidden rounded-xl border border-gray-200">
			<table class="min-w-full divide-y divide-gray-200">
				<thead class="bg-gray-50">
					<tr>
						<th class="px-4 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase">Name</th>
						<th class="px-4 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase">Type</th>
						<th class="px-4 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase">Results</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 bg-white">
					{#each pageItems as r}
						{@const name = resourceName(r)}
						{@const rtype = (r.resourceType ?? '').split(',')[0].trim()}
						{@const rResults = resourceResults(r.id)}
						{@const hasBad = rResults.some((res) => res.compliant === false)}
						<tr class="hover:bg-gray-50">
							<td class="px-4 py-3 text-sm text-gray-900">
								<button
									onclick={() => ongraphselect?.(r.id)}
									class="text-left font-medium hover:text-blue-600 hover:underline"
								>
									{name}
								</button>
								<p class="max-w-xs truncate font-mono text-xs text-gray-400">{r.id}</p>
							</td>
							<td class="px-4 py-3 text-sm text-gray-500">{rtype}</td>
							<td class="px-4 py-3">
								{#if rResults.length === 0}
									<span class="text-xs text-gray-400">—</span>
								{:else}
									<span
										class="{hasBad
											? 'bg-red-50 text-red-700'
											: 'bg-green-50 text-green-700'} inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium"
									>
										{rResults.length}
										{hasBad ? 'failing' : 'passing'}
									</span>
								{/if}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>

			{#if totalPages > 1}
				<div class="flex items-center justify-between border-t border-gray-200 bg-gray-50 px-4 py-3">
					<p class="text-sm text-gray-500">
						{(currentPage - 1) * PAGE_SIZE + 1}–{Math.min(currentPage * PAGE_SIZE, filtered.length)} of {filtered.length}
					</p>
					<div class="flex gap-x-2">
						<button
							onclick={() => (currentPage = Math.max(1, currentPage - 1))}
							disabled={currentPage === 1}
							class="rounded-md border border-gray-300 bg-white px-3 py-1.5 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-40"
						>Previous</button>
						<button
							onclick={() => (currentPage = Math.min(totalPages, currentPage + 1))}
							disabled={currentPage === totalPages}
							class="rounded-md border border-gray-300 bg-white px-3 py-1.5 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-40"
						>Next</button>
					</div>
				</div>
			{/if}
		</div>
	{/if}
</div>
