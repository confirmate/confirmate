<script lang="ts">
	import { Card } from '$lib/ds';
	import type { SourceStatus } from '../source';

	let { sources }: { sources: SourceStatus[] } = $props();

	const tone = { ok: 'text-ok-ink', partial: 'text-warn-ink', none: 'text-ink-muted' };
</script>

<Card>
	<h2 class="text-lg font-semibold">Sources</h2>
	<ul class="m-0 flex list-none flex-col p-0 text-sm">
		{#each sources as s (s.name)}
			<li class="flex justify-between border-t border-line py-2">
				<span>{s.name}</span>
				<span class={tone[s.state]}>{s.detail}</span>
			</li>
		{/each}
	</ul>
</Card>
