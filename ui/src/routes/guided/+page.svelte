<script lang="ts">
	import TargetList from '$lib/guided/components/TargetList.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<TargetList targets={data.targets} />
