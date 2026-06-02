<script lang="ts">
	import type { SchemaAssessmentResult } from '$lib/api/openapi/orchestrator';
	import SectionHeader from '$lib/components/ui/SectionHeader.svelte';
	import DataTable from '$lib/components/ui/DataTable.svelte';
	import { Icon } from '@steeze-ui/svelte-icon';
	import { CheckCircle, XCircle } from '@steeze-ui/heroicons';
	import { page } from '$app/state';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const toeId = $derived(page.params.id);

	const columns = $derived([
		{
			key: 'status',
			label: 'Status',
			sortable: true,
			class: 'w-32',
			render: (r: SchemaAssessmentResult) => r.compliant ? 'Compliant' : 'Non-Compliant'
		},
		{
			key: 'metricId',
			label: 'Metric',
			sortable: true,
			render: (r: SchemaAssessmentResult) => r.metricId
		},
		{
			key: 'resourceId',
			label: 'Resource',
			sortable: true,
			render: (r: SchemaAssessmentResult) => r.resourceId
		},
		{
			key: 'toolId',
			label: 'Tool',
			sortable: true,
			render: (r: SchemaAssessmentResult) => r.toolId
		},
		{
			key: 'comment',
			label: 'Comment',
			class: 'max-w-xs',
			render: (r: SchemaAssessmentResult) => r.complianceComment ?? ''
		},
		{
			key: 'createdAt',
			label: 'Date',
			sortable: true,
			render: (r: SchemaAssessmentResult) => new Date(r.createdAt).toLocaleDateString()
		}
	]);

	const compliantCount = $derived(data.assessmentResults.filter((r) => r.compliant).length);
	const nonCompliantCount = $derived(data.assessmentResults.filter((r) => !r.compliant).length);
</script>

<div>
	<SectionHeader
		title="Assessment Results"
		description="Assessment results from automated and manual evidence assessment for this target of evaluation."
	/>

	<div class="mt-6 grid grid-cols-3 gap-4">
		<div class="rounded-lg border border-gray-200 bg-white px-5 py-4">
			<div class="text-sm font-medium text-gray-500">Total</div>
			<div class="mt-1 text-2xl font-bold text-confirmate">{data.assessmentResults.length}</div>
		</div>
		<div class="rounded-lg border border-gray-200 bg-white px-5 py-4">
			<div class="text-sm font-medium text-gray-500">Compliant</div>
			<div class="mt-1 text-2xl font-bold text-green-600">{compliantCount}</div>
		</div>
		<div class="rounded-lg border border-gray-200 bg-white px-5 py-4">
			<div class="text-sm font-medium text-gray-500">Non-Compliant</div>
			<div class="mt-1 text-2xl font-bold text-red-600">{nonCompliantCount}</div>
		</div>
	</div>

	<div class="mt-6">
		<div class="overflow-hidden rounded-lg border border-gray-200">
			<table class="min-w-full divide-y divide-gray-200">
				<thead class="bg-gray-50">
					<tr>
						<th scope="col" class="w-32 px-4 py-3 text-left text-xs font-medium uppercase text-gray-500">
							Status
						</th>
						<th scope="col" class="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500">
							Metric
						</th>
						<th scope="col" class="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500">
							Resource
						</th>
						<th scope="col" class="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500">
							Tool
						</th>
						<th scope="col" class="max-w-xs px-4 py-3 text-left text-xs font-medium uppercase text-gray-500">
							Comment
						</th>
						<th scope="col" class="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500">
							Date
						</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-200 bg-white">
					{#each data.assessmentResults as result}
						<tr class="hover:bg-gray-50">
							<td class="whitespace-nowrap px-4 py-3">
								{#if result.compliant}
									<div class="flex items-center gap-1.5 text-green-600">
										<Icon src={CheckCircle} class="h-4 w-4" />
										<span class="text-sm font-medium">Compliant</span>
									</div>
								{:else}
									<div class="flex items-center gap-1.5 text-red-600">
										<Icon src={XCircle} class="h-4 w-4" />
										<span class="text-sm font-medium">Non-Compliant</span>
									</div>
								{/if}
							</td>
							<td class="whitespace-nowrap px-4 py-3">
								<a href="/toe/{toeId}/assessment-results/{result.id}" class="text-sm font-mono text-gray-900 hover:text-confirmate">
									{result.metricId}
								</a>
							</td>
							<td class="whitespace-nowrap px-4 py-3">
								<span class="text-sm text-gray-500">{result.resourceId}</span>
							</td>
							<td class="whitespace-nowrap px-4 py-3">
								<span class="text-sm text-gray-500">{result.toolId}</span>
							</td>
							<td class="max-w-xs px-4 py-3">
								<span class="block truncate text-sm text-gray-600" title={result.complianceComment}>
									{result.complianceComment || '—'}
								</span>
							</td>
							<td class="whitespace-nowrap px-4 py-3">
								<span class="text-sm text-gray-500">
									{new Date(result.createdAt).toLocaleDateString()}
								</span>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</div>
</div>