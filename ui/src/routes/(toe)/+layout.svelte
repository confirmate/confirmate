<script lang="ts">
	import GlobalTopBar from '$lib/components/nav/GlobalTopBar.svelte';
	import Sidebar from '$lib/components/nav/Sidebar.svelte';
	import type { LayoutProps } from './$types';

	let { children, data }: LayoutProps = $props();
</script>

<div class="flex h-screen flex-col">
	<GlobalTopBar />
	<div class="flex min-h-0 flex-1">
		<Sidebar services={data.services} />
		<main class="flex-1 overflow-y-auto">
			<div class="px-8 py-8">
				{@render children()}
			</div>
		</main>
	</div>
</div>
