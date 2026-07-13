<script lang="ts">
	import type { SchemaAssessmentResult } from '$lib/api/openapi/orchestrator';
	import SectionHeader from '$lib/components/ui/SectionHeader.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import { Icon } from '@steeze-ui/svelte-icon';
	import { CheckCircle, XCircle } from '@steeze-ui/heroicons';
	import { page } from '$app/state';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const toeId = $derived(page.params.id);

	// Codyze packs a "[label](url)" markdown link at the end of comments — pull
	// it out so we can render it as a real anchor and keep the rest as plain
	// text instead of letting the truncate class swallow the URL.
	const linkRe = /\[([^\]]+)\]\((https?:\/\/[^)]+)\)/;
	function splitComment(raw: string | undefined | null) {
		if (!raw) return { text: '', label: '', href: '' };
		const m = raw.match(linkRe);
		if (!m) return { text: raw, label: '', href: '' };
		const text = (raw.slice(0, m.index) + raw.slice(m.index! + m[0].length)).trim();
		return { text, label: m[1], href: m[2] };
	}

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

	{#if data.metricFilter}
		<div class="mt-4 inline-flex items-center gap-2 rounded-full border border-purple-200 bg-purple-50 px-3 py-1 text-xs text-purple-700">
			Filtered by metric
			<span class="font-mono font-medium">{data.metricFilter}</span>
			<a
				href="/toe/{toeId}/assessment-results"
				class="ml-1 inline-flex h-4 w-4 items-center justify-center rounded-full text-purple-500 hover:bg-purple-100 hover:text-purple-700"
				aria-label="Clear filter"
				title="Clear filter"
			>×</a>
		</div>
	{/if}

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
		{#if data.assessmentResults.length === 0}
			<EmptyState title="No assessment results yet" description="Assessment results are generated when collectors gather evidence. Run a collection or evaluation to see results here." />
		{:else}
		<div class="overflow-x-auto rounded-lg border border-gray-200">
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
							<td class="px-4 py-3">
								{#if result.complianceComment}
									{@const parts = splitComment(result.complianceComment)}
									<div class="text-sm text-gray-600 whitespace-pre-wrap">
										{parts.text}
										{#if parts.href}
											{' '}
											<a
												href={parts.href}
												target="_blank"
												rel="noreferrer"
												class="text-confirmate hover:underline"
											>{parts.label}</a>
										{/if}
									</div>
								{:else}
									—
								{/if}
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
		{/if}
	</div>
</div>