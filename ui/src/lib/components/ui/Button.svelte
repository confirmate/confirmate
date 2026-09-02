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
					: 'border border-gray-300 bg-white text-gray-700 hover:bg-gray-50',
			// The disabled: pseudo-class only applies to <button>, not <a>, so apply the same
			// dimming explicitly for the disabled+href case.
			disabled ? 'opacity-50' : ''
		].join(' ')
	);
</script>

{#if href}
	<a
		href={disabled ? undefined : href}
		aria-disabled={disabled}
		tabindex={disabled ? -1 : undefined}
		class={cls}
		onclick={(e) => {
			if (disabled) e.preventDefault();
		}}
	>{@render children()}</a>
{:else}
	<button {onclick} {disabled} class={cls}>{@render children()}</button>
{/if}
