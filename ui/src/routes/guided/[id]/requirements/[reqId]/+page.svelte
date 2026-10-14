<script lang="ts">
	import RequirementView from '$lib/guided/components/RequirementView.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<RequirementView target={data.target!} requirement={data.requirement} />
