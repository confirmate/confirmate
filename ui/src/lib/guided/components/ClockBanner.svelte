<script lang="ts">
	import { Button } from '$lib/ds';
	import type { Clock } from '../source';

	let { clock }: { clock: Clock } = $props();
</script>

<section class="flex flex-col gap-4 rounded-card border border-bad/30 bg-bad-soft px-6 py-5 text-bad-ink">
	<div class="flex flex-wrap items-start justify-between gap-4">
		<div class="flex max-w-3xl flex-col gap-1">
			<strong class="text-lg">{clock.title}</strong>
			<span>{clock.text}</span>
		</div>
		<Button variant="danger">{clock.action}</Button>
	</div>
	<ol class="m-0 grid list-none grid-cols-[repeat(auto-fit,minmax(min(13.75rem,100%),1fr))] gap-3 p-0">
		{#each clock.steps as step (step.label)}
			<li class="flex flex-col rounded-control bg-surface px-4 py-3 text-ink">
				<span class="text-xs uppercase tracking-wider text-ink-muted">{step.label}</span>
				<strong>{step.due}</strong>
			</li>
		{/each}
	</ol>
</section>
