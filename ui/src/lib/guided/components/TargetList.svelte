<script lang="ts">
	import { resolve } from '$app/paths';
	import { Button, FilterChips, PageHeading } from '$lib/ds';
	import type { TargetSummary } from '../source';
	import TargetTable from './TargetTable.svelte';

	let { targets }: { targets: TargetSummary[] } = $props();

	const types = $derived(['All', ...new Set(targets.map((t) => t.typeLabel))]);
	let type = $state('All');
	const shown = $derived(type === 'All' ? targets : targets.filter((t) => t.typeLabel === type));
</script>

<main class="mx-auto flex w-full max-w-7xl flex-col gap-10 px-10 pt-14 pb-20">
	<PageHeading
		title="Your Targets of Evaluation"
		intro="Products, services and organisations that you check against a framework. Open one to see what is left to do, or add a Target of Evaluation to get started."
	>
		{#snippet actions()}
			<Button href={resolve('/guided/new/')}>Add a Target of Evaluation</Button>
		{/snippet}
	</PageHeading>

	<FilterChips label="Filter by type" options={types} bind:value={type} />

	{#if targets.length === 0}
		<div class="flex flex-col items-start gap-4 rounded-card border border-line bg-surface p-8">
			<h2 class="text-lg font-semibold">No Targets of Evaluation yet</h2>
			<p class="max-w-xl text-ink-soft">
				A Target of Evaluation is a product, a cloud service or an organisation that you check against a framework. The
				setup takes about ten minutes.
			</p>
			<Button href={resolve('/guided/new/')}>Add your first Target of Evaluation</Button>
		</div>
	{:else}
		<TargetTable targets={shown} />
	{/if}
</main>
