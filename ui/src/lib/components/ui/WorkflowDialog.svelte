<script lang="ts">
	import { XMark } from '@steeze-ui/heroicons';
	import { Icon } from '@steeze-ui/svelte-icon';
	import { invalidateAll } from '$app/navigation';
	import { orchestratorClient } from '$lib/api/client';
	import type { SchemaControlInScope } from '$lib/api/openapi/orchestrator';

	type State = NonNullable<SchemaControlInScope['state']>;

	const STATE_LABELS: Record<State, string> = {
		CONTROL_IN_SCOPE_STATE_UNSPECIFIED: 'Unspecified',
		CONTROL_IN_SCOPE_STATE_OPEN: 'Open',
		CONTROL_IN_SCOPE_STATE_IN_PROGRESS: 'In Progress',
		CONTROL_IN_SCOPE_STATE_IMPLEMENTED: 'Implemented',
		CONTROL_IN_SCOPE_STATE_READY_FOR_REVIEW: 'Ready for Review',
		CONTROL_IN_SCOPE_STATE_ACCEPTED: 'Accepted'
	};

	const ALL_STATES: State[] = [
		'CONTROL_IN_SCOPE_STATE_OPEN',
		'CONTROL_IN_SCOPE_STATE_IN_PROGRESS',
		'CONTROL_IN_SCOPE_STATE_IMPLEMENTED',
		'CONTROL_IN_SCOPE_STATE_READY_FOR_REVIEW',
		'CONTROL_IN_SCOPE_STATE_ACCEPTED'
	];

	// Valid transitions per state — mirrors the backend state machine in workflow.go
	const VALID_TRANSITIONS: Partial<Record<State, State[]>> = {
		CONTROL_IN_SCOPE_STATE_OPEN: ['CONTROL_IN_SCOPE_STATE_IN_PROGRESS'],
		CONTROL_IN_SCOPE_STATE_IN_PROGRESS: ['CONTROL_IN_SCOPE_STATE_OPEN', 'CONTROL_IN_SCOPE_STATE_IMPLEMENTED'],
		CONTROL_IN_SCOPE_STATE_IMPLEMENTED: ['CONTROL_IN_SCOPE_STATE_IN_PROGRESS', 'CONTROL_IN_SCOPE_STATE_READY_FOR_REVIEW'],
		CONTROL_IN_SCOPE_STATE_READY_FOR_REVIEW: ['CONTROL_IN_SCOPE_STATE_IN_PROGRESS', 'CONTROL_IN_SCOPE_STATE_ACCEPTED'],
		CONTROL_IN_SCOPE_STATE_ACCEPTED: ['CONTROL_IN_SCOPE_STATE_IN_PROGRESS']
	};

	let {
		open = $bindable(false),
		controlInScope,
		controlName
	}: {
		open: boolean;
		controlInScope: SchemaControlInScope;
		controlName: string;
	} = $props();

	const availableTransitions = $derived(
		VALID_TRANSITIONS[(controlInScope.state as State) ?? 'CONTROL_IN_SCOPE_STATE_OPEN'] ?? []
	);

	let toState = $state<State>('CONTROL_IN_SCOPE_STATE_OPEN');
	let comment = $state('');
	let saving = $state(false);

	let errorMsg = $state<string | null>(null);

	$effect(() => {
		if (open) {
			toState = availableTransitions[0] ?? 'CONTROL_IN_SCOPE_STATE_OPEN';
			comment = '';
			errorMsg = null;
		}
	});

	async function handleSave() {
		if (!comment.trim() || saving) return;
		saving = true;
		errorMsg = null;
		try {
			const { error } = await orchestratorClient().POST(
				'/v1/orchestrator/controls_in_scope/{id}/transition',
				{
					params: { path: { id: controlInScope.id! } },
					body: { id: controlInScope.id!, toState, comment }
				}
			);
			if (error) {
				errorMsg = (error as { message?: string }).message ?? 'Failed to transition state';
				return;
			}
			await invalidateAll();
			open = false;
		} catch {
			errorMsg = 'Failed to transition state';
		} finally {
			saving = false;
		}
	}
</script>

{#if open}
	<div class="fixed inset-0 z-50 flex items-center justify-center">
		<div class="absolute inset-0 bg-black/50" onclick={() => (open = false)}></div>
		<div class="relative z-10 w-full max-w-lg rounded-lg bg-white p-6 shadow-xl">
			<div class="mb-4 flex items-center justify-between">
				<h3 class="text-lg font-semibold">Workflow State: {controlName}</h3>
				<button type="button" class="text-gray-400 hover:text-gray-600" onclick={() => (open = false)}>
					<Icon src={XMark} class="h-5 w-5" />
				</button>
			</div>

			<div class="mb-4">
				<label class="mb-2 block text-sm font-medium text-gray-700">New State</label>
				<div class="grid grid-cols-1 gap-2">
					{#each ALL_STATES as state}
						{@const available = availableTransitions.includes(state)}
						{@const current = controlInScope.state === state}
						<label
							class="flex items-center gap-3 rounded-lg border p-3
								{available ? 'cursor-pointer hover:bg-gray-50' : 'cursor-not-allowed opacity-40'}
								{toState === state && available ? 'border-blue-500 bg-blue-50' : 'border-gray-200'}"
							title={available ? '' : 'Not reachable from the current state'}
						>
							<input
								type="radio"
								bind:group={toState}
								value={state}
								disabled={!available}
								class="text-blue-600 disabled:cursor-not-allowed"
							/>
							<span class="text-sm font-medium {toState === state && available ? 'text-blue-700' : 'text-gray-700'}">
								{STATE_LABELS[state]}
							</span>
							{#if current}
								<span class="ml-auto rounded-full bg-gray-100 px-2 py-0.5 text-xs text-gray-500">current</span>
							{/if}
						</label>
					{/each}
				</div>
				{#if availableTransitions.length < ALL_STATES.length - 1}
					<p class="mt-2 text-xs text-gray-400">Grayed-out states must be reached in order — complete earlier steps first.</p>
				{/if}
			</div>

			<div class="mb-6">
				<label class="mb-2 block text-sm font-medium text-gray-700">
					Comment <span class="font-normal text-red-500">*</span>
				</label>
				<textarea
					bind:value={comment}
					rows="3"
					class="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
					placeholder="Explain this state transition..."
				></textarea>
			</div>

			<div class="flex justify-end gap-3">
				<button
					type="button"
					class="rounded-md px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-100"
					onclick={() => (open = false)}
				>
					Cancel
				</button>
				<button
					type="button"
					disabled={saving || !comment.trim()}
					class="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
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
