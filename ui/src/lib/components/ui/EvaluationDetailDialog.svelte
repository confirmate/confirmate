<script lang="ts">
	import { XMark } from '@steeze-ui/heroicons';
	import { Icon } from '@steeze-ui/svelte-icon';
	import ManualEvaluationDialog from './ManualEvaluationDialog.svelte';
	import type { SchemaEvaluationResult } from '$lib/api/openapi/orchestrator';

	let {
		open = $bindable(false),
		control,
		evaluation,
		allEvaluations = [],
		assessmentById = {},
		auditScopeId,
		targetId
	}: {
		open: boolean;
		control: { id: string; name: string; shortName: string; description?: string; controlCatalogId?: string; parentControlId?: string };
		evaluation?: { status: string; timestamp?: string; comment?: string; id?: string; assessmentResultIds?: string[] };
		allEvaluations?: SchemaEvaluationResult[];
		assessmentById?: Record<string, { metricId?: string; compliant?: boolean; createdAt?: string }>;
		auditScopeId: string;
		targetId: string;
	} = $props();

	let showManualDialog = $state(false);

	function handleManualSave() {
		showManualDialog = false;
	}

	function getStatusLabel(status: string): string {
		switch (status) {
			case 'EVALUATION_STATUS_COMPLIANT': return 'Compliant';
			case 'EVALUATION_STATUS_NOT_COMPLIANT': return 'Not Compliant';
			case 'EVALUATION_STATUS_COMPLIANT_MANUALLY': return 'Compliant (manual)';
			case 'EVALUATION_STATUS_NOT_COMPLIANT_MANUALLY': return 'Not Compliant (manual)';
			case 'EVALUATION_STATUS_PENDING': return 'Pending';
			default: return 'Unknown';
		}
	}

	function getStatusColor(status: string): string {
		switch (status) {
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
	}

	function formatDate(timestamp?: string): string {
		if (!timestamp) return 'N/A';
		return new Date(timestamp).toLocaleString();
	}

	function handleClose() {
		open = false;
		showManualDialog = false;
	}
</script>

{#if open}
	<div class="fixed inset-0 z-50 flex items-center justify-center">
		<div class="absolute inset-0 bg-black/50" onclick={handleClose}></div>
		<div class="relative z-10 w-full max-w-2xl max-h-[80vh] overflow-y-auto rounded-lg bg-white p-6 shadow-xl">
			<div class="mb-4 flex items-center justify-between">
				<h3 class="text-lg font-semibold">Evaluation Details: {control.shortName ?? control.id}</h3>
				<button type="button" class="text-gray-400 hover:text-gray-600" onclick={handleClose}>
					<Icon src={XMark} class="h-5 w-5" />
				</button>
			</div>

			<div class="mb-6">
				<h4 class="font-medium text-gray-900">{control.name}</h4>
				{#if control.description}
					<p class="mt-1 text-sm text-gray-500">{control.description}</p>
				{/if}
			</div>

			{#if evaluation}
				<div class="mb-6 rounded-lg border border-gray-200 p-4">
					<h5 class="mb-3 text-sm font-medium text-gray-700">Current Status</h5>
					<div class="flex items-center gap-3">
						<span class="rounded px-2 py-1 text-sm font-medium {getStatusColor(evaluation.status)}">
							{getStatusLabel(evaluation.status)}
						</span>
						<span class="text-sm text-gray-500">{formatDate(evaluation.timestamp)}</span>
					</div>
					{#if evaluation.comment}
						<div class="mt-3 rounded bg-gray-50 p-3">
							<p class="text-xs font-medium text-gray-500">Comment</p>
							<p class="mt-1 text-sm text-gray-700">{evaluation.comment}</p>
						</div>
					{/if}
					{#if evaluation.assessmentResultIds && evaluation.assessmentResultIds.length > 0}
						<div class="mt-3">
							<p class="text-xs font-medium text-gray-500">Assessment Results ({evaluation.assessmentResultIds.length})</p>
							<div class="mt-1 space-y-1">
								{#each evaluation.assessmentResultIds as arId}
									{@const ar = assessmentById[arId]}
									<div class="flex items-center gap-2 text-xs">
										{#if ar?.compliant}
											<span class="rounded bg-emerald-50 px-1.5 py-0.5 text-emerald-600">pass</span>
										{:else}
											<span class="rounded bg-red-50 px-1.5 py-0.5 text-red-600">fail</span>
										{/if}
										<span class="font-mono text-gray-600">{arId.slice(0, 8)}...</span>
										<span class="text-gray-500">{ar?.metricId ?? 'Unknown metric'}</span>
									</div>
								{/each}
							</div>
						</div>
					{/if}
				</div>

				{#if allEvaluations && allEvaluations.length > 0}
					{@const history = allEvaluations.filter(e => e.controlId === control.id).sort((a, b) => new Date(b.timestamp ?? '').getTime() - new Date(a.timestamp ?? '').getTime())}
					{#if history.length > 1}
						<div class="mb-6 rounded-lg border border-gray-200 p-4">
							<h5 class="mb-3 text-sm font-medium text-gray-700">History ({history.length - 1} previous)</h5>
							<div class="max-h-40 space-y-2 overflow-y-auto">
								{#each history.slice(1) as hist}
									<div class="flex items-center gap-2 text-xs">
										<span class="rounded px-1.5 py-0.5 {getStatusColor(hist.status ?? '')}">{getStatusLabel(hist.status ?? '')}</span>
										<span class="text-gray-500">{formatDate(hist.timestamp)}</span>
										{#if hist.comment}
											<span class="text-gray-400">- {hist.comment}</span>
										{/if}
									</div>
								{/each}
							</div>
						</div>
					{/if}
				{/if}
			{:else}
				<div class="mb-6 rounded-lg border border-gray-200 p-4">
					<p class="text-sm text-gray-500">No evaluation result yet.</p>
				</div>
			{/if}

			<div class="flex justify-end gap-3">
				<button
					type="button"
					class="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
					onclick={() => showManualDialog = true}
				>
					Add Manual Evaluation
				</button>
			</div>
		</div>
	</div>
{/if}

<ManualEvaluationDialog 
	bind:open={showManualDialog} 
	{control}
	{targetId}
	{auditScopeId}
	onSave={handleManualSave} 
/>