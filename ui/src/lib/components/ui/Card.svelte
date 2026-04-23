<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		padding = 'md',
		element,
		children
	}: {
		padding?: 'sm' | 'md' | 'lg';
		element?: HTMLDivElement | undefined;
		children: Snippet;
	} = $props();

	const paddingClass = $derived(
		padding === 'sm' ? 'px-4 py-3' : padding === 'lg' ? 'px-8 py-6' : 'px-5 py-4'
	);
</script>

<div bind:this={element} class="overflow-hidden rounded-lg border border-gray-200 bg-white {paddingClass}">
	{@render children()}
</div>