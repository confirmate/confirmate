<script lang="ts">
	import type { Snippet } from 'svelte';
	import { borderTone, type Tone } from './tones';

	let {
		accent,
		padding = 'md',
		children
	}: { accent?: Tone; padding?: 'sm' | 'md' | 'lg'; children: Snippet } = $props();

	const pad = { sm: 'p-4', md: 'p-6', lg: 'p-7' };
</script>

<div
	class="flex flex-col gap-3 rounded-card border border-line bg-surface {pad[padding]} {accent
		? `border-l-4 ${borderTone[accent]}`
		: ''}"
>
	{@render children()}
</div>
