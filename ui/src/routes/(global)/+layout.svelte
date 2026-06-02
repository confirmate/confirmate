<script lang="ts">
	import GlobalTopBar from '$lib/components/nav/GlobalTopBar.svelte';
	import type { LayoutProps } from './$types';

	let { children }: LayoutProps = $props();
</script>

<div class="flex h-screen flex-col">
	<GlobalTopBar />
	<main class="flex-1 overflow-y-auto">
		<div class="mx-auto max-w-4xl px-8 py-8">
			{@render children()}
		</div>
	</main>
</div>
