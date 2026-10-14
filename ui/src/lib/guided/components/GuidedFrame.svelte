<script lang="ts">
	import type { Snippet } from 'svelte';
	import { onMount } from 'svelte';
	import type { SchemaUser } from '$lib/api/openapi/orchestrator';
	import { theme } from '$lib/theme/theme.svelte';
	import TopBar from './TopBar.svelte';

	let { currentUser = null, children }: { currentUser?: SchemaUser | null; children: Snippet } = $props();

	onMount(() => theme.apply());
</script>

<div class="min-h-screen bg-canvas font-body text-[0.9375rem] leading-normal text-ink">
	<TopBar {currentUser} />
	{@render children()}
</div>
