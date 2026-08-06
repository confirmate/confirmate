<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { orchestratorClient } from '$lib/api/client';
	import WorkflowDialog from '$lib/components/ui/WorkflowDialog.svelte';
	import AuditTrail from '$lib/components/ui/AuditTrail.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	let actionError = $state<string | null>(null);

	// Assignee
	let assigneeBusy = $state(false);
	async function updateAssignee(assigneeId: string) {
		if (!data.cis?.id || assigneeBusy) return;
		assigneeBusy = true;
		actionError = null;
		try {
			const { error } = await orchestratorClient().PUT('/v1/orchestrator/controls_in_scope/{id}', {
				params: { path: { id: data.cis.id } },
				body: { id: data.cis.id, assigneeId: assigneeId || undefined, implementationDetails: data.cis.implementationDetails }
			});
			if (error) throw error;
			await invalidateAll();
		} catch {
			actionError = 'Failed to update assignee';
		} finally {
			assigneeBusy = false;
		}
	}

	// Implementation comment
	let commentDraft = $state(data.cis?.implementationDetails ?? '');
	let commentBusy = $state(false);
	let commentDirty = $derived(commentDraft !== (data.cis?.implementationDetails ?? ''));

	$effect(() => {
		commentDraft = data.cis?.implementationDetails ?? '';
	});

	async function saveComment() {
		if (!data.cis?.id || commentBusy) return;
		commentBusy = true;
		actionError = null;
		try {
			const { error } = await orchestratorClient().PUT('/v1/orchestrator/controls_in_scope/{id}', {
				params: { path: { id: data.cis.id } },
				body: { id: data.cis.id, implementationDetails: commentDraft, assigneeId: data.cis.assigneeId }
			});
			if (error) throw error;
			await invalidateAll();
		} catch {
			actionError = 'Failed to save implementation notes';
		} finally {
			commentBusy = false;
		}
	}

	// State
	let showWorkflowDialog = $state(false);

	const STATE_CONFIG: Record<string, { label: string; dot: string; pill: string }> = {
		CONTROL_IN_SCOPE_STATE_OPEN: { label: 'Open', dot: 'bg-gray-400', pill: 'bg-gray-100 text-gray-600' },
		CONTROL_IN_SCOPE_STATE_IN_PROGRESS: { label: 'In Progress', dot: 'bg-blue-500', pill: 'bg-blue-50 text-blue-700' },
		CONTROL_IN_SCOPE_STATE_IMPLEMENTED: { label: 'Implemented', dot: 'bg-purple-500', pill: 'bg-purple-50 text-purple-700' },
		CONTROL_IN_SCOPE_STATE_READY_FOR_REVIEW: { label: 'Ready for Review', dot: 'bg-amber-400', pill: 'bg-amber-50 text-amber-700' },
		CONTROL_IN_SCOPE_STATE_ACCEPTED: { label: 'Accepted', dot: 'bg-emerald-500', pill: 'bg-emerald-50 text-emerald-700' }
	};

	const stateCfg = $derived(
		data.cis?.state ? (STATE_CONFIG[data.cis.state] ?? STATE_CONFIG.CONTROL_IN_SCOPE_STATE_OPEN) : null
	);

	// Evaluation
	const EVAL_CONFIG: Record<string, { label: string; cls: string }> = {
		EVALUATION_STATUS_COMPLIANT: { label: 'Compliant', cls: 'bg-emerald-50 text-emerald-700' },
		EVALUATION_STATUS_COMPLIANT_MANUALLY: { label: 'Compliant (manual)', cls: 'bg-emerald-50 text-emerald-700' },
		EVALUATION_STATUS_NOT_COMPLIANT: { label: 'Not compliant', cls: 'bg-red-50 text-red-700' },
		EVALUATION_STATUS_NOT_COMPLIANT_MANUALLY: { label: 'Not compliant (manual)', cls: 'bg-red-50 text-red-700' },
		EVALUATION_STATUS_PENDING: { label: 'Pending', cls: 'bg-amber-50 text-amber-700' }
	};
	const evalCfg = $derived(
		data.evaluation?.status ? (EVAL_CONFIG[data.evaluation.status] ?? null) : null
	);

	const assignee = $derived(data.users.find((u) => u.id === data.cis?.assigneeId));
	const assigneeName = $derived(
		assignee
			? [assignee.firstName, assignee.lastName].filter(Boolean).join(' ') || assignee.username
			: null
	);
</script>

<div>
	<!-- Breadcrumb -->
	<a
		href="/toe/{data.auditScope?.targetOfEvaluationId}/audit-scopes/{data.auditScope?.id}/"
		class="text-sm text-gray-500 hover:text-gray-700"
	>
		← {data.auditScope?.name ?? 'Audit Scope'}
	</a>

	<div class="mt-4">
		<div class="flex items-baseline gap-3">
			<span class="rounded bg-gray-100 px-2 py-0.5 font-mono text-sm text-gray-500">
				{data.control.shortName}
			</span>
			<h1 class="text-xl font-semibold text-gray-900">{data.control.name}</h1>
		</div>
		{#if data.control.description}
			<p class="mt-2 text-sm text-gray-600">{data.control.description}</p>
		{/if}
		{#if actionError}
			<p class="mt-2 text-sm text-red-600">{actionError}</p>
		{/if}
	</div>

	<div class="mt-6 space-y-5">

			<!-- Status card -->
			<div class="rounded-xl border border-gray-200 bg-white p-5">
				<h2 class="mb-4 text-sm font-semibold text-gray-700">Implementation</h2>
				<div class="grid grid-cols-2 gap-4">
					<!-- Assignee -->
					<div>
						<div class="mb-1 text-xs font-medium text-gray-500">Assignee</div>
						{#if data.cis}
							<select
								class="w-full rounded-lg border border-gray-200 px-3 py-1.5 text-sm text-gray-700 focus:border-blue-400 focus:outline-none disabled:opacity-50"
								value={data.cis.assigneeId ?? ''}
								disabled={assigneeBusy}
								onchange={(e) => updateAssignee((e.target as HTMLSelectElement).value)}
							>
								<option value="">— Unassigned —</option>
								{#each data.users as u}
									<option value={u.id}>
										{[u.firstName, u.lastName].filter(Boolean).join(' ') || u.username || u.id}
									</option>
								{/each}
							</select>
						{:else}
							<span class="text-sm text-gray-400">Not in scope</span>
						{/if}
					</div>

					<!-- State -->
					<div>
						<div class="mb-1 text-xs font-medium text-gray-500">Status</div>
						{#if data.cis && stateCfg}
							<button
								type="button"
								onclick={() => (showWorkflowDialog = true)}
								class="inline-flex items-center gap-2 rounded-lg border border-gray-200 px-3 py-1.5 text-sm hover:bg-gray-50"
							>
								<span class="h-2 w-2 rounded-full {stateCfg.dot}"></span>
								<span class="{stateCfg.pill.split(' ')[1]}">{stateCfg.label}</span>
								<span class="text-xs text-gray-400">click to change</span>
							</button>
						{:else}
							<span class="text-sm text-gray-400">Not in scope</span>
						{/if}
					</div>

					<!-- Assessment -->
					<div>
						<div class="mb-1 text-xs font-medium text-gray-500">Assessment</div>
						{#if evalCfg}
							<span class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium {evalCfg.cls}">
								{evalCfg.label}
							</span>
						{:else}
							<span class="text-sm text-gray-400">Not evaluated</span>
						{/if}
					</div>

					<!-- Last updated -->
					<div>
						<div class="mb-1 text-xs font-medium text-gray-500">Last updated</div>
						<span class="text-sm text-gray-700">
							{data.cis?.updatedAt ? new Date(data.cis.updatedAt).toLocaleString() : '—'}
						</span>
					</div>
				</div>
			</div>

			<!-- Implementation notes -->
			<div class="rounded-xl border border-gray-200 bg-white p-5">
				<h2 class="mb-3 text-sm font-semibold text-gray-700">Implementation Notes</h2>
				{#if data.cis}
					<textarea
						bind:value={commentDraft}
						rows="5"
						placeholder="Describe how this control is being addressed…"
						class="w-full rounded-lg border border-gray-200 px-3 py-2 text-sm text-gray-700 focus:border-blue-400 focus:outline-none focus:ring-1 focus:ring-blue-200"
					></textarea>
					<div class="mt-2 flex justify-end">
						<button
							type="button"
							disabled={!commentDirty || commentBusy}
							onclick={saveComment}
							class="rounded-lg bg-blue-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-40"
						>
							{commentBusy ? 'Saving…' : 'Save notes'}
						</button>
					</div>
				{:else}
					<p class="text-sm text-gray-400">This control is not in scope for this audit.</p>
				{/if}
			</div>

			<!-- Metrics -->
			{#if data.control.metrics?.length}
				<div class="rounded-xl border border-gray-200 bg-white p-5">
					<h2 class="mb-3 text-sm font-semibold text-gray-700">Metrics</h2>
					<div class="space-y-1">
						{#each data.control.metrics as metric}
							<div class="flex items-center gap-2 rounded-lg px-2 py-1.5 hover:bg-gray-50">
								<a
									href="/toe/{data.auditScope?.targetOfEvaluationId}/assessment-results/?metric={encodeURIComponent(metric.id ?? '')}"
									class="font-mono text-xs text-blue-600 hover:underline"
								>
									{metric.id}
								</a>
								<span class="text-sm text-gray-700">{metric.name}</span>
							</div>
						{/each}
					</div>
				</div>
			{/if}

		<AuditTrail events={data.auditTrailEvents} users={data.users} />
	</div>
</div>

{#if data.cis}
	<WorkflowDialog
		bind:open={showWorkflowDialog}
		controlInScope={data.cis}
		controlName={data.control.shortName ?? data.control.name ?? ''}
	/>
{/if}
