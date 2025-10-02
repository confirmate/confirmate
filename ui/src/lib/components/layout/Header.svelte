<!--
@component Header

A page header component with title, description, optional icon, and action buttons.
Supports custom actions via snippets and emits save/remove events.
-->
<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import Button from '../ui/Button.svelte';
	import type { Snippet } from 'svelte';

	/**
	 * Props for the Header component
	 */
	interface Props {
		/** Main title text */
		title: string;
		/** Optional description text below the title */
		description?: string;
		/** Whether to show default action buttons (Save/Delete) @default true */
		showActions?: boolean;
		/** Icon URL or false to disable icon. Empty string shows initials @default '' */
		icon?: string | false;
		/** Custom content to render in the actions area */
		children?: Snippet;
	}

	let { title, description = '', showActions = true, icon = '', children }: Props = $props();

	const dispatch = createEventDispatcher<{
		remove: void;
		save: void;
	}>();
</script>

<div class="border-b border-gray-200 pb-5 sm:flex sm:items-center sm:justify-between">
	<div class="flex items-start space-x-5">
		{#if icon !== false}
			<div class="flex-shrink-0">
				<div class="relative">
					{#if icon}
						<img class="h-16 w-16 rounded-full" src={icon} alt="" />
					{:else}
						<div class="flex h-16 w-16 items-center justify-center rounded-full bg-gray-300">
							<span class="text-2xl font-semibold text-gray-600">
								{title.charAt(0).toUpperCase()}
							</span>
						</div>
					{/if}
				</div>
			</div>
		{/if}
		<div>
			{#if title}
				<h1 class="text-2xl font-bold text-gray-900">{title}</h1>
			{:else}
				<h1 class="text-2xl font-bold text-gray-400">Enter a name</h1>
			{/if}
			{#if description}
				<p class="text-sm font-medium text-gray-500">
					{description}
				</p>
			{/if}
		</div>
	</div>

	{#if showActions || children}
		<div
			class="mt-6 flex flex-col-reverse justify-stretch space-y-4 space-y-reverse sm:flex-row-reverse sm:justify-end sm:space-y-0 sm:space-x-3 sm:space-x-reverse md:mt-0 md:flex-row md:space-x-3"
		>
			{#if showActions}
				<Button variant="danger" onclick={() => dispatch('remove')}>Delete</Button>
				<Button onclick={() => dispatch('save')}>Save</Button>
			{/if}
			{#if children}
				{@render children()}
			{/if}
		</div>
	{/if}
</div>
