<script lang="ts">
	import { Card, FilterChips, PageHeading } from '$lib/ds';
	import type { TargetDetail } from '../source';
	import TaskList from './TaskList.svelte';

	let { target }: { target: TargetDetail } = $props();

	const assignees = $derived(['Everyone', ...new Set(target.tasks.map((t) => t.assignee))]);
	let assignee = $state('Everyone');
	const shown = $derived(assignee === 'Everyone' ? target.tasks : target.tasks.filter((t) => t.assignee === assignee));
</script>

<div class="flex max-w-4xl flex-col gap-6">
	<PageHeading
		title="To do"
		intro="Everything that needs a person: failing checks, missing documents and open deadlines."
	/>
	<FilterChips label="Filter by person" options={assignees} bind:value={assignee} />
	<Card>
		{#if shown.length}
			<TaskList tasks={shown} targetId={target.id} />
		{:else}
			<p class="text-ink-muted">Nothing to do for {assignee}.</p>
		{/if}
	</Card>
</div>
