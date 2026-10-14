<script lang="ts">
	import { Card, StackedProgress } from '$lib/ds';
	import type { PartProgress, Progress } from '../source';

	let { progress, parts }: { progress: Progress; parts: PartProgress[] } = $props();
</script>

<Card>
	<div class="flex items-baseline gap-2.5">
		<span class="font-display text-figure font-(--cf-display-weight) leading-none">{progress.met}</span>
		<span class="text-ink-soft">of {progress.total} requirements met</span>
	</div>
	<StackedProgress
		total={progress.total}
		segments={[
			{ value: progress.met, tone: 'ok', label: 'met' },
			{ value: progress.attention, tone: 'bad', label: 'need attention' }
		]}
	/>
	<ul class="m-0 flex list-none flex-col p-0 text-sm">
		{#each parts as part (part.id)}
			<li class="flex justify-between border-t border-line py-2">
				<span>{part.label}</span>
				<span class="text-ink-soft">{part.met} of {part.total} met</span>
			</li>
		{/each}
	</ul>
</Card>
