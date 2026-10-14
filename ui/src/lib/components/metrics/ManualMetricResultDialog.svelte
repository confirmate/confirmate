<script lang="ts">
	import { XMark } from '@steeze-ui/heroicons';
	import { Icon } from '@steeze-ui/svelte-icon';
	import { invalidateAll } from '$app/navigation';
	import { orchestratorClient } from '$lib/api/client';

	let {
		open = $bindable(false),
		metric,
		targetId
	}: {
		open: boolean;
		metric: { id?: string; name?: string };
		targetId: string;
	} = $props();

	let compliant = $state(true);
	let comment = $state('');
	let evidenceId = $state('');
	let validUntil = $state('');
	let saving = $state(false);
	let errorMsg = $state<string | null>(null);

	function reset() {
		compliant = true;
		comment = '';
		evidenceId = '';
		validUntil = '';
		errorMsg = null;
	}

	async function handleSave() {
		if (saving) return;
		if (!comment.trim()) {
			errorMsg = 'Please give a reason for the result';
			return;
		}
		saving = true;
		errorMsg = null;
		try {
			const { error } = await orchestratorClient().POST('/v1/orchestrator/manual_assessment_results', {
				body: {
					targetOfEvaluationId: targetId,
					metricId: metric.id ?? '',
					compliant,
					comment: comment.trim(),
					evidenceId: evidenceId.trim() || undefined,
					validUntil: validUntil ? new Date(validUntil).toISOString() : undefined
				}
			});
			if (error) {
				errorMsg = (error as { message?: string }).message ?? 'Failed to save the result';
				return;
			}
			await invalidateAll();
			open = false;
			reset();
		} catch {
			errorMsg = 'Failed to save the result';
		} finally {
			saving = false;
		}
	}

	function handleCancel() {
		open = false;
		reset();
	}
</script>

{#if open}
	<div class="fixed inset-0 z-50 flex items-center justify-center">
		<div class="absolute inset-0 bg-black/50" onclick={handleCancel}></div>
		<div class="relative z-10 w-full max-w-lg rounded-lg bg-white p-6 shadow-xl">
			<div class="mb-4 flex items-center justify-between">
				<h3 class="text-lg font-semibold">Manual result: {metric.name || metric.id}</h3>
				<button type="button" class="text-gray-400 hover:text-gray-600" onclick={handleCancel}>
					<Icon src={XMark} class="h-5 w-5" />
				</button>
			</div>

			<p class="mb-4 text-sm text-gray-500">
				A manual result is additional information. As soon as an automatic result exists for this
				metric, the automatic result takes precedence.
			</p>

			<div class="mb-4">
				<label class="mb-2 block text-sm font-medium text-gray-700">Is the metric met?</label>
				<div class="flex gap-4">
					<label class="flex items-center gap-2">
						<input type="radio" bind:group={compliant} value={true} class="text-emerald-600" />
						<span class="text-sm">Compliant</span>
					</label>
					<label class="flex items-center gap-2">
						<input type="radio" bind:group={compliant} value={false} class="text-red-600" />
						<span class="text-sm">Not compliant</span>
					</label>
				</div>
			</div>

			<div class="mb-4">
				<label class="mb-2 block text-sm font-medium text-gray-700" for="manual-comment">
					Reason
				</label>
				<textarea
					id="manual-comment"
					bind:value={comment}
					rows="4"
					class="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
					placeholder="Why is this metric met or not met?"
				></textarea>
			</div>

			<div class="mb-4">
				<label class="mb-2 block text-sm font-medium text-gray-700" for="manual-evidence">
					Evidence ID <span class="font-normal text-gray-500">(optional)</span>
				</label>
				<input
					id="manual-evidence"
					bind:value={evidenceId}
					class="w-full rounded-md border border-gray-300 px-3 py-2 font-mono text-sm focus:border-blue-500 focus:outline-none"
					placeholder="ID of a stored evidence, e.g. a document"
				/>
			</div>

			<div class="mb-6">
				<label class="mb-2 block text-sm font-medium text-gray-700" for="manual-valid-until">
					Valid until <span class="font-normal text-gray-500">(optional)</span>
				</label>
				<input
					id="manual-valid-until"
					type="date"
					bind:value={validUntil}
					class="rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none"
				/>
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
