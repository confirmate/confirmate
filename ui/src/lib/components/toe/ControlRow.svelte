<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { orchestratorClient } from '$lib/api/client';
	import type { SchemaControl, SchemaControlInScope, SchemaEvaluationResult, SchemaUser } from '$lib/api/openapi/orchestrator';
	import EvaluationDetailDialog from '$lib/components/ui/EvaluationDetailDialog.svelte';
	import WorkflowDialog from '$lib/components/ui/WorkflowDialog.svelte';
	import { ChevronDown } from '@steeze-ui/heroicons';
	import { Icon } from '@steeze-ui/svelte-icon';

	let {
		control,
		evaluation,
		evaluationByControl = {},
		allEvaluations = [],
		assessmentCountByMetric = {},
		assessmentById = {},
		controlInScopeByControlId = {},
		controlInScope,
		auditScopeId,
		targetId,
		users = [],
		depth = 0
	}: {
		control: SchemaControl;
		evaluation?: SchemaEvaluationResult;
		evaluationByControl?: Record<string, SchemaEvaluationResult>;
		allEvaluations?: SchemaEvaluationResult[];
		assessmentCountByMetric?: Record<string, { passing: number; failing: number }>;
		assessmentById?: Record<string, { metricId?: string; compliant?: boolean; createdAt?: string }>;
		controlInScopeByControlId?: Record<string, SchemaControlInScope>;
		controlInScope?: SchemaControlInScope;
		auditScopeId: string;
		targetId: string;
		users?: SchemaUser[];
		depth?: number;
	} = $props();

	type State = NonNullable<SchemaControlInScope['state']>;

	const STATE_LABELS: Partial<Record<State, string>> = {
		CONTROL_IN_SCOPE_STATE_OPEN: 'Open',
		CONTROL_IN_SCOPE_STATE_IN_PROGRESS: 'In Progress',
		CONTROL_IN_SCOPE_STATE_IMPLEMENTED: 'Implemented',
		CONTROL_IN_SCOPE_STATE_READY_FOR_REVIEW: 'Ready for Review',
		CONTROL_IN_SCOPE_STATE_ACCEPTED: 'Accepted'
	};

	const STATE_COLORS: Partial<Record<State, string>> = {
		CONTROL_IN_SCOPE_STATE_OPEN: 'bg-gray-100 text-gray-600',
		CONTROL_IN_SCOPE_STATE_IN_PROGRESS: 'bg-blue-100 text-blue-700',
		CONTROL_IN_SCOPE_STATE_IMPLEMENTED: 'bg-purple-100 text-purple-700',
		CONTROL_IN_SCOPE_STATE_READY_FOR_REVIEW: 'bg-amber-100 text-amber-700',
		CONTROL_IN_SCOPE_STATE_ACCEPTED: 'bg-emerald-100 text-emerald-700'
	};

	const inScope = $derived(!!controlInScope);
	let scopeBusy = $state(false);
	let showWorkflowDialog = $state(false);
	let assigneeBusy = $state(false);
	let actionError = $state<string | null>(null);

	const stateLabel = $derived(
		controlInScope?.state ? (STATE_LABELS[controlInScope.state as State] ?? 'Open') : 'Open'
	);
	const stateColor = $derived(
		controlInScope?.state ? (STATE_COLORS[controlInScope.state as State] ?? 'bg-gray-100 text-gray-600') : 'bg-gray-100 text-gray-600'
	);

	const assignee = $derived(
		users.find((u) => u.id === controlInScope?.assigneeId)
	);

	async function updateAssignee(assigneeId: string) {
		if (!controlInScope?.id || assigneeBusy) return;
		assigneeBusy = true;
		actionError = null;
		try {
			const { error } = await orchestratorClient().PUT('/v1/orchestrator/controls_in_scope/{id}', {
				params: { path: { id: controlInScope.id } },
				body: { id: controlInScope.id, assigneeId: assigneeId || undefined }
			});
			if (error) throw error;
			await invalidateAll();
		} catch {
			actionError = 'Failed to update assignee';
		} finally {
			assigneeBusy = false;
		}
	}

	async function toggleScope() {
		if (scopeBusy) return;
		scopeBusy = true;
		actionError = null;
		try {
			const client = orchestratorClient();
			if (controlInScope?.id) {
				const { error } = await client.DELETE('/v1/orchestrator/controls_in_scope/{id}', {
					params: { path: { id: controlInScope.id } }
				});
				if (error) throw error;
			} else {
				const { error } = await client.POST('/v1/orchestrator/controls_in_scope', {
					body: {
						auditScopeId,
						controlId: control.id,
						targetOfEvaluationId: targetId,
						state: 'CONTROL_IN_SCOPE_STATE_OPEN'
					}
				});
				if (error) throw error;
			}
			await invalidateAll();
		} catch {
			actionError = 'Failed to update scope';
		} finally {
			scopeBusy = false;
		}
	}

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

<div class="{depth > 0 ? 'ml-5 border-l border-gray-100 pl-4' : ''} {inScope ? '' : 'opacity-50'}">
	<div class="flex items-start gap-3 py-2.5">
		<span class="mt-0.5 shrink-0 rounded bg-gray-100 px-1.5 py-0.5 font-mono text-xs text-gray-500">
			{control.shortName ?? control.id}
		</span>
		<div class="min-w-0 flex-1">
			<div class="text-sm font-medium text-gray-900">{control.name}</div>
			{#if control.description}
				<div class="mt-0.5 whitespace-pre-wrap text-sm text-gray-500">{control.description}</div>
			{/if}
			{#if inScope && assignee}
				<div class="mt-0.5 text-xs text-gray-400">
					{[assignee.firstName, assignee.lastName].filter(Boolean).join(' ') || assignee.username || assignee.email}
				</div>
			{/if}
		</div>
		{#if inScope && controlInScope}
			<button
				type="button"
				class="mt-0.5 shrink-0 rounded px-2 py-0.5 text-xs font-medium {stateColor} hover:opacity-80"
				title="Click to change workflow state"
				onclick={() => (showWorkflowDialog = true)}
			>
				{stateLabel}
			</button>
		{/if}
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
					{controlInScopeByControlId}
					controlInScope={controlInScopeByControlId[sub.id ?? '']}
					{auditScopeId}
					{targetId}
					{users}
					depth={depth + 1}
				/>
			{/each}
		</div>
	{/if}

	{#if hasMetrics && open}
		<div class="ml-8 mt-1 space-y-1 border-l-2 border-gray-100 pl-3">
			{#each metrics as metric}
				{@const counts = assessmentCountByMetric[metric.id ?? ''] ?? { passing: 0, failing: 0 }}
				<div class="flex items-center gap-2 text-xs">
					<a
						href="/toe/{targetId}/assessment-results/?metric={encodeURIComponent(metric.id)}"
						class="shrink-0 rounded bg-purple-50 px-1.5 py-0.5 font-mono text-purple-600 hover:bg-purple-100 hover:underline"
						title="Show assessment results for this metric"
					>
						{metric.id}
					</a>
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

{#if controlInScope}
	<WorkflowDialog
		bind:open={showWorkflowDialog}
		{controlInScope}
		controlName={control.shortName ?? control.name ?? control.id ?? ''}
	/>
{/if}