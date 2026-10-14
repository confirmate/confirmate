<script lang="ts">
	import RequirementList from '$lib/guided/components/RequirementList.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<RequirementList target={data.target!} />
