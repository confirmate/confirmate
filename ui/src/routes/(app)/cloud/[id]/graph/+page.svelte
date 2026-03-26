<script lang="ts">
	import DiscoveryGraph, { shouldCenter } from '$lib/components/DiscoveryGraph.svelte';
	import type { NodeDefinition, EdgeDefinition, ElementDefinition } from 'cytoscape';
	import type { PageData } from './$types';
	import { page } from '$app/state';
	import NodeDetail from '$lib/components/NodeDetail.svelte';
	import Button from '$lib/components/Button.svelte';
	import { ViewfinderCircle } from '@steeze-ui/heroicons';
	import { Icon } from '@steeze-ui/svelte-icon';

	interface Props {
		data: PageData;
	}

	let { data = $bindable() }: Props = $props();

	enum Status {
		WAITING,
		GOOD,
		BAD
	}

	let nodes = $derived(
		data.resources.map((n) => {
			let status = Status.WAITING;

			// fetch assessment results
			let results = data.results.filter((r) => {
				return r.resourceId == n.id;
			});

			if (results.length >= 1) {
				if (results.filter((r) => r.compliant == false).length > 0) {
					status = Status.BAD;
				} else {
					status = Status.GOOD;
				}
			}

			return {
				group: 'nodes',
				data: {
					id: n.id,
					status: status,
					label: n.properties.name,
					type: n.resourceType.split(',').reduce((a, v) => ({ ...a, [v]: true }), {})
				}
			} satisfies NodeDefinition;
		})
	);

	let id = $state(page.url.searchParams.get('id'));
	let tab = $state(page.url.searchParams.get('tab') ?? 'results');

	let edges = $derived(
		data.edges.map((e) => {
			return {
				group: 'edges',
				data: e
			} satisfies EdgeDefinition;
		})
	);

	let selected = $derived(
		data.resources.find((r) => {
			return r.id == id;
		})
	);

	function select(e: CustomEvent<ElementDefinition | null>) {
		if (e.detail == null) {
			id = null;
		} else {
			id = e.detail.data.id ?? null;
		}
	}

	let results = $derived(data.results.filter((r) => r.resourceId == id));

	let overlay = $state(false);

	$effect(() => {
		const url = new URL(page.url);
		if (id != null) {
			url.searchParams.set('id', id);
		} else {
			url.searchParams.set('id', '');
		}
		url.searchParams.set('tab', tab);

		history.replaceState({}, '', url);
	});
</script>

<div class="overflow-hidden rounded-xl border border-gray-200">
	<div class="flex items-center justify-between gap-x-4 border-b border-gray-900/5 bg-gray-50 p-4">
		<div class="text-sm text-gray-500">
			This graph provides an overview over all discovered resources of the Target of Evaluation,
			infrastructure as well as application source-code.
		</div>
		<div class="flex gap-x-1.5">
			<Button on:click={(e) => ($shouldCenter = true)}>
				<Icon src={ViewfinderCircle} class="h-4 w-4" />
			</Button>
		</div>
	</div>

	<div class="relative left-4 z-10 mt-4 flex items-start">
		<div class="flex h-6 items-center">
			<input
				id="overlay"
				aria-describedby="overlay-description"
				name="overlay"
				type="checkbox"
				class="text-confirmate focus:ring-confirmate h-4 w-4 rounded-sm border-gray-300"
				bind:checked={overlay}
			/>
		</div>
		<div class="ml-3 text-sm leading-6">
			<label for="overlay" class="font-medium text-gray-900">Show overlay</label>
			<span id="overlay-description" class="text-gray-500">
				<span class="sr-only">Show overlay </span>
				of assessment results
			</span>
		</div>
	</div>

	<dl class="-my-3 divide-y divide-gray-100 px-6 py-4 text-sm leading-6">
		<DiscoveryGraph {nodes} {edges} on:select={select} initialSelect={id} {overlay} />
	</dl>
</div>

<div class="absolute top-64 right-8 z-20 max-w-md">
	{#if selected}
		<NodeDetail {selected} {results} metrics={data.metrics} bind:tab />
	{/if}
</div>
