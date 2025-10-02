<!--
@component Card

A flexible card component for containing content with optional header and actions.
Supports title, description, main content, and action snippets.
-->
<script lang="ts">
	import type { Snippet } from 'svelte';

	/**
	 * Props for the Card component
	 */
	interface Props {
		/** Optional title for the card header */
		title?: string;
		/** Optional description text below the title */
		description?: string;
		/** Main content of the card */
		children: Snippet;
		/** Optional action buttons rendered at the bottom */
		actions?: Snippet;
		/** Additional CSS classes to apply */
		class?: string;
	}

	let { title, description, children, actions, class: className = '' }: Props = $props();
</script>

<div class="overflow-hidden rounded-lg bg-white shadow ring-1 ring-gray-900/5 {className}">
	{#if title || description || actions}
		<div class="px-4 py-5 sm:p-6">
			{#if title || description}
				<div class="mb-4">
					{#if title}
						<h3 class="text-base leading-6 font-semibold text-gray-900">
							{title}
						</h3>
					{/if}
					{#if description}
						<p class="mt-1 text-sm text-gray-500">
							{description}
						</p>
					{/if}
				</div>
			{/if}
			{@render children()}
			{#if actions}
				<div class="mt-4 flex justify-end space-x-3">
					{@render actions()}
				</div>
			{/if}
		</div>
	{:else}
		<div class="px-4 py-5 sm:p-6">
			{@render children()}
		</div>
	{/if}
</div>
