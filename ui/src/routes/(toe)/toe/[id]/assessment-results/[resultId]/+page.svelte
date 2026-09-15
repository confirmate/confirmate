<script lang="ts">
	import type { SchemaAssessmentResult, SchemaComparisonResult } from '$lib/api/openapi/orchestrator';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import { Icon } from '@steeze-ui/svelte-icon';
	import {
		ArrowLeft,
		CheckCircle,
		XCircle,
		Beaker,
		DocumentText,
		Clock,
		User,
		Tag
	} from '@steeze-ui/heroicons';
	import { page } from '$app/state';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const result: SchemaAssessmentResult = $derived(data.result);
	const toeId = $derived(page.params.id);

	function formatValue(val: unknown): string {
		if (val === null || val === undefined) return 'null';
		if (typeof val === 'object') return JSON.stringify(val);
		return String(val);
	}
</script>

<div>
	<div class="mb-6">
		<a
			href="/toe/{toeId}/assessment-results/"
			class="inline-flex items-center gap-1.5 text-sm text-gray-500 hover:text-confirmate"
		>
			<Icon src={ArrowLeft} class="h-4 w-4" />
			Back to Assessment Results
		</a>
	</div>

	<PageHeader
		title={result.id}
		description="Assessment result details"
	/>

	<div class="mt-6 grid grid-cols-3 gap-4">
		<div class="rounded-lg border border-gray-200 bg-white px-5 py-4">
			<div class="text-sm font-medium text-gray-500">Status</div>
			<div class="mt-1 flex items-center gap-2">
				{#if result.compliant}
					<Icon src={CheckCircle} class="h-6 w-6 text-green-600" />
					<span class="text-xl font-bold text-green-600">Compliant</span>
				{:else}
					<Icon src={XCircle} class="h-6 w-6 text-red-600" />
					<span class="text-xl font-bold text-red-600">Non-Compliant</span>
				{/if}
			</div>
		</div>

		<div class="rounded-lg border border-gray-200 bg-white px-5 py-4">
			<div class="text-sm font-medium text-gray-500">Resource</div>
			<div class="mt-1 text-xl font-bold text-gray-900">{result.resourceId}</div>
			<div class="mt-0.5 text-xs text-gray-400">{result.resourceTypes?.join(', ')}</div>
		</div>

		<div class="rounded-lg border border-gray-200 bg-white px-5 py-4">
			<div class="text-sm font-medium text-gray-500">Tool</div>
			<div class="mt-1 text-xl font-bold text-gray-900">{result.toolId}</div>
		</div>
	</div>

	<div class="mt-6 grid grid-cols-2 gap-4">
		<div class="rounded-lg border border-gray-200 bg-white px-5 py-4">
			<div class="flex items-center gap-2 text-sm font-medium text-gray-500">
				<Icon src={Tag} class="h-4 w-4" />
				Metric
			</div>
			<div class="mt-2 font-mono text-sm text-gray-900">{result.metricId}</div>
		</div>

		<div class="rounded-lg border border-gray-200 bg-white px-5 py-4">
			<div class="flex items-center gap-2 text-sm font-medium text-gray-500">
				<Icon src={DocumentText} class="h-4 w-4" />
				Evidence
			</div>
			<div class="mt-2">
				<a
					href="/toe/{toeId}/evidences/{result.evidenceId}"
					class="text-sm font-mono text-confirmate hover:underline"
				>
					{result.evidenceId}
				</a>
			</div>
		</div>

		<div class="rounded-lg border border-gray-200 bg-white px-5 py-4">
			<div class="flex items-center gap-2 text-sm font-medium text-gray-500">
				<Icon src={Clock} class="h-4 w-4" />
				Assessment Time
			</div>
			<div class="mt-2 text-sm text-gray-900">
				{new Date(result.createdAt).toLocaleString()}
			</div>
		</div>

		<div class="rounded-lg border border-gray-200 bg-white px-5 py-4">
			<div class="flex items-center gap-2 text-sm font-medium text-gray-500">
				<Icon src={User} class="h-4 w-4" />
				Target of Evaluation
			</div>
			<div class="mt-2 text-sm text-gray-900">{result.targetOfEvaluationId}</div>
		</div>
	</div>

	<div class="mt-6 rounded-lg border border-gray-200 bg-white px-5 py-4">
		<div class="text-sm font-medium text-gray-500">Compliance Comment</div>
		<div class="mt-2 text-sm text-gray-900">{result.complianceComment || 'No comment'}</div>
	</div>

	{#if result.complianceDetails && result.complianceDetails.length > 0}
		<div class="mt-6 rounded-lg border border-gray-200 bg-white px-5 py-4">
			<div class="text-sm font-medium text-gray-500">Compliance Details</div>
			<div class="mt-3 space-y-2">
				{#each result.complianceDetails as detail}
					{@const d = detail as SchemaComparisonResult}
					<div class="flex items-start gap-2 rounded bg-gray-50 px-3 py-2">
						{#if d.success}
							<Icon src={CheckCircle} class="h-4 w-4 shrink-0 text-green-600" />
						{:else}
							<Icon src={XCircle} class="h-4 w-4 shrink-0 text-red-600" />
						{/if}
						<div class="text-sm">
							<span class="font-mono text-gray-700">{d.property}</span>
							<span class="text-gray-500"> — {d.operator}</span>
							<span class="text-gray-400">(actual: {formatValue(d.value)})</span>
							<span class="text-gray-400">(expected: {formatValue(d.targetValue)})</span>
						</div>
					</div>
				{/each}
			</div>
		</div>
	{/if}

	{#if result.history && result.history.length > 0}
		<div class="mt-6 rounded-lg border border-gray-200 bg-white">
			<div class="border-b border-gray-200 px-5 py-4">
				<div class="flex items-center gap-2 text-sm font-medium text-gray-500">
					<Icon src={Beaker} class="h-4 w-4" />
					Evidence History
				</div>
				<div class="mt-1 text-xs text-gray-400">
					{result.history.length} previous evidence record(s) with the same content
				</div>
			</div>
			<div class="overflow-x-auto">
				<table class="min-w-full divide-y divide-gray-200">
					<thead class="bg-gray-50">
						<tr>
							<th
								scope="col"
								class="px-4 py-2 text-left text-xs font-medium uppercase text-gray-500"
							>
								Evidence ID
							</th>
							<th
								scope="col"
								class="px-4 py-2 text-left text-xs font-medium uppercase text-gray-500"
							>
								Recorded At
							</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-200">
						{#each result.history as record}
							<tr class="hover:bg-gray-50">
								<td class="whitespace-nowrap px-4 py-2">
									<a
										href="/toe/{toeId}/evidences/{record.evidenceId}"
										class="font-mono text-sm text-confirmate hover:underline"
									>
										{record.evidenceId}
									</a>
								</td>
								<td class="whitespace-nowrap px-4 py-2">
									<span class="text-sm text-gray-500">
										{new Date(record.evidenceRecordedAt).toLocaleString()}
									</span>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</div>
	{/if}
</div>