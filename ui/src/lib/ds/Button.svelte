<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		variant = 'primary',
		size = 'md',
		href,
		onclick,
		disabled = false,
		type = 'button',
		children
	}: {
		variant?: 'primary' | 'quiet' | 'danger';
		size?: 'sm' | 'md';
		href?: string;
		onclick?: () => void;
		disabled?: boolean;
		type?: 'button' | 'submit';
		children: Snippet;
	} = $props();

	const cls = $derived(
		[
			'inline-flex items-center justify-center gap-2 rounded-control font-body font-medium no-underline transition-colors',
			'focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand',
			size === 'sm' ? 'min-h-10 px-4 text-sm' : 'min-h-11 px-5 text-[0.9375rem]',
			variant === 'primary'
				? 'bg-brand text-on-brand hover:bg-brand-strong'
				: variant === 'danger'
					? 'bg-bad text-on-brand hover:opacity-90'
					: 'border border-line-strong bg-surface text-ink hover:bg-sunken',
			disabled ? 'pointer-events-none opacity-50' : ''
		].join(' ')
	);
</script>

{#if href}
	<a {href} class={cls} aria-disabled={disabled} tabindex={disabled ? -1 : undefined}>
		{@render children()}
	</a>
{:else}
	<button {type} {onclick} {disabled} class={cls}>{@render children()}</button>
{/if}
