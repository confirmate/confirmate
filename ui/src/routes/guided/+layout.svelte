<script lang="ts">
	import { onMount } from 'svelte';
	import TopBar from '$lib/guided/components/TopBar.svelte';
	import { theme } from '$lib/theme/theme.svelte';
	import type { LayoutProps } from './$types';

	let { children, data }: LayoutProps = $props();

	onMount(() => theme.apply());
</script>

<div class="min-h-screen bg-canvas font-body text-[0.9375rem] leading-normal text-ink">
	<TopBar currentUser={data.currentUser} />
	{@render children()}
</div>
