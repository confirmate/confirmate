<script lang="ts">
	import WizardStepPage from '$lib/guided/components/WizardStepPage.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<WizardStepPage profile={data.profile} step={data.step} />
