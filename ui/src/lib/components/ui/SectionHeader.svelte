<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		title,
		description,
		actions
	}: {
		title: string;
		description?: string;
		actions?: Snippet;
	} = $props();
</script>

<div class="flex items-center justify-between">
	<div>
		<h2 class="text-lg font-semibold text-gray-900">{title}</h2>
		{#if description}
			<p class="mt-1 text-sm text-gray-500">{description}</p>
		{/if}
	</div>
	{#if actions}
		<div class="ml-4 shrink-0">{@render actions()}</div>
	{/if}
</div>
