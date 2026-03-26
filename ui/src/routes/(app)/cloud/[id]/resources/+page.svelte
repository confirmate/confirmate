<script lang="ts">
	import { run } from 'svelte/legacy';

	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import type { Resource } from '$lib/api/evidence_store';
	import StarterHint from '$lib/components/StarterHint.svelte';
	import {
		Check,
		ChevronDown,
		ChevronUp,
		Clipboard,
		Funnel,
		MagnifyingGlass,
		QueueList,
		Squares2x2,
		XCircle
	} from '@steeze-ui/heroicons';
	import { Icon } from '@steeze-ui/svelte-icon';
	import type { PageData } from './$types';

	interface Props {
		data: PageData;
	}

	let { data }: Props = $props();

	let rowsPerPage = 8;

	let copyingId: string | null = $state(null);

	let filterOptions: string[] = $state([]);

	function toggleSearch() {
		searchActivated = !searchActivated;
		if (!searchActivated) {
			searchString = '';
			filteredData = data.resources;
		}
	}

	let sortAscending = $state(false);

	function sort(sortBy: string) {
		sortAscending = !sortAscending;

		switch (sortBy) {
			case 'name':
				filteredData.sort((a, b) => {
					if (sortAscending) return a.properties['name'].localeCompare(b.properties.name);
					else return b.properties['name'].localeCompare(a.properties.name);
				});
				break;
			case 'type':
				filteredData.sort((a, b) => {
					if (sortAscending) return a.resourceType.localeCompare(b.resourceType);
					else return b.resourceType.localeCompare(a.resourceType);
				});
				break;
		}
		currentData = paginate(filteredData, currentPage);
	}

	let types: Set<String> = new Set();
	let typeArray: String[] = $state([]);

	function paginate(results: Resource[], page: number) {
		if (currentPage > totalPages) {
			page = totalPages;
			currentPage = totalPages;
		}

		const startIndex = (page - 1) * rowsPerPage;
		const endIndex = startIndex + rowsPerPage;

		types = data.resources.reduce((acc, resource) => {
			acc.add(resource.resourceType.split(',')[0]);
			return acc;
		}, new Set<String>());
		typeArray = Array.from(types);

		return results.slice(startIndex, endIndex);
	}

	function updatePageQuery(currentPage: number) {
		$page.url.searchParams.set('page', String(currentPage));
		goto(`?${$page.url.searchParams.toString()}`);
	}

	function prevPage() {
		if (currentPage > 1) {
			currentPage--;
			updatePageQuery(currentPage);
		}
	}

	function nextPage() {
		if (currentPage < totalPages) {
			currentPage++;
			updatePageQuery(currentPage);
		}
	}

	async function copyId(id: string) {
		if (copyingId !== null) return;
		copyingId = id;
		try {
			await new Promise((resolve) => setTimeout(resolve, 250));
			navigator.clipboard.writeText(id);
		} catch (error) {
			console.error(error);
		} finally {
			copyingId = null;
		}
	}

	let filterOptionsVisible = $state(false);

	function toggleFilterOptions() {
		filterOptionsVisible = !filterOptionsVisible;
	}

	function goToAssessmentResults(resourceId: string) {
		goto(`/cloud/${data.service.id}/assessments/?filterResourceId=${resourceId}`);
	}

	async function countAssessmentResults(resourceId: string) {
		const results = data.results.filter((result) => result.resourceId === resourceId);
		return results.length;
	}
	let currentPage: number = $state(1);
	run(() => {
		currentPage = data.page ? data.page : 1;
	});
	let searchString = $state('');

	let query = $derived(searchString.toLowerCase());

	let filteredData: Resource[] = data.resources;
	run(() => {
		filteredData = data.resources.filter((resource) => {
			return (
				resource.properties.name.toLowerCase().includes(query) &&
				(filterOptions.length == 0 || filterOptions.includes(resource.resourceType.split(',')[0]))
			);
		});
	});
	let totalPages = $derived(Math.ceil(filteredData.length / rowsPerPage));
	let searchActivated = $state(false);

	let currentData: Resource[] = $state([]);
	run(() => {
		currentData = paginate(filteredData, currentPage);
	});
</script>

{#if data.resources.length == 0}
	<StarterHint type="discovered resources" icon={Squares2x2}>
		{#snippet component()}
			<span>
				Confirmate Discovery with the Target of Evaluation ID <pre>{data.service.id}</pre>
			</span>
		{/snippet}
	</StarterHint>
{:else}
	<div class="-mx-4 -my-2 overflow-x-auto sm:-mx-6 lg:-mx-8">
		<div class="inline-block min-w-full py-2 align-middle sm:px-6 lg:px-8">
			<table class="min-w-full divide-y divide-gray-300">
				<thead class="bg-gray-50">
					<tr>
						<th
							scope="col"
							class="w-1/4 px-3 py-3 text-left text-xs font-medium tracking-wider text-gray-500"
						>
							Actions
						</th>
						<th
							scope="col"
							class="group w-6/12 px-3 py-3 pl-0 text-left align-middle text-xs font-medium tracking-wider text-gray-500"
						>
							<div class="flex">
								<button
									type="button"
									class=" flex items-center align-middle"
									onclick={() => sort('name')}
								>
									Name
									<span
										class="invisible ml-2 flex items-center rounded-sm text-gray-400 group-hover:visible group-focus:visible"
									>
										<Icon src={sortAscending ? ChevronDown : ChevronUp} class="h-4 w-4" />
									</span>
								</button>
								<button
									type="button"
									class="flex items-center align-middle uppercase group-hover:visible group-focus:visible"
									onclick={() => toggleSearch()}
								>
									<span
										class="invisible ml-2 rounded-sm text-gray-400 group-hover:visible group-focus:visible"
									>
										<Icon
											src={searchActivated ? XCircle : MagnifyingGlass}
											class="h-5 w-5 text-gray-400"
										/>
									</span>
								</button>
								{#if searchActivated}
									<div>
										<div class="relative rounded-md shadow-xs">
											<div
												class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3"
											>
												<Icon src={MagnifyingGlass} class="h-5 w-5 text-gray-400" />
											</div>
											<input
												name="search"
												id="search"
												class="block h-7 w-full rounded-md border-0 pl-10 text-gray-900 ring-1 ring-gray-300 ring-inset placeholder:text-gray-400 focus:ring-2 focus:ring-indigo-600 focus:ring-inset sm:text-sm sm:leading-6"
												placeholder="Search Resources"
												bind:value={searchString}
											/>
										</div>
									</div>
								{/if}
							</div>
						</th>
						<th
							scope="col"
							class="group w-1/4 px-3 py-3 text-left text-xs font-medium text-gray-500 uppercase sm:pl-0"
						>
							<div class="flex">
								<button type="button" class="inline-flex" onclick={() => sort('type')}>
									Type
									<span
										class="invisible ml-2 flex-none rounded-sm text-gray-400 group-hover:visible group-focus:visible"
									>
										<Icon src={sortAscending ? ChevronDown : ChevronUp} class="h-4 w-4" />
									</span>
								</button>
								<div
									class="invisible max-w-3xl text-center group-hover:visible group-focus:visible lg:max-w-7xl"
								>
									<div class="flex items-center justify-between">
										<div class="flex inline-flex text-left">
											<div class="flex">
												<button
													type="button"
													class="inline-flex items-center justify-center text-sm font-medium text-gray-700 hover:text-gray-900"
													aria-expanded="false"
													onclick={toggleFilterOptions}
												>
													<Icon
														src={Funnel}
														class="-mr-1 ml-1 h-5 w-5 shrink-0 text-gray-400 group-hover:text-gray-500"
													/>
													<span
														class="ml-1.5 rounded-sm bg-gray-200 px-1.5 py-0.5 text-xs font-semibold text-gray-700 tabular-nums"
														>{filterOptions.length}</span
													>
												</button>
											</div>
											{#if filterOptionsVisible}
												<div
													class="ring-opacity-5 absolute z-10 mt-2 origin-top-right rounded-md bg-white p-4 shadow-2xl ring-1 ring-black focus:outline-hidden"
												>
													<form class="space-y-4">
														{#each typeArray as type}
															<div class="flex items-center">
																<input
																	id="filter-category-0"
																	name="category[]"
																	value={type}
																	type="checkbox"
																	class="h-4 w-4 rounded-sm border-gray-300 text-indigo-600 focus:ring-indigo-500"
																	bind:group={filterOptions}
																/>
																<label
																	for="filter-category-0"
																	class="ml-3 pr-6 text-sm font-medium whitespace-nowrap text-gray-900"
																	>{type}</label
																>
															</div>
														{/each}
													</form>
												</div>
											{/if}
										</div>
									</div>
								</div>
							</div>
						</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-200 bg-white">
					{#each currentData as resource}
						<tr>
							<td class="px-6 py-4 pl-2 whitespace-nowrap">
								<button
									class="inline-flex items-center rounded-md bg-gray-100 px-4 py-2 pl-0 text-sm font-medium text-gray-500 hover:bg-gray-200 focus:bg-gray-200 focus:outline-hidden"
									onclick={() => copyId(resource.id)}
									title={resource.id}
									disabled={copyingId !== null}
								>
									{#if copyingId === resource.id}
										<Icon src={Check} class="mr-2 h-5 w-5" />
									{:else}
										<Icon src={Clipboard} class="mr-2 h-5 w-5" />
									{/if}
									Copy ID
								</button>
								{#await countAssessmentResults(resource.id) then assessmentResultCount}
									<button
										class={'inline-flex items-center rounded-md bg-gray-100 px-4 py-2 text-sm font-medium text-gray-500 hover:bg-gray-200 focus:bg-gray-200 focus:outline-hidden' +
											(assessmentResultCount === 0 ? ' cursor-not-allowed opacity-50' : '')}
										title={'Go to ' +
											assessmentResultCount +
											' assessment result' +
											(assessmentResultCount === 1 ? '' : 's') +
											' for ' +
											resource.properties.name +
											'.'}
										onclick={() => goToAssessmentResults(resource.properties.id)}
										disabled={assessmentResultCount === 0}
									>
										<Icon src={QueueList} class="mr-2 h-5 w-5" />

										{assessmentResultCount}
									</button>
								{/await}
							</td>
							<td class="py-4 pr-3 pl-4 text-sm whitespace-nowrap text-gray-900 sm:pl-0">
								<span class="text-sm text-gray-900">{resource.properties.name}</span>
							</td>
							<td class="py-4 pr-3 pl-4 text-sm whitespace-nowrap text-gray-900 sm:pl-0">
								<span class="text-sm text-gray-900">{resource.resourceType.split(',')[0]}</span>
							</td>
						</tr>
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
{/if}
