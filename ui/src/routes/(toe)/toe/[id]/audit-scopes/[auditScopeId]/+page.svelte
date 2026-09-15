<script lang="ts">
	import { browser } from '$app/environment';
	import { page } from '$app/state';
	import { invalidate } from '$app/navigation';
	import { goto } from '$app/navigation';
	import SectionHeader from '$lib/components/ui/SectionHeader.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import EvaluationSummary from '$lib/components/ui/EvaluationSummary.svelte';
	import CategorySection from '$lib/components/toe/CategorySection.svelte';
	import ImplementationCategory from '$lib/components/toe/ImplementationCategory.svelte';
	import AuditTrail from '$lib/components/ui/AuditTrail.svelte';
	import { evaluationClient } from '$lib/api/client';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	let tab: 'implementation' | 'compliance' | 'auditTrail' = $state(data.initialTab);
	let onlyMyControls = $state(false);
	let hideOutOfScope = $state(true);
	let filterAssignee = $state<string>('');
	let filterStatus = $state<string>('');

	function switchTab(t: 'implementation' | 'compliance' | 'auditTrail') {
		tab = t;
		const url = new URL(page.url);
		if (t === 'implementation') {
			url.searchParams.delete('tab');
		} else {
			url.searchParams.set('tab', t);
		}
		goto(url, { keepFocus: true, noScroll: true, invalidateAll: false });
	}

	const currentUserId = $derived((page.data as { currentUser?: { id?: string } }).currentUser?.id ?? null);

	// Check if the current user has ADMIN permission on this audit scope.
	// Removing a control from scope (delete) requires ADMIN.
	const canManageScope = $derived(
		data.userPermissions.some(
			(p) => p.userId === currentUserId && p.permission === 'PERMISSION_ADMIN'
		)
	);

	let evaluationRunning = $state(false);
	let evaluationBusy = $state(false);
	let evaluationError = $state<string | null>(null);

	$effect(() => {
		if (!browser) return;
		evaluationClient()
			.GET('/v1/evaluation/evaluate', {
				params: { query: { 'filter.auditScopeId': data.auditScope.id } }
			})
			.then((res) => {
				if (res.response.ok) {
					evaluationRunning = (res.data?.evaluationJobs ?? []).length > 0;
				}
			})
			.catch(() => {});

		// Auto-refresh evaluation results every 30 seconds.
		const interval = setInterval(() => {
			invalidate('evaluation:results');
		}, 30_000);

		return () => clearInterval(interval);
	});

	async function startEvaluation() {
		evaluationBusy = true;
		evaluationError = null;
		try {
			const { error } = await evaluationClient().POST('/v1/evaluation/evaluate/{auditScopeId}/start', {
				params: { path: { auditScopeId: data.auditScope.id } }
			});
			if (error) throw error;
			evaluationRunning = true;
		} catch {
			evaluationError = 'Failed to start evaluation';
		} finally {
			evaluationBusy = false;
		}
	}

	async function stopEvaluation() {
		evaluationBusy = true;
		evaluationError = null;
		try {
			const { error } = await evaluationClient().POST('/v1/evaluation/evaluate/{auditScopeId}/stop', {
				params: { path: { auditScopeId: data.auditScope.id } }
			});
			if (error) throw error;
			evaluationRunning = false;
		} catch {
			evaluationError = 'Failed to stop evaluation';
		} finally {
			evaluationBusy = false;
		}
	}

</script>

<div>
	<a
		href="/toe/{data.auditScope.targetOfEvaluationId}/audit-scopes/"
		class="text-sm text-gray-500 hover:text-gray-700"
	>
		← Back to Audit Scopes
	</a>

	<div class="mt-4">
		<SectionHeader title={data.auditScope.name}>
			{#snippet actions()}
				{#if data.catalog}
					<span class="inline-flex items-center rounded-full bg-blue-50 px-2.5 py-1 text-xs font-medium text-blue-700">
						{data.catalog.name}
					</span>
				{/if}
				{#if data.auditScope.assuranceLevel}
					<span class="inline-flex items-center rounded-full bg-gray-100 px-2.5 py-1 text-xs font-medium text-gray-600">
						{data.auditScope.assuranceLevel}
					</span>
				{/if}
				{#if evaluationRunning}
					<span class="inline-flex items-center gap-1.5 rounded-full bg-green-50 px-2.5 py-1 text-xs font-medium text-green-700">
						<span class="h-1.5 w-1.5 animate-pulse rounded-full bg-green-500"></span>
						Automatic evaluation enabled
					</span>
					<Button variant="danger" size="sm" onclick={stopEvaluation} disabled={evaluationBusy}>
						Disable Automatic Evaluation
					</Button>
				{:else}
					<Button variant="secondary" size="sm" onclick={startEvaluation} disabled={evaluationBusy}>
						Enable Automatic Evaluation
					</Button>
				{/if}
			{/snippet}
		</SectionHeader>
	</div>

	{#if evaluationError}
		<p class="mt-2 text-sm text-red-600">{evaluationError}</p>
	{/if}

	<!-- Tab switcher -->
	<div class="mt-6 border-b border-gray-200">
		<nav class="-mb-px flex gap-6">
			<button
				onclick={() => switchTab('implementation')}
				class="border-b-2 pb-3 text-sm font-medium transition-colors {tab === 'implementation'
					? 'border-confirmate text-confirmate'
					: 'border-transparent text-gray-500 hover:text-gray-700'}"
			>
				Implementation
			</button>
			<button
				onclick={() => switchTab('compliance')}
				class="border-b-2 pb-3 text-sm font-medium transition-colors {tab === 'compliance'
					? 'border-confirmate text-confirmate'
					: 'border-transparent text-gray-500 hover:text-gray-700'}"
			>
				Compliance
			</button>
			<button
				onclick={() => switchTab('auditTrail')}
				class="border-b-2 pb-3 text-sm font-medium transition-colors {tab === 'auditTrail'
					? 'border-confirmate text-confirmate'
					: 'border-transparent text-gray-500 hover:text-gray-700'}"
			>
				Audit Trail
			</button>
		</nav>
	</div>

	<!-- Implementation tab -->
	{#if tab === 'implementation'}
		<!-- Global filter bar -->
		<div class="mt-4 flex items-center gap-4 text-sm">
			<label class="flex cursor-pointer items-center gap-2 text-gray-600">
				<input type="checkbox" bind:checked={onlyMyControls} class="rounded" />
				My controls only
			</label>
			<label class="flex cursor-pointer items-center gap-2 text-gray-600">
				<input type="checkbox" bind:checked={hideOutOfScope} class="rounded" />
				Hide out-of-scope
			</label>
			<div class="ml-auto flex items-center gap-3">
				<select
					bind:value={filterAssignee}
					class="rounded-md border border-gray-300 px-2 py-1 text-xs text-gray-700 focus:border-blue-500 focus:outline-none"
				>
					<option value="">All assignees</option>
					<option value="unassigned">Unassigned</option>
					{#each data.users as u}
						<option value={u.id}>{[u.firstName, u.lastName].filter(Boolean).join(' ') || u.username || u.id}</option>
					{/each}
				</select>
				<select
					bind:value={filterStatus}
					class="rounded-md border border-gray-300 px-2 py-1 text-xs text-gray-700 focus:border-blue-500 focus:outline-none"
				>
					<option value="">All statuses</option>
					<option value="CONTROL_IN_SCOPE_STATE_OPEN">Open</option>
					<option value="CONTROL_IN_SCOPE_STATE_IN_PROGRESS">In Progress</option>
					<option value="CONTROL_IN_SCOPE_STATE_IMPLEMENTED">Implemented</option>
					<option value="CONTROL_IN_SCOPE_STATE_READY_FOR_REVIEW">Ready for Review</option>
					<option value="CONTROL_IN_SCOPE_STATE_ACCEPTED">Accepted</option>
				</select>
			</div>
		</div>

		<!-- Column header row -->
		<div class="mt-3 flex items-center gap-3 px-2 py-1.5 text-xs font-medium uppercase tracking-wide text-gray-400">
			<span class="w-1 shrink-0"></span>
			<span class="min-w-0 flex-1">Control</span>
			<span class="w-32 shrink-0">Assignee</span>
			<span class="w-36 shrink-0">Status</span>
			<span class="w-14 shrink-0 text-right">Updated</span>
			<span class="shrink-0 invisible">Set out of scope</span>
		</div>

		<div class="mt-4 space-y-3">
			{#if data.catalog?.categories?.length}
				{#each data.catalog.categories as category}
					<ImplementationCategory
						{category}
						controls={data.controlsByCategory[category.name] ?? []}
						controlInScopeByControlId={data.controlInScopeByControlId}
						users={data.users}
						auditScopeId={data.auditScope.id}
						targetId={data.auditScope.targetOfEvaluationId}
						filterAssigneeId={onlyMyControls ? currentUserId : (filterAssignee || null)}
						filterStatus={filterStatus || null}
						{hideOutOfScope}
						{canManageScope}
					/>
				{/each}
			{:else}
				<EmptyState title="No controls found" description="This catalog has no categories or controls defined." />
			{/if}
		</div>
	{/if}

	<!-- Compliance tab -->
	{#if tab === 'compliance'}
		<div class="mt-6">
			<EvaluationSummary results={data.evaluationResults} />

			<div class="mt-6 space-y-3">
				{#if data.catalog?.categories?.length}
					{#each data.catalog.categories as category}
						<CategorySection
							{category}
							controls={data.controlsByCategory[category.name] ?? []}
							evaluationByControl={data.evaluationByControl}
							allEvaluations={data.allEvaluationResults}
							assessmentCountByMetric={data.assessmentCountByMetric}
							assessmentById={data.assessmentById}
							controlInScopeByControlId={data.controlInScopeByControlId}
							auditScopeId={data.auditScope.id}
							targetId={data.auditScope.targetOfEvaluationId}
							users={data.users}
						/>
					{/each}
				{:else}
					<EmptyState title="No controls found" description="This catalog has no categories or controls defined." />
				{/if}
			</div>
		</div>
	{/if}

	<!-- Audit Trail tab -->
	{#if tab === 'auditTrail'}
		<div class="mt-6">
			<AuditTrail events={data.auditTrailEvents} users={data.users} controlById={data.controlById} auditScopeId={data.auditScope.id} targetId={data.auditScope.targetOfEvaluationId} />
		</div>
	{/if}
</div>
