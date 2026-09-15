<script lang="ts">
	import { ChevronRight, Trash } from '@steeze-ui/heroicons';
	import { Icon } from '@steeze-ui/svelte-icon';
	import { orchestratorClient } from '$lib/api/client';
	import { invalidate } from '$app/navigation';
	import type { SchemaAuditScope, SchemaCatalog } from '$lib/api/openapi/orchestrator';

	let {
		auditScope,
		catalog,
		ondelete
	}: {
		auditScope: SchemaAuditScope;
		catalog: SchemaCatalog | undefined;
		ondelete?: (id: string) => void;
	} = $props();

	let deleting = $state(false);
	let showConfirm = $state(false);
	let error = $state<string | null>(null);

	async function handleDelete() {
		if (deleting) return;
		deleting = true;
		error = null;
		try {
			const { error: apiError } = await orchestratorClient().DELETE(
				'/v1/orchestrator/audit_scopes/{auditScopeId}',
				{ params: { path: { auditScopeId: auditScope.id } } }
			);
			if (apiError) {
				error = (apiError as { message?: string }).message ?? 'Failed to delete audit scope';
				return;
			}
			ondelete?.(auditScope.id);
			invalidate((url) => url.pathname.includes('/audit_scopes'));
		} catch {
			error = 'Failed to delete audit scope';
		} finally {
			deleting = false;
			showConfirm = false;
		}
	}
</script>

<div class="flex items-center justify-between rounded-lg border border-gray-200 bg-white p-4 hover:border-confirmate hover:bg-blue-50">
	<a
		href="/toe/{auditScope.targetOfEvaluationId}/audit-scopes/{auditScope.id}/"
		class="flex-1"
	>
		<div class="font-medium text-gray-900">{auditScope.name}</div>
		{#if catalog}
			<div class="mt-0.5 text-sm text-gray-500">{catalog.name}</div>
		{/if}
		{#if auditScope.assuranceLevel}
			<div class="mt-1">
				<span class="inline-flex items-center rounded-full bg-blue-50 px-2 py-0.5 text-xs font-medium text-blue-700">
					{auditScope.assuranceLevel}
				</span>
			</div>
		{/if}
	</a>
	<div class="flex items-center gap-2">
		{#if error}
			<span class="text-xs text-red-500">{error}</span>
		{/if}
		{#if showConfirm}
			<button
				type="button"
				onclick={handleDelete}
				disabled={deleting}
				class="rounded-md bg-red-600 px-2 py-1 text-xs font-medium text-white hover:bg-red-700 disabled:opacity-50"
			>
				{deleting ? 'Deleting…' : 'Confirm'}
			</button>
			<button
				type="button"
				onclick={() => { showConfirm = false; error = null; }}
				class="rounded-md px-2 py-1 text-xs text-gray-500 hover:bg-gray-100"
			>
				Cancel
			</button>
		{:else if ondelete}
			<button
				type="button"
				onclick={() => (showConfirm = true)}
				title="Delete audit scope"
				class="rounded p-1 text-gray-300 hover:bg-red-50 hover:text-red-500"
			>
				<Icon src={Trash} class="h-4 w-4" />
			</button>
		{/if}
		<a href="/toe/{auditScope.targetOfEvaluationId}/audit-scopes/{auditScope.id}/">
			<Icon src={ChevronRight} class="h-5 w-5 shrink-0 text-gray-400" />
		</a>
	</div>
</div>
