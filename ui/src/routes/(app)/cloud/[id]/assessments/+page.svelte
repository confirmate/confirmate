<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import type { AssessmentResult } from '$lib/api/assessment';
	import AssessmentIcon from '$lib/components/AssessmentIcon.svelte';
	import Button from '$lib/components/Button.svelte';
	import StarterHint from '$lib/components/StarterHint.svelte';
	import { QueueList } from '@steeze-ui/heroicons';
	import type { PageData } from './$types';

	interface Props {
		data: PageData;
	}

	let { data }: Props = $props();

	let currentPage = $derived(data.page ? data.page : 1);
	let rowsPerPage = 9;
	let filteredData = $derived(filter(data.results, data.filterIds, data.filterResourceId));
	let totalPages = $derived(Math.ceil(filteredData.length / rowsPerPage));
	let currentData = $derived(paginate(filteredData, currentPage));

	function filter(
		results: AssessmentResult[],
		ids: string[],
		resourceId: string | null
	): AssessmentResult[] {
		if (ids.length > 0) {
			results = results.filter((r) => ids.includes(r.id));
		}

		if (resourceId !== null) {
			results = results.filter((r) => r.resourceId == resourceId);
		}

		return results;
	}

	function paginate(results: AssessmentResult[], page: number) {
		const startIndex = (page - 1) * rowsPerPage;
		const endIndex = startIndex + rowsPerPage;

		return results.slice(startIndex, endIndex);
	}

	function prevPage() {
		if (currentPage > 1) {
			const p = currentPage - 1;
			updatePageQuery(p);
		}
	}

	function nextPage() {
		if (currentPage < totalPages) {
			const p = currentPage + 1;
			updatePageQuery(p);
		}
	}

	function updatePageQuery(p: number) {
		page.url.searchParams.set('page', String(p));
		goto(`?${page.url.searchParams.toString()}`);
	}

	let showModalId: string | null = $state(null);

	function showDetails(id: string) {
		showModalId = id;
	}

	function closeModal() {
		showModalId = null;
	}
</script>

{#snippet hint()}
	<span>Confirmate Assessment component</span>
{/snippet}

{#if data.resources.length == 0}
	<StarterHint type="assessment results" icon={QueueList} component={hint} />
{:else}
	<main>
		<div class="-mx-4 -my-2 overflow-x-clip sm:-mx-6 lg:-mx-8">
			<div class="inline-block min-w-full py-2 align-middle sm:px-6 lg:px-8">
				<table class="min-w-full table-fixed divide-y divide-gray-200">
					<thead class="bg-gray-50">
						<tr>
							<th
								scope="col"
								class="w-1/12 px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500"
							>
								Compliance
							</th>
							<th
								scope="col"
								class="w-1/12 px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500"
							>
								Time
							</th>
							<th
								scope="col"
								class="w-3/12 px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500"
							>
								Metric
							</th>
							<th
								scope="col"
								class="w-3/12 px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500"
							>
								Resource Name
							</th>
							<th
								scope="col"
								class="w-2/12 px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500"
							>
								Type
							</th>
							<th
								scope="col"
								class="w-1/12 px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500"
							>
							</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-200 bg-white">
						{#each currentData as assessment}
							<tr>
								<td class="px-6 py-4 whitespace-nowrap">
									<AssessmentIcon result={assessment} />
								</td>
								<td class="px-6 py-4 whitespace-nowrap">
									<span class="text-sm text-gray-900"
										>{new Date(assessment.timestamp).toLocaleString()}</span
									>
								</td>
								<td class="max-w-0 truncate px-6 py-4 whitespace-nowrap">
									<span class="text-sm text-gray-900">
										<a href="/metrics/{data.service.id}/{assessment.metricId}">
											{assessment.metricId}
										</a>
									</span>
								</td>
								<td class="max-w-xs truncate px-6 py-4 whitespace-nowrap">
									<span class="text-sm text-gray-900">
										<a href={`/cloud/${data.service.id}/resource/?id=${assessment.resourceId}`}>
											{assessment.resourceId.split('/')[
												assessment.resourceId.split('/').length - 1
											]}
										</a>
									</span>
								</td>
								<td class="px-6 py-4 whitespace-nowrap">
									<span class="text-sm text-gray-900">{assessment.resourceTypes[0]}</span>
								</td>
								<td class="px-6 py-4 whitespace-nowrap">
									<Button onclick={() => showDetails(assessment.id)}>More Details</Button>
								</td>
							</tr>
							{#if showModalId === assessment.id}
								<tr>
									<td colspan="6" class="bg-gray-50 px-6 py-4">
										<pre class="font-mono text-xs whitespace-pre-wrap">{JSON.stringify(
												assessment,
												null,
												2
											)}</pre>
										<Button onclick={closeModal}>Close</Button>
									</td>
								</tr>
							{/if}
						{/each}
					</tbody>
				</table>

				<div class="mt-4">
					<button
						class={'rounded-md bg-gray-100 px-4 py-2 text-sm font-medium text-gray-500 hover:bg-gray-200 focus:bg-gray-200 focus:outline-hidden' +
							(currentPage === 1 ? ' cursor-not-allowed opacity-50' : '')}
						onclick={prevPage}
						disabled={currentPage === 1}
					>
						Previous
					</button>
					<span class="mx-2 text-gray-500">{currentPage} / {totalPages}</span>
					<button
						class={'rounded-md bg-gray-100 px-4 py-2 text-sm font-medium text-gray-500 hover:bg-gray-200 focus:bg-gray-200 focus:outline-hidden' +
							(currentPage === totalPages ? ' cursor-not-allowed opacity-50' : '')}
						onclick={nextPage}
						disabled={currentPage === totalPages}
					>
						Next
					</button>
				</div>
			</div>
		</div>
	</main>
{/if}
