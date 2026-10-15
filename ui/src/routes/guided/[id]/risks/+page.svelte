<script lang="ts">
	import RiskList from '$lib/guided/components/RiskList.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<RiskList target={data.target!} risks={data.risks} />
