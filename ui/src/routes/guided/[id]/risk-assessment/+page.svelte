<script lang="ts">
	import AssessmentPage from '$lib/guided/components/AssessmentPage.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<AssessmentPage target={data.target!} assessment={data.assessment} />
