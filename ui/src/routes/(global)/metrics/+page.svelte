<script lang="ts">
	import type { SchemaMetric } from '$lib/api/openapi/orchestrator';
	import SectionHeader from '$lib/components/ui/SectionHeader.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import MetricGroup from '$lib/components/metrics/MetricGroup.svelte';
	import { Icon } from '@steeze-ui/svelte-icon';
	import { MagnifyingGlass, ChevronDown } from '@steeze-ui/heroicons';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const allMetrics = $derived([...data.metrics.values()]);
	const allCategories = $derived([...new Set(allMetrics.map((m) => m.category))].sort());
	const categoryCounts = $derived(
		allMetrics.reduce<Record<string, number>>((acc, m) => {
			acc[m.category] = (acc[m.category] ?? 0) + 1;
			return acc;
		}, {})
	);
	const implementedCount = $derived(allMetrics.filter((m) => m.implementation).length);

	let search = $state('');
	let selectedCategories = $state<Set<string>>(new Set());
	let filterOpen = $state(false);
	// Track which groups are collapsed; all start open
	let closedGroups = $state<Set<string>>(new Set());

	function toggleGroup(cat: string) {
		const next = new Set(closedGroups);
		if (next.has(cat)) next.delete(cat);
		else next.add(cat);
		closedGroups = next;
	}

	const filtered = $derived(
		allMetrics.filter((m) => {
			const q = search.toLowerCase();
			const matchSearch =
				!q ||
				m.name.toLowerCase().includes(q) ||
				m.description.toLowerCase().includes(q) ||
				m.category.toLowerCase().includes(q);
			const matchCat = selectedCategories.size === 0 || selectedCategories.has(m.category);
			return matchSearch && matchCat;
		})
	);

	const grouped = $derived(
		filtered.reduce<Map<string, SchemaMetric[]>>((acc, m) => {
			const list = acc.get(m.category) ?? [];
			list.push(m);
			acc.set(m.category, list);
			return acc;
		}, new Map())
	);

	function toggleCategoryFilter(cat: string) {
		const next = new Set(selectedCategories);
		if (next.has(cat)) next.delete(cat);
		else next.add(cat);
		selectedCategories = next;
	}
</script>

<div>
	<SectionHeader
		title="Metrics"
		description="Metrics define measurable security properties assessed against your resources. Each metric specifies what to measure and how to interpret the result; automated metrics include a Rego-based implementation for continuous evaluation."
	/>

	<!-- Stats -->
	<div class="mt-6 grid grid-cols-3 gap-4">
		<div class="rounded-lg border border-gray-200 bg-white px-5 py-4">
			<div class="text-sm font-medium text-gray-500">Total</div>
			<div class="mt-1 text-2xl font-bold text-confirmate">{allMetrics.length}</div>
		</div>
		<div class="rounded-lg border border-gray-200 bg-white px-5 py-4">
			<div class="text-sm font-medium text-gray-500">Categories</div>
			<div class="mt-1 text-2xl font-bold text-confirmate">{allCategories.length}</div>
		</div>
		<div class="rounded-lg border border-gray-200 bg-white px-5 py-4">
			<div class="text-sm font-medium text-gray-500">Automated</div>
			<div class="mt-1 text-2xl font-bold text-confirmate">{implementedCount}</div>
			<div class="mt-0.5 text-xs text-gray-400">of {allMetrics.length} with implementation</div>
		</div>
	</div>

	<!-- Search + category filter -->
	<div class="mt-6 flex items-center gap-3">
		<div class="relative flex-1">
			<div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
				<Icon src={MagnifyingGlass} class="h-4 w-4 text-gray-400" />
			</div>
			<input
				type="text"
				bind:value={search}
				placeholder="Search by name, description or category…"
				class="block w-full rounded-lg border border-gray-300 py-2 pr-3 pl-9 text-sm text-gray-900 placeholder:text-gray-400 focus:border-confirmate focus:ring-1 focus:ring-confirmate focus:outline-none"
			/>
		</div>

		<div class="relative">
			<button
				type="button"
				onclick={() => (filterOpen = !filterOpen)}
				class="inline-flex items-center gap-1.5 rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
			>
				Category
				{#if selectedCategories.size > 0}
					<span class="rounded-full bg-confirmate px-1.5 py-0.5 text-xs font-semibold text-white">
						{selectedCategories.size}
					</span>
				{/if}
				<Icon src={ChevronDown} class="h-4 w-4 text-gray-400" />
			</button>

			{#if filterOpen}
				<div class="absolute right-0 z-10 mt-1 w-64 origin-top-right overflow-hidden rounded-xl bg-white shadow-lg ring-1 ring-black/5">
					<div class="max-h-72 overflow-y-auto p-2">
						{#each allCategories as cat}
							{@const active = selectedCategories.has(cat)}
							<button
								type="button"
								onclick={() => toggleCategoryFilter(cat)}
								class="flex w-full items-center justify-between rounded-md px-3 py-2 text-sm hover:bg-gray-50 {active ? 'text-confirmate' : 'text-gray-700'}"
							>
								<span>{cat}</span>
								<span class="flex items-center gap-2">
									<span class="text-xs text-gray-400">{categoryCounts[cat]}</span>
									{#if active}
										<span class="h-1.5 w-1.5 rounded-full bg-confirmate"></span>
									{/if}
								</span>
							</button>
						{/each}
					</div>
					{#if selectedCategories.size > 0}
						<div class="border-t border-gray-100 p-2">
							<button
								type="button"
								onclick={() => (selectedCategories = new Set())}
								class="w-full rounded-md px-3 py-1.5 text-xs text-gray-500 hover:bg-gray-50"
							>
								Clear filter
							</button>
						</div>
					{/if}
				</div>
			{/if}
		</div>
	</div>

	<!-- Active filter chips -->
	{#if selectedCategories.size > 0}
		<div class="mt-3 flex flex-wrap gap-2">
			{#each [...selectedCategories].sort() as cat}
				<button
					type="button"
					onclick={() => toggleCategoryFilter(cat)}
					class="inline-flex items-center gap-1 rounded-full bg-blue-50 px-2.5 py-1 text-xs font-medium text-confirmate hover:bg-blue-100"
				>
					{cat} <span class="opacity-60">×</span>
				</button>
			{/each}
		</div>
	{/if}

	<p class="mt-4 text-sm text-gray-400">
		{filtered.length === allMetrics.length
			? `${allMetrics.length} metrics`
			: `${filtered.length} of ${allMetrics.length} metrics`}
	</p>

	<!-- Grouped metric list -->
	<div class="mt-2 space-y-2">
		{#if filtered.length === 0}
			<EmptyState title="No metrics found" description="Try adjusting your search or category filter." />
		{:else}
			{#each [...grouped.entries()].sort(([a], [b]) => a.localeCompare(b)) as [cat, metrics]}
				<MetricGroup
					category={cat}
					{metrics}
					open={!closedGroups.has(cat)}
					ontoggle={() => toggleGroup(cat)}
				/>
			{/each}
		{/if}
	</div>
</div>
