<script lang="ts">
	import TargetOverview from '$lib/guided/components/TargetOverview.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<!-- The layout only renders its children for a known target. -->
<TargetOverview target={data.target!} />
