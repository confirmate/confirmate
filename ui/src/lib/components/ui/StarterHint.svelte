<!--
@component StarterHint

An empty state component that encourages users to take action.
Displays an icon, message, and call-to-action snippet.
-->
<script lang="ts">
	import type { Snippet } from 'svelte';

	/**
	 * Props for the StarterHint component
	 */
	interface Props {
		/** Icon component to display */
		icon: any;
		/** Type of content that is missing (e.g., 'assessments', 'users') */
		type: string;
		/** Call-to-action component (usually a button or link) */
		component: Snippet;
	}

	let { icon, type, component }: Props = $props();
</script>

<div class="mt-20 text-center">
	{#if icon}
		{@const Icon = icon}
		<Icon class="mx-auto h-12 w-12 text-gray-400" />
	{/if}
	<h3 class="mt-2 text-sm font-semibold text-gray-900">No {type}</h3>
	<span class="mt-1 text-sm text-gray-500">
		Get started by launching a {@render component()}
	</span>
</div>
