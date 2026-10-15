<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		title,
		intro,
		above,
		actions
	}: { title: string; intro?: string; above?: Snippet; actions?: Snippet } = $props();
</script>

<div class="flex flex-wrap items-end justify-between gap-6 font-body">
	<div class="flex max-w-3xl flex-col gap-2">
		{#if above}{@render above()}{/if}
		<h1 class="font-display text-title font-(--cf-display-weight) leading-tight text-ink">{title}</h1>
		{#if intro}<p class="text-[1.0625rem] text-ink-muted">{intro}</p>{/if}
	</div>
	{#if actions}<div class="flex flex-wrap gap-2.5">{@render actions()}</div>{/if}
</div>
