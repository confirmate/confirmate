<script lang="ts">
	import BelowHeader from '$lib/components/BelowHeader.svelte';
	import Header from '$lib/components/Header.svelte';
	import type { PageData } from './$types';

	interface Props {
		data: PageData;
	}

	let { data }: Props = $props();
</script>

<Header name="Metrics" description="" icon={false} buttons={false} />

<BelowHeader>A list of metrics that Confirmate is aware of.</BelowHeader>

<div class="px-4 sm:px-6 lg:px-8">
	<div class="mt-8 flow-root">
		<div class="-mx-4 -my-2 overflow-x-auto sm:-mx-6 lg:-mx-8">
			<div class="inline-block min-w-full py-2 align-middle sm:px-6 lg:px-8">
				<table class="min-w-full divide-y divide-gray-300">
					<thead>
						<tr>
							<th
								scope="col"
								class="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-3"
								>ID</th
							>
							<th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900"
								>Name</th
							>
							<th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900"
								>Category</th
							>
						</tr>
					</thead>

					<tbody>
						{#each data.metrics as metric, i}
							<tr class={i % 2 == 0 ? 'bg-white' : 'bg-gray-50'}>
								<td class="whitespace-nowrap px-3 py-3.5 text-sm font-medium text-gray-900">
									{metric.id}
								</td>
								<td class="whitespace-nowrap px-3 py-3.5 text-sm text-gray-500">
									{metric.name}
								</td>
								<td class="whitespace-nowrap px-3 py-3.5 text-sm text-gray-500">
									{metric.category}
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</div>
	</div>
</div>

<style lang="postcss">
	:global(html) {
		background-color: var(--color-gray-100);
	}
</style>
