<script lang="ts">
	import type { Snippet } from 'svelte';
	import { borderTone, type Tone } from './tones';

	let {
		accent,
		padding = 'md',
		children
	}: { accent?: Tone; padding?: 'sm' | 'md' | 'lg'; children: Snippet } = $props();

	const pad = { sm: 'p-3.5', md: 'p-(--cf-card-pad)', lg: 'p-[calc(var(--cf-card-pad)*1.25)]' };
</script>

<div
	class="flex flex-col gap-3 rounded-card border border-line bg-surface {pad[padding]} {accent
		? `border-l-4 ${borderTone[accent]}`
		: ''}"
>
	{@render children()}
</div>
