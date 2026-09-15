<script lang="ts">
	import SectionHeader from '$lib/components/ui/SectionHeader.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<div>
	<SectionHeader title="Targets of Evaluation" description="All registered targets of evaluation in the system." />

	{#if data.services.length > 0}
		<ul class="mt-6 space-y-2">
			{#each data.services as service}
				<li>
					<a
						href="/toe/{service.id}/audit-scopes/"
						class="block rounded-lg border border-gray-200 bg-white p-4 hover:border-confirmate hover:bg-blue-50"
					>
						<div class="font-medium text-gray-900">{service.name}</div>
						{#if service.description}
							<div class="mt-1 text-sm text-gray-500">{service.description}</div>
						{/if}
					</a>
				</li>
			{/each}
		</ul>
	{:else}
		<div class="mt-8 rounded-lg border border-dashed border-gray-300 p-12 text-center">
			<p class="text-gray-500">No targets of evaluation yet.</p>
			<p class="mt-2 text-sm text-gray-400">Targets are created automatically when the server starts with <code>--create-default-target-of-evaluation</code>.</p>
		</div>
	{/if}
</div>
