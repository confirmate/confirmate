<script lang="ts">
	import type { SchemaControl, SchemaEvaluationResult } from '$lib/api/openapi/orchestrator';
	import ManualEvaluationDialog from '$lib/components/ui/ManualEvaluationDialog.svelte';
	import EvaluationDetailDialog from '$lib/components/ui/EvaluationDetailDialog.svelte';
	import { ChevronDown } from '@steeze-ui/heroicons';
	import { Icon } from '@steeze-ui/svelte-icon';

	let {
		control,
		evaluation,
		evaluationByControl = {},
		allEvaluations = [],
		assessmentCountByMetric = {},
		assessmentById = {},
		auditScopeId,
		targetId,
		depth = 0
	}: {
		control: SchemaControl;
		evaluation?: SchemaEvaluationResult;
		evaluationByControl?: Record<string, SchemaEvaluationResult>;
		allEvaluations?: SchemaEvaluationResult[];
		assessmentCountByMetric?: Record<string, number>;
		assessmentById?: Record<string, { metricId?: string; compliant?: boolean; timestamp?: string }>;
		auditScopeId: string;
		targetId: string;
		depth?: number;
	} = $props();

	const hasChildren = $derived((control.controls?.length ?? 0) > 0);
	const metrics = $derived(control.metrics ?? []);
	const hasMetrics = $derived(metrics.length > 0);

	let open = $state(false);
	let showDetailDialog = $state(false);

	const statusColor = $derived.by(() => {
		if (!evaluation) return 'bg-gray-100 text-gray-400';
		switch (evaluation.status) {
			case 'EVALUATION_STATUS_COMPLIANT':
			case 'EVALUATION_STATUS_COMPLIANT_MANUALLY':
				return 'bg-emerald-100 text-emerald-700';
			case 'EVALUATION_STATUS_NOT_COMPLIANT':
			case 'EVALUATION_STATUS_NOT_COMPLIANT_MANUALLY':
				return 'bg-red-100 text-red-700';
			case 'EVALUATION_STATUS_PENDING':
				return 'bg-amber-100 text-amber-700';
			default:
				return 'bg-gray-100 text-gray-400';
		}
	});

	const statusLabel = $derived.by(() => {
		if (!evaluation) return 'Not evaluated';
		switch (evaluation.status) {
			case 'EVALUATION_STATUS_COMPLIANT':
				return 'Compliant';
			case 'EVALUATION_STATUS_NOT_COMPLIANT':
				return 'Not compliant';
			case 'EVALUATION_STATUS_COMPLIANT_MANUALLY':
				return 'Compliant (manual)';
			case 'EVALUATION_STATUS_NOT_COMPLIANT_MANUALLY':
				return 'Not compliant (manual)';
			case 'EVALUATION_STATUS_PENDING':
				return 'Pending';
			default:
				return 'Unknown';
		}
	});

	</script>

<div class="{depth > 0 ? 'ml-5 border-l border-gray-100 pl-4' : ''}">
	<div class="flex items-start gap-3 py-2.5">
		<span class="mt-0.5 shrink-0 rounded bg-gray-100 px-1.5 py-0.5 font-mono text-xs text-gray-500">
			{control.id}
		</span>
		<div class="min-w-0 flex-1">
			<div class="text-sm font-medium text-gray-900">{control.name}</div>
			{#if control.description}
				<div class="mt-0.5 whitespace-pre-wrap text-sm text-gray-500">{control.description}</div>
			{/if}
		</div>
		<button
			type="button"
			class="mt-0.5 shrink-0 rounded px-2 py-0.5 text-xs font-medium {statusColor} hover:opacity-80"
			title="Click for details"
			onclick={() => showDetailDialog = true}
		>
			{statusLabel}
		</button>
		{#if hasMetrics}
			<button
				type="button"
				onclick={() => (open = !open)}
				class="mt-0.5 flex shrink-0 items-center gap-1 rounded-md px-1.5 py-0.5 text-xs text-gray-400 hover:bg-gray-100 hover:text-gray-600"
			>
				<Icon
					src={ChevronDown}
					class="h-3.5 w-3.5 transition-transform {open ? '' : '-rotate-90'}"
				/>
				{metrics.length} metric{metrics.length === 1 ? '' : 's'}
			</button>
		{:else if hasChildren}
			<button
				type="button"
				onclick={() => (open = !open)}
				class="mt-0.5 flex shrink-0 items-center gap-1 rounded-md px-1.5 py-0.5 text-xs text-gray-400 hover:bg-gray-100 hover:text-gray-600"
			>
				<Icon
					src={ChevronDown}
					class="h-3.5 w-3.5 transition-transform {open ? '' : '-rotate-90'}"
				/>
				{control.controls!.length}
			</button>
		{/if}
	</div>

	{#if hasChildren && open}
		<div class="mb-1">
			{#each control.controls! as sub}
				<svelte:self
					control={sub}
					evaluation={evaluationByControl[sub.id ?? '']}
					{evaluationByControl}
					{allEvaluations}
					{assessmentCountByMetric}
					{assessmentById}
					{auditScopeId}
					{targetId}
					depth={depth + 1}
				/>
			{/each}
		</div>
	{/if}

	{#if hasMetrics && open}
		<div class="ml-8 mt-1 space-y-1 border-l-2 border-gray-100 pl-3">
			{#each metrics as metric}
				{@const counts = assessmentCountByMetric[metric.name ?? ''] ?? { passing: 0, failing: 0 }}
				<div class="flex items-center gap-2 text-xs">
					<span class="shrink-0 rounded bg-purple-50 px-1.5 py-0.5 font-mono text-purple-600">
						{metric.id}
					</span>
					<span class="text-gray-600">{metric.name ?? ''}</span>
					{#if counts.passing > 0}
						<span class="rounded bg-emerald-50 px-1.5 py-0.5 text-xs text-emerald-600">
							{counts.passing} pass
						</span>
					{/if}
					{#if counts.failing > 0}
						<span class="rounded bg-red-50 px-1.5 py-0.5 text-xs text-red-600">
							{counts.failing} fail
						</span>
					{/if}
				</div>
			{/each}
		</div>
	{/if}
</div>

<EvaluationDetailDialog 
		bind:open={showDetailDialog} 
		{control} 
		evaluation={evaluation}
		{allEvaluations}
		{assessmentById}
		{auditScopeId} 
		{targetId}
	/>