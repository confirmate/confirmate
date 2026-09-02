<script lang="ts">
	import type { SchemaControl, SchemaControlInScope, SchemaUser } from '$lib/api/openapi/orchestrator';
	import WorkflowDialog from '$lib/components/ui/WorkflowDialog.svelte';
	import ScopeCommentDialog from '$lib/components/ui/ScopeCommentDialog.svelte';
	import { orchestratorClient } from '$lib/api/client';
	import { invalidateAll } from '$app/navigation';

	let {
		control,
		controlInScopeByControlId,
		users,
		auditScopeId,
		targetId,
		depth = 0,
		filterAssigneeId = null,
		filterStatus = null,
		hideOutOfScope = false,
		parentInScope = true,
		canManageScope = false
	}: {
		control: SchemaControl;
		controlInScopeByControlId: Record<string, SchemaControlInScope>;
		users: SchemaUser[];
		auditScopeId: string;
		targetId: string;
		depth?: number;
		filterAssigneeId?: string | null;
		filterStatus?: string | null;
		hideOutOfScope?: boolean;
		parentInScope?: boolean;
		canManageScope?: boolean;
	} = $props();

	const cis = $derived(control.id ? controlInScopeByControlId[control.id] : undefined);
	const assignee = $derived(users.find((u) => u.id === cis?.assigneeId));

	type State = NonNullable<SchemaControlInScope['state']>;

	const STATE_CONFIG: Partial<Record<State, { label: string; dot: string; pill: string }>> = {
		CONTROL_IN_SCOPE_STATE_OPEN: { label: 'Open', dot: 'bg-gray-400', pill: 'bg-gray-100 text-gray-600 hover:bg-gray-200' },
		CONTROL_IN_SCOPE_STATE_IN_PROGRESS: { label: 'In Progress', dot: 'bg-blue-500', pill: 'bg-blue-50 text-blue-700 hover:bg-blue-100' },
		CONTROL_IN_SCOPE_STATE_IMPLEMENTED: { label: 'Implemented', dot: 'bg-purple-500', pill: 'bg-purple-50 text-purple-700 hover:bg-purple-100' },
		CONTROL_IN_SCOPE_STATE_READY_FOR_REVIEW: { label: 'Ready for Review', dot: 'bg-amber-400', pill: 'bg-amber-50 text-amber-700 hover:bg-amber-100' },
		CONTROL_IN_SCOPE_STATE_ACCEPTED: { label: 'Accepted', dot: 'bg-emerald-500', pill: 'bg-emerald-50 text-emerald-700 hover:bg-emerald-100' }
	};

	const stateCfg = $derived(
		cis?.state
			? (STATE_CONFIG[cis.state as State] ?? STATE_CONFIG.CONTROL_IN_SCOPE_STATE_OPEN)
			: STATE_CONFIG.CONTROL_IN_SCOPE_STATE_OPEN
	);

	let showWorkflowDialog = $state(false);
	let showScopeDialog = $state(false);
	let scopeDialogAction: 'add' | 'remove' = $state('remove');
	let assigneeBusy = $state(false);
	let scopeBusy = $state(false);
	let actionError = $state<string | null>(null);

	async function updateAssignee(assigneeId: string) {
		if (!cis?.id || assigneeBusy) return;
		assigneeBusy = true;
		actionError = null;
		try {
			const { error } = await orchestratorClient().PUT('/v1/orchestrator/controls_in_scope/{id}', {
				params: { path: { id: cis.id } },
				body: { id: cis.id, assigneeId: assigneeId || undefined }
			});
			if (error) throw error;
			await invalidateAll();
		} catch {
			actionError = 'Failed to update assignee';
		} finally {
			assigneeBusy = false;
		}
	}

	function openScopeDialog(action: 'add' | 'remove') {
		scopeDialogAction = action;
		showScopeDialog = true;
	}

	async function doToggleScope(comment: string) {
		if (scopeBusy) return;
		scopeBusy = true;
		actionError = null;
		try {
			if (cis?.id) {
				const { error } = await orchestratorClient().DELETE('/v1/orchestrator/controls_in_scope/{id}', {
					params: { path: { id: cis.id }, query: { comment } }
				});
				if (error) throw error;
			} else {
				const { error } = await orchestratorClient().POST('/v1/orchestrator/controls_in_scope', {
					body: {
						auditScopeId,
						controlId: control.id,
						targetOfEvaluationId: targetId,
						comment
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

	// Apply filters to sub-controls, mirroring ImplementationCategory logic
	const subControls = $derived((control.controls ?? []).filter((c) => {
		const subCis = c.id ? controlInScopeByControlId[c.id] : undefined;
		if (hideOutOfScope && !subCis) return (c.controls ?? []).length > 0;
		if (filterAssigneeId) {
			if (filterAssigneeId === 'unassigned') {
				if (subCis?.assigneeId) return false;
			} else if (subCis?.assigneeId !== filterAssigneeId) return false;
		}
		if (filterStatus && subCis?.state !== filterStatus) return false;
		return true;
	}));

	const displayName = $derived(control.name || control.shortName || control.id || '');

	const updatedDate = $derived(
		cis?.updatedAt
			? new Date(cis.updatedAt).toLocaleDateString('en-GB', { day: '2-digit', month: 'short' })
			: null
	);

	const assigneeLabel = $derived(
		assignee
			? [assignee.firstName, assignee.lastName].filter(Boolean).join(' ') ||
				  assignee.username ||
				  assignee.id
			: null
	);
</script>

<div class="{depth > 0 ? 'ml-4 border-l border-gray-100 pl-2' : ''}">
	<div class="group flex items-center gap-3 rounded-lg px-2 py-2 hover:bg-gray-50 {cis ? '' : 'opacity-40 hover:opacity-70'}">
		<!-- ID badge -->
		<span
			class="shrink-0 whitespace-nowrap rounded bg-gray-100 px-1.5 py-0.5 font-mono text-xs text-gray-500"
		>
			{control.shortName ?? '—'}
		</span>

		<!-- Name -->
		<a
			href="/toe/{targetId}/audit-scopes/{auditScopeId}/controls/{control.id}/"
			class="min-w-0 flex-1 truncate text-sm hover:underline {depth === 0 ? 'font-medium text-gray-900' : 'text-gray-700'}"
		>
			{displayName}
		</a>

	{#if cis}
		<!-- Assignee selector -->
		<div class="w-32 shrink-0">
			<select
				class="cursor-pointer rounded-full border border-transparent bg-transparent py-0.5 pl-2 pr-6 text-xs font-medium hover:border-gray-200 hover:bg-white focus:border-gray-300 focus:outline-none disabled:opacity-50 {assignee ? 'text-gray-800' : 'text-gray-400'}"
				value={cis.assigneeId ?? ''}
				disabled={assigneeBusy}
				onchange={(e) => updateAssignee((e.target as HTMLSelectElement).value)}
			>
				<option value="">Unassigned</option>
				{#each users as u}
					<option value={u.id}>
						{[u.firstName, u.lastName].filter(Boolean).join(' ') || u.username || u.id}
					</option>
				{/each}
			</select>
		</div>

		<!-- State chip -->
		<button
			type="button"
			class="inline-flex w-36 shrink-0 items-center justify-center gap-1.5 rounded-full px-2.5 py-0.5 text-xs font-medium transition-colors {stateCfg?.pill}"
			onclick={() => (showWorkflowDialog = true)}
		>
			<span class="h-1.5 w-1.5 rounded-full {stateCfg?.dot}"></span>
			{stateCfg?.label}
		</button>

		<!-- Updated -->
		<span class="w-14 shrink-0 text-right text-xs text-gray-400">
			{updatedDate ?? '—'}
		</span>

		<!-- Remove from scope (visible on hover) -->
		<button
			type="button"
			class="shrink-0 rounded px-1.5 py-0.5 text-xs text-gray-300 transition-opacity hover:bg-red-50 hover:text-red-500 group-hover:opacity-100 disabled:opacity-30 max-sm:opacity-70 sm:opacity-0"
			title={canManageScope ? 'Set out of scope' : 'Requires admin permission'}
			disabled={scopeBusy || !canManageScope}
			onclick={() => openScopeDialog('remove')}
		>
			Set out of scope
		</button>
	{:else if parentInScope}
		<!-- Add to scope (visible on hover) — only when parent is in scope -->
		<button
			type="button"
			class="shrink-0 rounded px-1.5 py-0.5 text-xs text-gray-300 transition-opacity hover:bg-blue-50 hover:text-blue-600 group-hover:opacity-100 disabled:opacity-30 max-sm:opacity-70 sm:opacity-0"
			title={canManageScope ? 'Bring back into scope' : 'Requires admin permission'}
			disabled={scopeBusy || !canManageScope}
			onclick={() => openScopeDialog('add')}
		>
			Add to scope
		</button>
	{/if}
	</div>

	{#if actionError}
		<p class="px-2 pb-1 text-xs text-red-500">{actionError}</p>
	{/if}

	{#each subControls as sub}
		<svelte:self
			control={sub}
			{controlInScopeByControlId}
			{users}
			{auditScopeId}
			{targetId}
			depth={depth + 1}
			{filterAssigneeId}
			{filterStatus}
			{hideOutOfScope}
			parentInScope={!!cis}
			{canManageScope}
		/>
	{/each}
</div>

{#if cis}
	<WorkflowDialog
		bind:open={showWorkflowDialog}
		controlInScope={cis}
		controlName={control.shortName ?? control.name ?? control.id ?? ''}
	/>
{/if}

<ScopeCommentDialog
	bind:open={showScopeDialog}
	controlName={control.shortName ?? control.name ?? control.id ?? ''}
	action={scopeDialogAction}
	onconfirm={doToggleScope}
/>
