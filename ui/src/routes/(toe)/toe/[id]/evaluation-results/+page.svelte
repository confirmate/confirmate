<script lang="ts">
	import type { SchemaEvaluationResult } from '$lib/api/openapi/orchestrator';
	import SectionHeader from '$lib/components/ui/SectionHeader.svelte';
	import { Icon } from '@steeze-ui/svelte-icon';
	import { CheckCircle, XCircle, Clock, QuestionMarkCircle } from '@steeze-ui/heroicons';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const statusColors: Record<string, string> = {
		EVALUATION_STATUS_COMPLIANT: 'text-green-600',
		EVALUATION_STATUS_COMPLIANT_MANUALLY: 'text-green-600',
		EVALUATION_STATUS_NOT_COMPLIANT: 'text-red-600',
		EVALUATION_STATUS_NOT_COMPLIANT_MANUALLY: 'text-red-600',
		EVALUATION_STATUS_PENDING: 'text-yellow-600',
		EVALUATION_STATUS_UNSPECIFIED: 'text-gray-500'
	};

	const statusLabels: Record<string, string> = {
		EVALUATION_STATUS_COMPLIANT: 'Compliant',
		EVALUATION_STATUS_COMPLIANT_MANUALLY: 'Compliant (Manual)',
		EVALUATION_STATUS_NOT_COMPLIANT: 'Non-Compliant',
		EVALUATION_STATUS_NOT_COMPLIANT_MANUALLY: 'Non-Compliant (Manual)',
		EVALUATION_STATUS_PENDING: 'Pending',
		EVALUATION_STATUS_UNSPECIFIED: 'Unknown'
	};

	const compliantCount = $derived(data.evaluationResults.filter((r) =>
		r.status === 'EVALUATION_STATUS_COMPLIANT' || r.status === 'EVALUATION_STATUS_COMPLIANT_MANUALLY'
	).length);

	const nonCompliantCount = $derived(data.evaluationResults.filter((r) =>
		r.status === 'EVALUATION_STATUS_NOT_COMPLIANT' || r.status === 'EVALUATION_STATUS_NOT_COMPLIANT_MANUALLY'
	).length);

	const pendingCount = $derived(data.evaluationResults.filter((r) =>
		r.status === 'EVALUATION_STATUS_PENDING'
	).length);
</script>

<div>
	<SectionHeader
		title="Evaluation Results"
		description="Evaluation results based on assessment results for this target of evaluation."
	/>

	<div class="mt-6 grid grid-cols-4 gap-4">
		<div class="rounded-lg border border-gray-200 bg-white px-5 py-4">
			<div class="text-sm font-medium text-gray-500">Total</div>
			<div class="mt-1 text-2xl font-bold text-confirmate">{data.evaluationResults.length}</div>
		</div>
		<div class="rounded-lg border border-gray-200 bg-white px-5 py-4">
			<div class="text-sm font-medium text-gray-500">Compliant</div>
			<div class="mt-1 text-2xl font-bold text-green-600">{compliantCount}</div>
		</div>
		<div class="rounded-lg border border-gray-200 bg-white px-5 py-4">
			<div class="text-sm font-medium text-gray-500">Non-Compliant</div>
			<div class="mt-1 text-2xl font-bold text-red-600">{nonCompliantCount}</div>
		</div>
		<div class="rounded-lg border border-gray-200 bg-white px-5 py-4">
			<div class="text-sm font-medium text-gray-500">Pending</div>
			<div class="mt-1 text-2xl font-bold text-yellow-600">{pendingCount}</div>
		</div>
	</div>

	<div class="mt-6 overflow-hidden rounded-lg border border-gray-200">
		<table class="min-w-full divide-y divide-gray-200">
			<thead class="bg-gray-50">
				<tr>
					<th scope="col" class="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500">
						Status
					</th>
					<th scope="col" class="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500">
						Control
					</th>
					<th scope="col" class="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500">
						Catalog
					</th>
					<th scope="col" class="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500">
						Category
					</th>
					<th scope="col" class="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500">
						Assessment Results
					</th>
					<th scope="col" class="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500">
						Date
					</th>
				</tr>
			</thead>
			<tbody class="divide-y divide-gray-200 bg-white">
				{#if data.evaluationResults.length === 0}
					<tr>
						<td colspan="6" class="px-4 py-8 text-center text-sm text-gray-500">
							No evaluation results found.
						</td>
					</tr>
				{:else}
					{#each data.evaluationResults as result}
						<tr class="hover:bg-gray-50">
							<td class="whitespace-nowrap px-4 py-3">
								<div class="flex items-center gap-1.5 {statusColors[result.status]}">
									{#if result.status === 'EVALUATION_STATUS_COMPLIANT' || result.status === 'EVALUATION_STATUS_COMPLIANT_MANUALLY'}
										<Icon src={CheckCircle} class="h-4 w-4" />
									{:else if result.status === 'EVALUATION_STATUS_NOT_COMPLIANT' || result.status === 'EVALUATION_STATUS_NOT_COMPLIANT_MANUALLY'}
										<Icon src={XCircle} class="h-4 w-4" />
									{:else if result.status === 'EVALUATION_STATUS_PENDING'}
										<Icon src={Clock} class="h-4 w-4" />
									{:else}
										<Icon src={QuestionMarkCircle} class="h-4 w-4" />
									{/if}
									<span class="text-sm font-medium">{statusLabels[result.status]}</span>
								</div>
							</td>
							<td class="whitespace-nowrap px-4 py-3">
								<span class="text-sm font-mono text-gray-900">{result.controlId ?? '—'}</span>
							</td>
							<td class="whitespace-nowrap px-4 py-3">
								<span class="text-sm text-gray-500">{result.controlCatalogId ?? '—'}</span>
							</td>
							<td class="whitespace-nowrap px-4 py-3">
								<span class="text-sm text-gray-500">{result.controlCategoryName ?? '—'}</span>
							</td>
							<td class="whitespace-nowrap px-4 py-3">
								<span class="text-sm text-gray-500">{result.assessmentResultIds?.length ?? 0}</span>
							</td>
							<td class="whitespace-nowrap px-4 py-3">
								<span class="text-sm text-gray-500">
									{result.timestamp ? new Date(result.timestamp).toLocaleString() : '—'}
								</span>
							</td>
						</tr>
					{/each}
				{/if}
			</tbody>
		</table>
	</div>
</div>