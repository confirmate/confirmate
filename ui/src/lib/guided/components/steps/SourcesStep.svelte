<script lang="ts">
	import { Button, Card, Chip } from '$lib/ds';
	import type { FrameworkProfile } from '../../profile';

	let { profile }: { profile: FrameworkProfile } = $props();
</script>

<div class="flex flex-col gap-4">
	{#each profile.sources as source (source.id)}
		<Card>
			<div class="flex flex-wrap items-center justify-between gap-3">
				<h2 class="text-lg font-semibold">{source.title}</h2>
				<Chip size="sm" tone={source.state === 'connected' ? 'ok' : 'neutral'}>{source.stateLabel}</Chip>
			</div>
			<p class="text-ink-soft">{source.text}</p>
			{#if source.detail}<p class="text-[0.8125rem] text-ink-muted">{source.detail}</p>{/if}
			{#if source.snippet}
				<pre class="m-0 overflow-x-auto rounded-control bg-sunken px-4 py-3.5 font-code text-[0.8125rem]">{source.snippet.replaceAll('\\n', '\n')}</pre>
			{/if}
			{#if source.actions}
				<div class="flex flex-wrap gap-2.5">
					{#each source.actions as action (action)}<Button variant="quiet" size="sm">{action}</Button>{/each}
				</div>
			{/if}
		</Card>
	{/each}
</div>
