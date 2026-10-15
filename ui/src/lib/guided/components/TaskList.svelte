<script lang="ts">
	import { resolve } from '$app/paths';
	import type { Task } from '../source';
	import StatusTag from './StatusTag.svelte';

	let { tasks, targetId }: { tasks: Task[]; targetId: string } = $props();
</script>

<ul class="m-0 flex list-none flex-col p-0">
	{#each tasks as task (task.id)}
		<li class="flex items-start gap-4 border-t border-line py-3.5">
			<span class="mt-0.5"><StatusTag status={task.urgency} /></span>
			<span class="flex flex-1 flex-col">
				{#if task.requirementId}
					<a href={resolve(`/guided/${targetId}/requirements/${task.requirementId}/`)} class="font-medium no-underline">{task.title}</a>
				{:else}
					<span class="font-medium">{task.title}</span>
				{/if}
				<span class="text-[0.8125rem] text-ink-muted">{task.detail}</span>
			</span>
			<span class="whitespace-nowrap text-[0.8125rem] text-ink-soft">{task.assignee}{task.due ? `, ${task.due}` : ''}</span>
		</li>
	{/each}
</ul>
