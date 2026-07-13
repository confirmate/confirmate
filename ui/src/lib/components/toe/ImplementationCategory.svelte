<script lang="ts">
	import type { SchemaCategory, SchemaControl, SchemaControlInScope, SchemaUser } from '$lib/api/openapi/orchestrator';
	import ImplementationRow from './ImplementationRow.svelte';

	let {
		category,
		controls,
		controlInScopeByControlId,
		users,
		auditScopeId,
		targetId,
		filterAssigneeId = null,
		filterStatus = null,
		hideOutOfScope = false,
		canManageScope = false
	}: {
		category: SchemaCategory;
		controls: SchemaControl[];
		controlInScopeByControlId: Record<string, SchemaControlInScope>;
		users: SchemaUser[];
		auditScopeId: string;
		targetId: string;
		filterAssigneeId?: string | null;
		filterStatus?: string | null;
		hideOutOfScope?: boolean;
		canManageScope?: boolean;
	} = $props();

	// Count implementation states (including sub-controls) in this category.
	// Only leaf controls (those with sub-controls of their own or no children) are
	// counted — parent controls are just containers whose state is always OPEN.
	const stats = $derived.by(() => {
		let total = 0;
		let inProgress = 0;
		let implemented = 0;
		let readyForReview = 0;
		let accepted = 0;

		function walk(c: SchemaControl, isTopLevel: boolean) {
			const hasSubs = (c.controls ?? []).length > 0;
			if (!isTopLevel || !hasSubs) {
				const cis = c.id ? controlInScopeByControlId[c.id] : undefined;
				if (cis) {
					total++;
					const s = cis.state ?? '';
					if (s === 'CONTROL_IN_SCOPE_STATE_IN_PROGRESS') inProgress++;
					else if (s === 'CONTROL_IN_SCOPE_STATE_IMPLEMENTED') implemented++;
					else if (s === 'CONTROL_IN_SCOPE_STATE_READY_FOR_REVIEW') readyForReview++;
					else if (s === 'CONTROL_IN_SCOPE_STATE_ACCEPTED') accepted++;
				}
			}
			for (const sub of c.controls ?? []) walk(sub, false);
		}

		for (const c of controls) walk(c, true);
		const done = implemented + accepted;
		return { total, done, inProgress, implemented, readyForReview, accepted, open: total - done - inProgress - readyForReview };
	});

	const pctInProgress = $derived(stats.total > 0 ? Math.round((stats.inProgress / stats.total) * 100) : 0);
	const pctImplemented = $derived(stats.total > 0 ? Math.round((stats.implemented / stats.total) * 100) : 0);
	const pctReadyForReview = $derived(stats.total > 0 ? Math.round((stats.readyForReview / stats.total) * 100) : 0);
	const pctAccepted = $derived(stats.total > 0 ? Math.round((stats.accepted / stats.total) * 100) : 0);

	// A control passes the filter if it or any descendant matches.
	function controlMatchesFilter(c: SchemaControl): boolean {
		const cis = c.id ? controlInScopeByControlId[c.id] : undefined;
		if (hideOutOfScope && !cis) {
			// Still show if a sub-control is in scope
			return (c.controls ?? []).some(controlMatchesFilter);
		}
		if (filterAssigneeId) {
			if (filterAssigneeId === 'unassigned') {
				if (cis?.assigneeId) {
					return (c.controls ?? []).some(controlMatchesFilter);
				}
			} else if (cis?.assigneeId !== filterAssigneeId) {
				return (c.controls ?? []).some(controlMatchesFilter);
			}
		}
		if (filterStatus && cis?.state !== filterStatus) {
			return (c.controls ?? []).some(controlMatchesFilter);
		}
		return true;
	}

	const visibleControls = $derived(controls.filter(controlMatchesFilter));
</script>

<div class="overflow-hidden rounded-xl border border-gray-200 bg-white shadow-sm">
	<div class="flex w-full items-center gap-3 px-4 py-3">
		<span class="font-semibold text-gray-900">{category.name}</span>
		<span class="rounded-full bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-500">
			{stats.total}
		</span>

		<!-- Progress bar -->
		<div class="ml-2 flex flex-1 items-center gap-2">
			<div class="flex h-1.5 w-32 overflow-hidden rounded-full bg-gray-100">
				<div class="h-full bg-blue-400 transition-all" style="width: {pctInProgress}%"></div>
				<div class="h-full bg-purple-500 transition-all" style="width: {pctImplemented}%"></div>
				<div class="h-full bg-amber-400 transition-all" style="width: {pctReadyForReview}%"></div>
				<div class="h-full bg-emerald-500 transition-all" style="width: {pctAccepted}%"></div>
			</div>
			{#if stats.inProgress > 0}
				<span class="text-xs text-blue-600">{stats.inProgress} in progress</span>
			{/if}
			{#if stats.implemented > 0}
				<span class="text-xs text-purple-600">{stats.implemented} implemented</span>
			{/if}
			{#if stats.readyForReview > 0}
				<span class="text-xs text-amber-600">{stats.readyForReview} review</span>
			{/if}
			{#if stats.accepted > 0}
				<span class="text-xs text-emerald-600">{stats.accepted} accepted</span>
			{/if}
		</div>
	</div>

	{#if visibleControls.length > 0}
		<div class="divide-y divide-gray-50 border-t border-gray-100 px-2 py-1">
			{#each visibleControls as control}
			<ImplementationRow
				{control}
				{controlInScopeByControlId}
				{users}
				{auditScopeId}
				{targetId}
				{filterAssigneeId}
				{filterStatus}
				{hideOutOfScope}
				{canManageScope}
			/>
			{/each}
		</div>
	{/if}
</div>
