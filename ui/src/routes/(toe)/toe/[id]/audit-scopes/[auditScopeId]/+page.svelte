<script lang="ts">
	import SectionHeader from '$lib/components/ui/SectionHeader.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import CategorySection from '$lib/components/toe/CategorySection.svelte';
	import { evaluationClient } from '$lib/api/client';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	// Client-side tracking of evaluation running state.
	// The backend doesn't expose this as persistent state, so we initialise to
	// false and flip it optimistically on start/stop.
	let evaluationRunning = $state(false);
	let evaluationBusy = $state(false);

	async function startEvaluation() {
		evaluationBusy = true;
		try {
			await evaluationClient().POST('/v1/evaluation/evaluate/{auditScopeId}/start', {
				params: { path: { auditScopeId: data.auditScope.id } }
			});
			evaluationRunning = true;
		} finally {
			evaluationBusy = false;
		}
	}

	async function stopEvaluation() {
		evaluationBusy = true;
		try {
			await evaluationClient().POST('/v1/evaluation/evaluate/{auditScopeId}/stop', {
				params: { path: { auditScopeId: data.auditScope.id } }
			});
			evaluationRunning = false;
		} finally {
			evaluationBusy = false;
		}
	}
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

				{#if evaluationRunning}
					<span class="inline-flex items-center gap-1.5 rounded-full bg-green-50 px-2.5 py-1 text-xs font-medium text-green-700">
						<span class="h-1.5 w-1.5 animate-pulse rounded-full bg-green-500"></span>
						Evaluation running
					</span>
					<Button variant="danger" size="sm" onclick={stopEvaluation} disabled={evaluationBusy}>
						Stop Evaluation
					</Button>
				{:else}
					<Button variant="secondary" size="sm" onclick={startEvaluation} disabled={evaluationBusy}>
						Start Evaluation
					</Button>
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
