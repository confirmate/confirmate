<script lang="ts">
	import type { Snippet } from 'svelte';
	import { softTone, type Tone } from './tones';

	let {
		tone = 'neutral',
		size = 'md',
		children
	}: { tone?: Tone; size?: 'sm' | 'md'; children: Snippet } = $props();
</script>

<span
	class="inline-flex items-center gap-1.5 whitespace-nowrap font-body font-medium {softTone[tone]} {size ===
	'sm'
		? 'rounded-pill px-2 py-0.5 text-xs font-semibold'
		: 'h-[1.625rem] rounded-pill px-2.5 text-[0.8125rem]'}"
>
	{@render children()}
</span>
