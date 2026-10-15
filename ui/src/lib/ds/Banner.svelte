<script lang="ts">
	import type { Snippet } from 'svelte';
	import { softTone, type Tone } from './tones';

	let {
		tone = 'brand',
		title,
		children,
		actions
	}: { tone?: Tone; title?: string; children: Snippet; actions?: Snippet } = $props();
</script>

<div class="flex flex-wrap items-start justify-between gap-4 rounded-control px-5 py-4 font-body {softTone[tone]}">
	<div class="flex max-w-3xl flex-col gap-1">
		{#if title}<strong class="text-[1.0625rem]">{title}</strong>{/if}
		<div>{@render children()}</div>
	</div>
	{#if actions}<div class="flex gap-2.5">{@render actions()}</div>{/if}
</div>
