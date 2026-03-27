<script lang="ts">
	import SectionHeader from '$lib/components/ui/SectionHeader.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import CategorySection from '$lib/components/toe/CategorySection.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<div>
	<a
		href="/toe/{data.auditScope.targetOfEvaluationId}/audit-scopes/"
		class="text-sm text-gray-500 hover:text-gray-700"
	>
		← Back to Audit Scopes
	</a>

	<div class="mt-4">
		<SectionHeader title={data.auditScope.name}>
			{#snippet actions()}
				{#if data.catalog}
					<span class="inline-flex items-center rounded-full bg-blue-50 px-2.5 py-1 text-xs font-medium text-blue-700">
						{data.catalog.name}
					</span>
				{/if}
				{#if data.auditScope.assuranceLevel}
					<span class="inline-flex items-center rounded-full bg-gray-100 px-2.5 py-1 text-xs font-medium text-gray-600">
						{data.auditScope.assuranceLevel}
					</span>
				{/if}
			{/snippet}
		</SectionHeader>
	</div>

	<div class="mt-6 space-y-3">
		{#if data.catalog?.categories?.length}
			{#each data.catalog.categories as category}
				<CategorySection {category} controls={data.controlsByCategory[category.name] ?? []} />
			{/each}
		{:else}
			<EmptyState title="No controls found" description="This catalog has no categories or controls defined." />
		{/if}
	</div>
</div>
