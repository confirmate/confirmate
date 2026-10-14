<script lang="ts">
	import { Button, Card, Chip, type Tone } from '$lib/ds';
	import type { Check } from '../source';

	let { check }: { check: Check } = $props();

	const accent: Record<Check['status'], Tone> = { met: 'ok', 'not-met': 'bad', open: 'neutral' };
	const label: Record<Check['status'], string> = { met: 'Met', 'not-met': 'Not met', open: 'Open' };
	const chipTone: Record<Check['status'], Tone> = { met: 'ok', 'not-met': 'bad', open: 'neutral' };
</script>

<Card accent={accent[check.status]}>
	<div class="flex flex-wrap items-start justify-between gap-3">
		<h2 class="text-lg font-semibold">{check.title}</h2>
		<span class="flex gap-2">
			<Chip size="sm">{check.kind === 'automatic' ? 'Automatic' : 'Statement'}</Chip>
			<Chip size="sm" tone={chipTone[check.status]}>{label[check.status]}</Chip>
		</span>
	</div>
	<p class="text-ink-soft">{check.text}</p>
	{#if check.code}<p class="rounded-control bg-sunken px-3.5 py-2.5 font-code text-[0.8125rem]">{check.code}</p>{/if}
	{#if check.note}<p class="text-[0.8125rem] text-ink-muted">{check.note}</p>{/if}
	{#if check.actions.length}
		<div class="flex flex-wrap gap-2.5">
			{#each check.actions as action (action.label)}
				<Button size="sm" variant={action.primary ? 'primary' : 'quiet'}>{action.label}</Button>
			{/each}
		</div>
	{/if}
</Card>
