<script lang="ts">
	import { XMark } from '@steeze-ui/heroicons';
	import { Icon } from '@steeze-ui/svelte-icon';

	let {
		open = $bindable(false),
		controlName,
		action,
		onconfirm
	}: {
		open: boolean;
		controlName: string;
		action: 'add' | 'remove';
		onconfirm: (comment: string) => void;
	} = $props();

	let comment = $state('');
	let selectedReason = $state('');

	const removeReasons = [
		'Not applicable — risk analysis determined this control is not relevant',
		'Not applicable — covered by existing organizational processes',
		'Managed by external sub-contractor or third party',
		'Compensating control in place',
		'Inherited from underlying platform or infrastructure',
		'Deferred — to be addressed in a future release',
		'Other'
	];

	$effect(() => {
		if (open) {
			comment = '';
			selectedReason = '';
		}
	});

	function submit() {
		const finalComment = selectedReason === 'Other'
			? comment.trim()
			: selectedReason || comment.trim();
		if (!finalComment) return;
		onconfirm(finalComment);
		open = false;
	}

	function isOther() {
		return selectedReason === 'Other';
	}
</script>

{#if open}
	<div class="fixed inset-0 z-50 flex items-center justify-center">
		<div class="absolute inset-0 bg-black/50" onclick={() => (open = false)}></div>
		<div class="relative z-10 w-full max-w-md rounded-lg bg-white p-6 shadow-xl">
			<div class="mb-4 flex items-center justify-between">
				<h3 class="text-base font-semibold text-gray-900">
					{action === 'remove' ? 'Remove from scope' : 'Add to scope'}: {controlName}
				</h3>
				<button type="button" class="text-gray-400 hover:text-gray-600" onclick={() => (open = false)}>
					<Icon src={XMark} class="h-5 w-5" />
				</button>
			</div>

			<div class="mb-5">
				<label class="mb-1.5 block text-sm font-medium text-gray-700">
					Reason <span class="font-normal text-red-500">*</span>
				</label>
				{#if action === 'remove'}
					<div class="mb-3 space-y-1.5">
						{#each removeReasons as reason}
							<label class="flex cursor-pointer items-start gap-2 rounded-md px-2 py-1.5 text-sm hover:bg-gray-50 {selectedReason === reason ? 'bg-blue-50' : ''}">
								<input
									type="radio"
									name="reason"
									value={reason}
									bind:group={selectedReason}
									class="mt-0.5"
								/>
								<span class="text-gray-700">{reason}</span>
							</label>
						{/each}
					</div>
					{#if isOther()}
						<textarea
							bind:value={comment}
							rows="3"
							autofocus
							class="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
							placeholder="Please specify the reason..."
							onkeydown={(e) => { if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) submit(); }}
						></textarea>
					{/if}
				{:else}
					<textarea
						bind:value={comment}
						rows="3"
						autofocus
						class="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
						placeholder="Why is this control being added to scope?"
						onkeydown={(e) => { if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) submit(); }}
					></textarea>
				{/if}
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
					disabled={action === 'remove' ? !selectedReason || (isOther() && !comment.trim()) : !comment.trim()}
					class="rounded-md px-4 py-2 text-sm font-medium text-white disabled:opacity-50 {action === 'remove' ? 'bg-red-600 hover:bg-red-700' : 'bg-blue-600 hover:bg-blue-700'}"
					onclick={submit}
				>
					{action === 'remove' ? 'Remove from scope' : 'Add to scope'}
				</button>
			</div>
		</div>
	</div>
{/if}
