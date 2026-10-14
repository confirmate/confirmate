<script lang="ts">
	import { resolve } from '$app/paths';
	import { Button, Chip, StackedProgress } from '$lib/ds';
	import type { TargetSummary } from '../source';

	let { target, primary = false }: { target: TargetSummary; primary?: boolean } = $props();

	const p = $derived(target.progress);
</script>

<article class="flex flex-col gap-5 rounded-card border border-line bg-surface p-(--cf-card-pad) font-body">
	<div class="flex items-start justify-between gap-4">
		<div class="flex flex-col gap-1.5">
			<h2 class="font-display text-heading font-(--cf-display-weight) leading-tight text-ink">{target.name}</h2>
			<p class="text-ink-muted">{target.subtitle}</p>
		</div>
		<span class="flex flex-wrap justify-end gap-2">
			<Chip>{target.kindLabel}</Chip>
			<Chip tone="brand">{target.frameworkLabel}</Chip>
		</span>
	</div>

	<div class="flex flex-col gap-2.5">
		<div class="flex items-baseline gap-2.5">
			<span class="font-display text-figure font-(--cf-display-weight) leading-none text-ink">{p.met}</span>
			<span class="text-ink-soft">of {p.total} requirements met</span>
		</div>
		<StackedProgress
			total={p.total}
			segments={[
				{ value: p.met, tone: 'ok', label: 'met' },
				{ value: p.attention, tone: 'bad', label: 'need attention' }
			]}
		/>
		<div class="flex flex-wrap gap-5 text-sm text-ink-soft">
			<span><strong class="text-ok">{p.met}</strong> met</span>
			<span><strong class="text-bad">{p.attention}</strong> need attention</span>
			<span><strong class="text-ink-muted">{p.open}</strong> still open</span>
		</div>
	</div>

	{#if target.nextDeadline}
		<div class="flex flex-col gap-1.5 border-t border-line pt-4">
			<span class="text-[0.8125rem] uppercase tracking-wider text-ink-muted">Next deadline</span>
			<span class="font-medium text-ink">{target.nextDeadline}</span>
		</div>
	{/if}

	{#if target.tags.length}
		<div class="flex flex-wrap gap-2">
			{#each target.tags as tag (tag)}<Chip>{tag}</Chip>{/each}
		</div>
	{/if}

	<div>
		<Button variant={primary ? 'primary' : 'quiet'} href={resolve(`/guided/${target.id}/`)}>Open {target.name}</Button>
	</div>
</article>
