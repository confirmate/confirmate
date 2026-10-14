<script lang="ts">
	import { Card } from '$lib/ds';
	import TaskList from '$lib/guided/components/TaskList.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const target = $derived(data.target!);
	const assignees = $derived(['Everyone', ...new Set(target.tasks.map((t) => t.assignee))]);
	let assignee = $state('Everyone');
	const shown = $derived(assignee === 'Everyone' ? target.tasks : target.tasks.filter((t) => t.assignee === assignee));
</script>

<div class="flex max-w-4xl flex-col gap-6">
	<div class="flex flex-col gap-2">
		<h1 class="font-display text-4xl leading-tight">To do</h1>
		<p class="text-ink-muted">Everything that needs a person: failing checks, missing documents and open deadlines.</p>
	</div>

	<div role="group" aria-label="Filter by person" class="flex flex-wrap gap-2">
		{#each assignees as a (a)}
			<button
				type="button"
				aria-pressed={assignee === a}
				onclick={() => (assignee = a)}
				class="h-9 cursor-pointer rounded-full px-4 font-medium {assignee === a
					? 'bg-brand text-on-brand'
					: 'border border-line-strong bg-surface text-ink hover:bg-sunken'}"
			>
				{a}
			</button>
		{/each}
	</div>

	<Card>
		{#if shown.length}
			<TaskList tasks={shown} targetId={target.id} />
		{:else}
			<p class="text-ink-muted">Nothing to do for {assignee}.</p>
		{/if}
	</Card>
</div>
