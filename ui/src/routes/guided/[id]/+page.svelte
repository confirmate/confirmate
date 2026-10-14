<script lang="ts">
	import { resolve } from '$app/paths';
	import { Button, Card, Chip } from '$lib/ds';
	import ClockBanner from '$lib/guided/components/ClockBanner.svelte';
	import PartsCard from '$lib/guided/components/PartsCard.svelte';
	import SourcesCard from '$lib/guided/components/SourcesCard.svelte';
	import TaskList from '$lib/guided/components/TaskList.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	// The layout only renders its children for a known target.
	const target = $derived(data.target!);
</script>

<div class="flex flex-col gap-7">
	<div class="flex flex-wrap items-end justify-between gap-5">
		<div class="flex flex-col gap-2">
			<a href={resolve('/guided/')} class="text-sm no-underline">All Targets of Evaluation</a>
			<h1 class="font-display text-[2.375rem] leading-tight">{target.name}</h1>
			<div class="flex flex-wrap gap-2">
				<Chip tone="brand">{target.frameworkLabel}</Chip>
				{#each target.chips as chip (chip)}<Chip>{chip}</Chip>{/each}
			</div>
		</div>
		<div class="flex flex-wrap gap-2.5">
			<Button variant="quiet">Risk assessment, version 1</Button>
			<Button>Create declaration of conformity</Button>
		</div>
	</div>

	{#if target.clock}<ClockBanner clock={target.clock} />{/if}

	<div class="flex flex-wrap items-start gap-6">
		<div class="min-w-0 flex-[3_1_32rem]">
			<Card>
				<div class="flex flex-wrap items-baseline justify-between gap-3">
					<h2 class="text-xl font-semibold">What to do next</h2>
					<a href={resolve(`/guided/${target.id}/todo/`)} class="text-sm">All {target.tasks.length} tasks</a>
				</div>
				<TaskList tasks={target.tasks} targetId={target.id} />
			</Card>
		</div>
		<div class="flex min-w-0 flex-[2_1_20rem] flex-col gap-6">
			<PartsCard progress={target.progress} parts={target.parts} />
			<SourcesCard sources={target.sources} />
		</div>
	</div>
</div>
