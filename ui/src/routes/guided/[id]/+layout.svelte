<script lang="ts">
	import TargetShell from '$lib/guided/components/TargetShell.svelte';
	import TargetUnavailable from '$lib/guided/components/TargetUnavailable.svelte';
	import type { LayoutProps } from './$types';

	let { children, data }: LayoutProps = $props();
</script>

{#if data.target}
	<TargetShell target={data.target}>{@render children()}</TargetShell>
{:else}
	<TargetUnavailable />
{/if}
