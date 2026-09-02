<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		title,
		subtitle,
		open = $bindable(false),
		children
	}: {
		title: string;
		subtitle?: string;
		open?: boolean;
		children: Snippet;
	} = $props();
</script>

<div class="overflow-hidden rounded-lg border border-gray-200 bg-white">
	<button
		class="flex w-full items-center justify-between px-6 py-4 text-left hover:bg-gray-50"
		onclick={() => (open = !open)}
	>
		<div class="flex items-center gap-2">
			<svg
				class="h-5 w-5 text-gray-400 transition-transform duration-200 {open ? 'rotate-90' : ''}"
				fill="none"
				viewBox="0 0 24 24"
				stroke="currentColor"
			>
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
			</svg>
			<span class="text-sm font-semibold text-gray-900">{title}</span>
		</div>
		{#if subtitle}
			<span class="text-xs text-gray-400">{subtitle}</span>
		{/if}
	</button>

	{#if open}
		<div class="border-t border-gray-100 px-6 pb-6">
			{@render children()}
		</div>
	{/if}
</div>