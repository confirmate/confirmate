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
		control: { id: string; name: string; categoryName?: string; categoryCatalogId?: string; parentControlId?: string };
		targetId: string;
		auditScopeId: string;
		onSave?: (status: 'EVALUATION_STATUS_COMPLIANT_MANUALLY' | 'EVALUATION_STATUS_NOT_COMPLIANT_MANUALLY', comment: string) => void;
	} = $props();

	let status: 'EVALUATION_STATUS_COMPLIANT_MANUALLY' | 'EVALUATION_STATUS_NOT_COMPLIANT_MANUALLY' = $state('EVALUATION_STATUS_COMPLIANT_MANUALLY');
	let comment = $state('');
	let saving = $state(false);

	function generateUUID(): string {
		return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
			const r = Math.random() * 16 | 0;
			const v = c === 'x' ? r : (r & 0x3 | 0x8);
			return v.toString(16);
		});
	}

	async function handleSave() {
		saving = true;
		try {
			const body = {
				id: generateUUID(),
				targetOfEvaluationId: targetId,
				auditScopeId: auditScopeId,
				controlId: control.id,
				controlCategoryName: control.categoryName ?? '',
				controlCatalogId: control.categoryCatalogId ?? '',
				parentControlId: control.parentControlId ?? null,
				status: status,
				assessmentResultIds: [],
				comment: comment || undefined,
				timestamp: new Date().toISOString()
			};
			console.log('Saving evaluation:', body);
			const response = await orchestratorClient().POST('/v1/evaluation/results', { body });
			console.log('Response:', response);
			if (!response.response.ok) {
				const errText = await response.response.text();
				console.error('Error:', response.response.status, errText);
				alert('Failed: ' + response.response.status + ' ' + errText);
				saving = false;
				return;
			}
			console.log('Saved successfully!');
			onSave?.(status, comment);
			await invalidateAll();
			open = false;
			status = 'EVALUATION_STATUS_COMPLIANT_MANUALLY';
			comment = '';
		} catch (e) {
			alert('Failed to save: ' + e);
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
				<h3 class="text-lg font-semibold">Manual Evaluation: {control.id}</h3>
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
					class="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
					onclick={handleSave}
				>
					Save
				</button>
			</div>
		</div>
	</div>
{/if}