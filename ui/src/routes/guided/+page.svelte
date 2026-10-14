<script lang="ts">
	import { resolve } from '$app/paths';
	import { Button } from '$lib/ds';
	import TargetCard from '$lib/guided/components/TargetCard.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const kinds = $derived(['All', ...new Set(data.targets.map((t) => t.kindLabel))]);
	let kind = $state('All');
	const shown = $derived(kind === 'All' ? data.targets : data.targets.filter((t) => t.kindLabel === kind));
</script>

<main class="mx-auto flex w-full max-w-310 flex-col gap-10 px-10 pt-14 pb-20">
	<div class="flex flex-wrap items-end justify-between gap-6">
		<div class="flex max-w-2xl flex-col gap-2">
			<h1 class="font-display text-[2.5rem] leading-tight">Your Targets of Evaluation</h1>
			<p class="text-[1.0625rem] text-ink-muted">
				Products, services and organisations that you check against a framework. Open one to see what is left to
				do, or add a Target of Evaluation to get started.
			</p>
		</div>
		<Button href={resolve('/guided/new/')}>Add a Target of Evaluation</Button>
	</div>

	<div role="group" aria-label="Filter by kind" class="flex flex-wrap gap-2">
		{#each kinds as k (k)}
			<button
				type="button"
				aria-pressed={kind === k}
				onclick={() => (kind = k)}
				class="h-9 cursor-pointer rounded-full px-4 font-medium {kind === k
					? 'bg-brand text-on-brand'
					: 'border border-line-strong bg-surface text-ink hover:bg-sunken'}"
			>
				{k}
			</button>
		{/each}
	</div>

	<section class="grid grid-cols-[repeat(auto-fit,minmax(min(32rem,100%),1fr))] gap-6">
		{#each shown as target, i (target.id)}
			<TargetCard {target} primary={i === 0} />
		{/each}
	</section>
</main>
