<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		label,
		hint,
		id,
		children
	}: { label: string; hint?: string; id: string; children: Snippet } = $props();
</script>

<div class="flex flex-col gap-1.5">
	<label for={id} class="font-body font-medium text-ink">{label}</label>
	{@render children()}
	{#if hint}<span class="text-[0.8125rem] text-ink-muted">{hint}</span>{/if}
</div>
