<!--
@component BelowHeader

A simple component for displaying descriptive text below page headers.
Typically used for additional context or instructions.
-->
<script lang="ts">
	import type { Snippet } from 'svelte';

	/**
	 * Props for the BelowHeader component
	 */
	interface Props {
		/** Content to display below the header */
		children: Snippet;
	}

	let { children }: Props = $props();
</script>

<div class="mt-4 text-sm text-gray-700">
	{@render children()}
</div>
