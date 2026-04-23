<script lang="ts" generics="T">
	import { Icon } from '@steeze-ui/svelte-icon';
	import { MagnifyingGlass } from '@steeze-ui/heroicons';
	import EmptyState from './EmptyState.svelte';

	interface Column<T> {
		key: string;
		label: string;
		sortable?: boolean;
		render?: (item: T) => string | undefined;
		href?: (item: T) => string | undefined;
		class?: string;
	}

	interface Props<T> {
		data: T[];
		columns: Column<T>[];
		searchPlaceholder?: string;
		searchKeys?: (keyof T)[];
		getRowKey: (item: T) => string;
	}

	let { data, columns, searchPlaceholder = 'Search...', searchKeys = [], getRowKey }: Props<T> = $props();

	let search = $state('');

	const filtered = $derived(
		data.filter((item) => {
			if (!search) return true;
			const q = search.toLowerCase();
			return searchKeys.some((key) => {
				const value = item[key];
				if (value === null || value === undefined) return false;
				return String(value).toLowerCase().includes(q);
			});
		})
	);

	let sortKey = $state<string | null>(null);
	let sortAsc = $state(true);

	function toggleSort(key: string) {
		if (sortKey === key) {
			sortAsc = !sortAsc;
		} else {
			sortKey = key;
			sortAsc = true;
		}
	}

	const sorted = $derived(() => {
		if (!sortKey) return filtered;
		const col = columns.find((c) => c.key === sortKey);
		if (!col?.render) return filtered;

		return [...filtered].sort((a, b) => {
			const aVal = col.render!(a) ?? '';
			const bVal = col.render!(b) ?? '';
			const cmp = aVal.localeCompare(bVal);
			return sortAsc ? cmp : -cmp;
		});
	});
</script>

<div class="space-y-4">
	<div class="relative">
		<div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
			<Icon src={MagnifyingGlass} class="h-4 w-4 text-gray-400" />
		</div>
		<input
			type="text"
			bind:value={search}
			placeholder={searchPlaceholder}
			class="block w-full rounded-lg border border-gray-300 py-2 pr-3 pl-9 text-sm text-gray-900 placeholder:text-gray-400 focus:border-confirmate focus:ring-1 focus:ring-confirmate focus:outline-none"
		/>
	</div>

	<p class="text-sm text-gray-400">
		{filtered.length === data.length
			? `${data.length} items`
			: `${filtered.length} of ${data.length} items`}
	</p>

	<div class="overflow-hidden rounded-lg border border-gray-200">
		<table class="min-w-full divide-y divide-gray-200">
			<thead class="bg-gray-50">
				<tr>
					{#each columns as col}
						<th
							scope="col"
							class="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500 {col.sortable
								? 'cursor-pointer hover:text-gray-700'
								: ''} {col.class ?? ''}"
							onclick={() => col.sortable && toggleSort(col.key)}
						>
							<div class="flex items-center gap-1">
								{col.label}
								{#if sortKey === col.key}
									<span class="text-confirmate">{sortAsc ? '↑' : '↓'}</span>
								{/if}
							</div>
						</th>
					{/each}
				</tr>
			</thead>
			<tbody class="divide-y divide-gray-200 bg-white">
				{#if sorted().length === 0}
					<tr>
						<td colspan={columns.length} class="px-4 py-8">
							<EmptyState
								title="No items found"
								description="Try adjusting your search criteria."
							/>
						</td>
					</tr>
				{:else}
					{#each sorted() as item}
						<tr class="hover:bg-gray-50">
							{#each columns as col}
								{@const value = col.render?.(item)}
								{@const href = col.href?.(item)}
								<td class="whitespace-nowrap px-4 py-3 {col.class ?? ''}">
									{#if href}
										<a {href} class="block text-sm text-gray-900 hover:text-confirmate">
											{value ?? '—'}
										</a>
									{:else}
										<span class="text-sm text-gray-900">{value ?? '—'}</span>
									{/if}
								</td>
							{/each}
						</tr>
					{/each}
				{/if}
			</tbody>
		</table>
	</div>
</div>