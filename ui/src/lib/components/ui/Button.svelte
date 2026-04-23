<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		variant = 'primary',
		size = 'md',
		href,
		onclick,
		disabled = false,
		children
	}: {
		variant?: 'primary' | 'secondary' | 'danger';
		size?: 'sm' | 'md';
		href?: string;
		onclick?: () => void;
		disabled?: boolean;
		children: Snippet;
	} = $props();

	const cls = $derived(
		[
			'inline-flex items-center gap-2 rounded-md font-medium transition-colors',
			size === 'sm' ? 'px-3 py-1.5 text-sm' : 'px-4 py-2 text-sm',
			variant === 'primary'
				? 'bg-confirmate text-white hover:bg-confirmate-light disabled:opacity-50'
				: variant === 'danger'
					? 'bg-red-50 text-red-700 hover:bg-red-100'
					: 'border border-gray-300 bg-white text-gray-700 hover:bg-gray-50'
		].join(' ')
	);
</script>

{#if href}
	<a {href} class={cls}>{@render children()}</a>
{:else}
	<button {onclick} {disabled} class={cls}>{@render children()}</button>
{/if}
