<script lang="ts">
	import { resolve } from '$app/paths';
	import { Button, FilterChips, PageHeading } from '$lib/ds';
	import type { TargetSummary } from '../source';
	import TargetCard from './TargetCard.svelte';

	let { targets }: { targets: TargetSummary[] } = $props();

	const types = $derived(['All', ...new Set(targets.map((t) => t.typeLabel))]);
	let type = $state('All');
	const shown = $derived(type === 'All' ? targets : targets.filter((t) => t.typeLabel === type));
</script>

<main class="mx-auto flex w-full max-w-310 flex-col gap-10 px-10 pt-14 pb-20">
	<PageHeading
		title="Your Targets of Evaluation"
		intro="Products, services and organisations that you check against a framework. Open one to see what is left to do, or add a Target of Evaluation to get started."
	>
		{#snippet actions()}
			<Button href={resolve('/guided/new/')}>Add a Target of Evaluation</Button>
		{/snippet}
	</PageHeading>

	<FilterChips label="Filter by type" options={types} bind:value={type} />

	<section class="grid grid-cols-[repeat(auto-fit,minmax(min(32rem,100%),1fr))] gap-6">
		{#each shown as target, i (target.id)}
			<TargetCard {target} primary={i === 0} />
		{/each}
	</section>
</main>
