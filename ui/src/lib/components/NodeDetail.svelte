<script lang="ts">
	import { run } from 'svelte/legacy';

	import type { AssessmentResult, Metric } from '$lib/api/assessment';
	import type { Resource } from '$lib/api/evidence_store';
	import AssessmentIcon from './AssessmentIcon.svelte';
	import { flatten } from 'flat';

	interface Props {
		selected: Resource;
		results: AssessmentResult[];
		// TODO(oxisto): convert this to a store?!
		metrics: Map<string, Metric>;
		tab?: string;
	}

	let { selected, results, metrics, tab = $bindable('results') }: Props = $props();

	function name(id: string) {
		let rr = id.split('/');
		return rr[rr.length - 1];
	}

	const tabs = $state([
		{ name: 'Assessment Results', id: 'results', count: 0 },
		{ name: 'Properties', id: 'properties', count: 0 }
	]);

	/**
	 * This function returns an appropriate subset of properties out of the
	 * complete set, that are suitable for human viewing.
	 */
	function humanProperties(resource: Resource) {
		return Object.entries(flatten(resource.properties) ?? {}).filter(([k, v]) => {
			// raw is too big and labels contains a map which we cannot display for now
			if (k == 'raw' || k == 'labels') {
				return false;
			}

			// skip empty values
			if (v == 0 || v == undefined || v == null || v == '') {
				return false;
			}

			return true;
		});
	}

	run(() => {
		(() => {
			tabs[0].count = results.length;
		})();
	});
</script>

<div class="flex flex-col bg-white shadow-xl">
	<div class="bg-confirmate px-4 py-6 sm:px-6">
		<div class="flex items-center justify-between">
			<div class="truncate text-base leading-6 font-semibold text-white ...">
				{name(selected.id)}
			</div>
		</div>
		<div class="mt-1">
			<p class="text-sm text-gray-300">
				This resource is of type
				<i>
					{selected.resourceType.split(',')[0]}
				</i>
				and has <i>{results.length}</i> assessment results associated to it.
			</p>
		</div>
	</div>
	<div class="flex-1 px-4 sm:px-6">
		<!-- tab header -->
		<div>
			<div class="sm:hidden">
				<label for="tabs" class="sr-only">Select a tab</label>
				<!-- Use an "onChange" listener to redirect the user to the selected tab URL. -->
				<select
					id="tabs"
					name="tabs"
					class="focus:border-confirmate-light focus:ring-confirmate-light block w-full rounded-md border-gray-300 py-2 pr-10 pl-3 text-base focus:outline-hidden sm:text-sm"
				>
					{#each tabs as t (t.id)}
						<option selected={t.id == tab}>{t.name}</option>
					{/each}
				</select>
			</div>
			<div class="hidden sm:block">
				<div class="border-b border-gray-200">
					<nav class="-mb-px flex space-x-8" aria-label="Tabs">
						{#each tabs as t (t.name)}
							<button
								onclick={() => (tab = t.id)}
								class="{t.id == tab
									? 'border-confirmate-light text-confirmate'
									: 'border-transparent text-gray-500 hover:border-gray-200 hover:text-gray-700'} flex border-b-2 px-1 py-4 text-sm font-medium whitespace-nowrap"
								aria-current={t.id == tab ? 'page' : undefined}
							>
								{t.name}
								&nbsp;
								{#if t.count}
									<span
										class="{t.id == tab
											? 'text-confirmate bg-indigo-100'
											: 'bg-gray-100 text-gray-900'}ml-3 hidden rounded-full px-2.5 py-0.5 text-xs font-medium md:inline-block"
										>{t.count}</span
									>
								{/if}
							</button>
						{/each}
					</nav>
				</div>
			</div>
		</div>
		<!-- content -->
		<div class="pt-6 pb-5">
			{#if tab == 'results'}
				{#if results.length > 0}
					<div>
						<div class="space-y-4">
							{#each results as result}
								<div class="relative flex items-start">
									<div class="absolute flex h-6 items-center">
										<AssessmentIcon {result} />
									</div>
									<div class="pl-7 text-sm leading-6">
										<label for="privacy-public" class="font-medium text-gray-900">
											<a
												href={`/cloud/${selected.targetOfEvaluationId}/assessments/?filterIds=${result.id}`}
											>
												{metrics.get(result.metricId)?.name}
											</a>
										</label>
										<p id="privacy-public-description" class="text-gray-500">
											{metrics.get(result.metricId)?.description}
										</p>
									</div>
								</div>
							{/each}
						</div>
					</div>
				{:else}
					<div class="text-sm text-gray-900">No assessment results found for this resource.</div>
				{/if}
			{/if}
			{#if tab == 'properties'}
				<div>
					<div class="mt-2">
						<div class="flex space-x-2">
							<div class="px-4 pt-5 sm:px-0 sm:pt-0">
								<dl class="space-y-2">
									{#each humanProperties(selected) as [k, v] (k)}
										<div>
											<dt class="text-sm font-medium text-gray-500 sm:w-40 sm:shrink-0">
												{k}
											</dt>
											<dd class="text-gray-900sm:col-span-2 mt-1 w-96 text-sm">
												<p class="truncate ...">
													{#if k == 'url'}
														<a href={v} target="_blank">{v}</a>
													{:else if k == 'creationTime'}
														{#if !isNaN(v.valueOf())}
															<time datetime={new Date(v * 1000).toISOString()}>
																{Intl.DateTimeFormat(navigator.language, {
																	weekday: 'long',
																	year: 'numeric',
																	month: 'long',
																	hour: 'numeric',
																	minute: 'numeric',
																	day: 'numeric'
																}).format(v * 1000)}
															</time>
														{:else}
															<time>Invalid Date</time>
														{/if}
													{:else}
														{v}
													{/if}
												</p>
											</dd>
										</div>
									{/each}
								</dl>
							</div>
						</div>
					</div>
				</div>
			{/if}
		</div>
	</div>
</div>
