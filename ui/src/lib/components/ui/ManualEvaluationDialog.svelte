<script lang="ts">
	import { XMark } from '@steeze-ui/heroicons';
	import { Icon } from '@steeze-ui/svelte-icon';
	import { invalidateAll } from '$app/navigation';
	import { orchestratorClient } from '$lib/api/client';

	let {
		open = $bindable(false),
		control,
		targetId,
		auditScopeId,
		onSave
	}: {
		open: boolean;
		control: { id: string; name: string; shortName: string; controlCatalogId?: string; parentControlId?: string };
		targetId: string;
		auditScopeId: string;
		onSave?: (status: 'EVALUATION_STATUS_COMPLIANT_MANUALLY' | 'EVALUATION_STATUS_NOT_COMPLIANT_MANUALLY', comment: string) => void;
	} = $props();

	let status: 'EVALUATION_STATUS_COMPLIANT_MANUALLY' | 'EVALUATION_STATUS_NOT_COMPLIANT_MANUALLY' = $state('EVALUATION_STATUS_COMPLIANT_MANUALLY');
	let comment = $state('');
	let saving = $state(false);
	let errorMsg = $state<string | null>(null);

	async function handleSave() {
		if (saving) return;
		saving = true;
		errorMsg = null;
		try {
			const body = {
				id: crypto.randomUUID(),
				targetOfEvaluationId: targetId,
				auditScopeId: auditScopeId,
				controlId: control.id,
				controlCatalogId: control.controlCatalogId ?? '',
				parentControlId: control.parentControlId ?? undefined,
				status: status,
				assessmentResultIds: [],
				comment: comment || undefined,
				timestamp: new Date().toISOString()
			};
			const { error } = await orchestratorClient().POST('/v1/orchestrator/evaluation_results', { body });
			if (error) {
				errorMsg = (error as { message?: string }).message ?? 'Failed to save evaluation';
				return;
			}
			onSave?.(status, comment);
			await invalidateAll();
			open = false;
			status = 'EVALUATION_STATUS_COMPLIANT_MANUALLY';
			comment = '';
		} catch {
			errorMsg = 'Failed to save evaluation';
		} finally {
			saving = false;
		}
	}

	function handleCancel() {
		open = false;
		status = 'EVALUATION_STATUS_COMPLIANT_MANUALLY';
		comment = '';
	}
</script>

{#if open}
	<div class="fixed inset-0 z-50 flex items-center justify-center">
		<div class="absolute inset-0 bg-black/50" onclick={handleCancel}></div>
		<div class="relative z-10 w-full max-w-lg rounded-lg bg-white p-6 shadow-xl">
			<div class="mb-4 flex items-center justify-between">
				<h3 class="text-lg font-semibold">Manual Evaluation: {control.shortName ?? control.id}</h3>
				<button type="button" class="text-gray-400 hover:text-gray-600" onclick={handleCancel}>
					<Icon src={XMark} class="h-5 w-5" />
				</button>
			</div>

			<div class="mb-4">
				<label class="mb-2 block text-sm font-medium text-gray-700">Status</label>
				<div class="flex gap-4">
					<label class="flex items-center gap-2">
						<input type="radio" bind:group={status} value="EVALUATION_STATUS_COMPLIANT_MANUALLY" class="text-emerald-600" />
						<span class="text-sm">Compliant (manual)</span>
					</label>
					<label class="flex items-center gap-2">
						<input type="radio" bind:group={status} value="EVALUATION_STATUS_NOT_COMPLIANT_MANUALLY" class="text-red-600" />
						<span class="text-sm">Not compliant (manual)</span>
					</label>
				</div>
			</div>

			<div class="mb-6">
				<label class="mb-2 block text-sm font-medium text-gray-700">
					Comment / Reasoning
					<span class="font-normal text-gray-500">(optional, Markdown supported)</span>
				</label>
				<textarea
					bind:value={comment}
					rows="4"
					class="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
					placeholder="Enter your reasoning or justification..."
				></textarea>
			</div>

			<div class="flex justify-end gap-3">
				<button
					type="button"
					class="rounded-md px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-100"
					onclick={handleCancel}
				>
					Cancel
				</button>
				<button
					type="button"
					class="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
					disabled={saving}
					onclick={handleSave}
				>
					{saving ? 'Saving…' : 'Save'}
				</button>
			</div>
			{#if errorMsg}
				<p class="mt-2 text-sm text-red-600">{errorMsg}</p>
			{/if}
		</div>
	</div>
{/if}