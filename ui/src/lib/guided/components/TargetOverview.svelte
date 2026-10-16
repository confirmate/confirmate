<script lang="ts">
	import { resolve } from '$app/paths';
	import { Button, Card, Chip, PageHeading } from '$lib/ds';
	import type { TargetDetail } from '../source';
	import ClockBanner from './ClockBanner.svelte';
	import PartsCard from './PartsCard.svelte';
	import SourcesCard from './SourcesCard.svelte';
	import TaskList from './TaskList.svelte';

	let { target }: { target: TargetDetail } = $props();
</script>

<div class="flex flex-col gap-7">
	<PageHeading title={target.name}>
		{#snippet above()}
			<a href={resolve('/guided/')} class="text-sm no-underline">All Targets of Evaluation</a>
		{/snippet}
		{#snippet actions()}
			<Button variant="quiet" href={resolve(`/guided/${target.id}/risk-assessment/`)}>Risk assessment</Button>
			<Button>Create declaration of conformity</Button>
		{/snippet}
	</PageHeading>

	<div class="-mt-4 flex flex-wrap gap-2">
		<Chip tone="brand">{target.frameworkLabel}</Chip>
		{#each target.chips as chip (chip)}<Chip>{chip}</Chip>{/each}
	</div>

	{#if target.clock}<ClockBanner clock={target.clock} />{/if}

	<div class="flex flex-wrap items-start gap-6">
		<div class="min-w-0 flex-[3_1_32rem]">
			<Card>
				<div class="flex flex-wrap items-baseline justify-between gap-3">
					<h2 class="text-xl font-semibold">What to do next</h2>
					<a href={resolve(`/guided/${target.id}/todo/`)} class="text-sm">All {target.tasks.length} tasks</a>
				</div>
				<TaskList tasks={target.tasks.slice(0, 5)} targetId={target.id} />
			</Card>
		</div>
		<div class="flex min-w-0 flex-[2_1_20rem] flex-col gap-6">
			<PartsCard progress={target.progress} parts={target.parts} />
			<SourcesCard sources={target.sources} />
		</div>
	</div>
</div>
