<script lang="ts">
	import type { SchemaResourceSnapshot as Resource } from '$lib/api/openapi/evidence';
	import type { SchemaAssessmentResult as AssessmentResult } from '$lib/api/openapi/orchestrator';
	import { Icon } from '@steeze-ui/svelte-icon';
	import { CheckCircle, XCircle, MagnifyingGlass, Funnel, MinusCircle } from '@steeze-ui/heroicons';

	interface Props {
		resources: Resource[];
		results: AssessmentResult[];
		onselect: (resource: Resource) => void;
	}
	let { resources, results, onselect }: Props = $props();

	let search = $state('');
	let typeFilter = $state<string[]>([]);
	let filterOpen = $state(false);

	const PAGE_SIZE = 15;
	let currentPage = $state(1);

	// Extract a human-readable name from the ontology union resource
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
		[...new Set(resources.map((r) => r.resourceType.split(',')[0].trim()))].sort()
	);

	let filtered = $derived(
		resources.filter((r) => {
			const name = resourceName(r).toLowerCase();
			const matchSearch = name.includes(search.toLowerCase());
			const matchType =
				typeFilter.length === 0 || typeFilter.includes(r.resourceType.split(',')[0].trim());
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

	function lastAssessed(id: string): string | null {
		const dates = results
			.filter((r) => r.resourceId === id && r.createdAt)
			.map((r) => r.createdAt!);
		if (dates.length === 0) return null;
		return dates.reduce((a, b) => (a > b ? a : b));
	}

	function formatRelative(iso: string): string {
		const diff = Date.now() - new Date(iso).getTime();
		const s = Math.floor(diff / 1000);
		if (s < 60) return `${s}s ago`;
		const m = Math.floor(s / 60);
		if (m < 60) return `${m}m ago`;
		const h = Math.floor(m / 60);
		if (h < 24) return `${h}h ago`;
		return `${Math.floor(h / 24)}d ago`;
	}
</script>

<div>
	<!-- Search + filter bar -->
	<div class="mb-4 flex items-center gap-x-3">
		<div class="relative max-w-sm flex-1">
			<div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
				<Icon src={MagnifyingGlass} class="h-4 w-4 text-gray-400" />
			</div>
			<input
				type="text"
				bind:value={search}
				placeholder="Search resources..."
				class="block w-full rounded-lg border border-gray-300 py-2 pr-3 pl-9 text-sm placeholder:text-gray-400 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 focus:outline-none"
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
				<div class="absolute right-0 z-10 mt-1 w-56 rounded-xl bg-white p-3 shadow-lg ring-1 ring-black/5">
					{#each allTypes as t}
						<label class="flex cursor-pointer items-center gap-2 rounded-md px-2 py-1.5 hover:bg-gray-50">
							<input type="checkbox" value={t} bind:group={typeFilter} class="rounded border-gray-300 text-blue-600" />
							<span class="text-sm text-gray-700">{t}</span>
						</label>
					{/each}
				</div>
			{/if}
		</div>

		<span class="ml-auto text-sm text-gray-400">{filtered.length} resource{filtered.length !== 1 ? 's' : ''}</span>
	</div>

	{#if filtered.length === 0}
		<div class="rounded-xl border border-dashed border-gray-300 p-16 text-center">
			<p class="text-sm text-gray-400">No resources match your filters.</p>
		</div>
	{:else}
		<div class="overflow-hidden rounded-xl border border-gray-200">
			<table class="min-w-full divide-y divide-gray-200">
				<thead class="bg-gray-50">
					<tr>
						<th class="px-4 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase">Resource</th>
						<th class="px-4 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase">Type</th>
						<th class="px-4 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase">Status</th>
						<th class="px-4 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase">Last assessed</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 bg-white">
					{#each pageItems as r}
						{@const rResults = resourceResults(r.id)}
						{@const hasBad = rResults.some((res) => !res.compliant)}
						{@const lastSeen = lastAssessed(r.id)}
						<tr class="cursor-pointer hover:bg-blue-50/40" onclick={() => onselect(r)}>
							<td class="px-4 py-3">
								<p class="text-sm font-medium text-gray-900">{resourceName(r)}</p>
								<p class="max-w-xs truncate font-mono text-xs text-gray-400">{r.id}</p>
							</td>
							<td class="px-4 py-3 text-sm text-gray-500">
								{r.resourceType.split(',')[0].trim()}
							</td>
							<td class="px-4 py-3">
								{#if rResults.length === 0}
									<span class="inline-flex items-center gap-1 text-xs text-gray-400">
										<Icon src={MinusCircle} class="h-4 w-4" />
										No results
									</span>
								{:else if hasBad}
									<span class="inline-flex items-center gap-1 text-xs font-medium text-red-600">
										<Icon src={XCircle} class="h-4 w-4" />
										{rResults.filter((r) => !r.compliant).length} failing
									</span>
								{:else}
									<span class="inline-flex items-center gap-1 text-xs font-medium text-green-600">
										<Icon src={CheckCircle} class="h-4 w-4" />
										All passing
									</span>
								{/if}
							</td>
							<td class="px-4 py-3 text-sm text-gray-400">
								{lastSeen ? formatRelative(lastSeen) : '—'}
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
