<script lang="ts">
	import GuidedFrame from '$lib/guided/components/GuidedFrame.svelte';
	import type { LayoutProps } from './$types';

	let { children, data }: LayoutProps = $props();
</script>

<GuidedFrame currentUser={data.currentUser}>{@render children()}</GuidedFrame>
