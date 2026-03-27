<script lang="ts">
	import SectionHeader from '$lib/components/ui/SectionHeader.svelte';
	import AccessControl from '$lib/components/ui/AccessControl.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<div>
	<SectionHeader
		title="Access"
		description="Users with access to this target of evaluation."
	/>

	<div class="mt-6 max-w-xl">
		<AccessControl
			readers={data.toe.readers}
			contributors={data.toe.contributors}
			admins={data.toe.admins}
		/>
	</div>
</div>
