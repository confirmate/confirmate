<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { Button, Dialog, Field, OptionCard, Select, TextArea } from '$lib/ds';
	import { notApplicableReasons, type NotApplicableReason } from '../scope';
	import { guidedSource } from '../select';
	import type { RequirementDetail } from '../source';

	let {
		open = $bindable(false),
		targetId,
		requirement
	}: { open: boolean; targetId: string; requirement: RequirementDetail } = $props();

	let treatment = $state<'applies' | 'na' | 'risk'>('na');
	let reason = $state('');
	let naReason = $state<NotApplicableReason>('TARGET_ABSENT');
	let compensatedBy = $state('');
	// A requirement can only be compensated if the catalog names the requirements that replace it.
	const reasons = $derived(
		notApplicableReasons
			.filter((r) => r.value !== 'COMPENSATED' || requirement.compensatedBy.length > 0)
			.map((r) => ({ value: r.value, label: r.label }))
	);
	let busy = $state(false);
	let error = $state<string | null>(null);

	async function save() {
		if (treatment === 'applies') {
			open = false;
			return;
		}
		if (!reason.trim()) {
			error = 'Please give a reason.';
			return;
		}
		if (treatment === 'na' && naReason === 'COMPENSATED' && !compensatedBy) {
			error = 'Please choose the requirement that is fulfilled instead.';
			return;
		}
		busy = true;
		error = null;
		try {
			await guidedSource.changeScope({
				targetId,
				requirementId: requirement.id,
				treatment,
				reason: reason.trim(),
				naReason: treatment === 'na' ? naReason : undefined,
				compensatedBy: treatment === 'na' && naReason === 'COMPENSATED' ? compensatedBy : undefined
			});
			await invalidateAll();
			reason = '';
			open = false;
		} catch (e) {
			error = e instanceof Error ? e.message : 'The change could not be saved.';
		} finally {
			busy = false;
		}
	}
</script>

<Dialog bind:open title={requirement.title}>
	<p class="-mt-3 text-sm text-ink-muted">
		<span class="font-code">{requirement.reference}</span>
	</p>

	<div role="radiogroup" aria-label="What is true for this requirement?" class="flex flex-col gap-2">
		<OptionCard
			label="It applies"
			hint="Keep it in the scope and answer its checks."
			selected={treatment === 'applies'}
			onselect={() => (treatment = 'applies')}
		/>
		<OptionCard
			label="Not applicable"
			hint="The risk assessment shows that the product has no risk that this requirement addresses."
			selected={treatment === 'na'}
			onselect={() => (treatment = 'na')}
		/>
		<OptionCard
			label="Applies, but I accept the residual risk"
			hint="The risk stays. You describe it and it is shown in the user information."
			selected={treatment === 'risk'}
			onselect={() => (treatment = 'risk')}
		/>
	</div>

	{#if treatment === 'na'}
		<Field label="Why is it not applicable?" id="scope-na-reason">
			<Select id="scope-na-reason" options={reasons} bind:value={naReason} />
		</Field>
		{#if naReason === 'COMPENSATED'}
			<Field label="Which requirement is fulfilled instead?" id="scope-compensation">
				<Select
					id="scope-compensation"
					options={requirement.compensatedBy.map((c) => ({ value: c, label: c }))}
					bind:value={compensatedBy}
				/>
			</Field>
		{/if}
	{/if}

	{#if treatment !== 'applies'}
		<Field label="Explanation" id="scope-reason">
			<TextArea id="scope-reason" bind:value={reason} rows={3} />
		</Field>
		<p class="rounded-control bg-sunken px-3.5 py-3 text-sm text-ink-soft">
			<strong class="font-semibold">What happens:</strong> a new version of the risk assessment is created with all earlier
			decisions and this one. When it is approved, the requirement leaves the scope, its results are removed and your
			reason is written to the audit trail.
		</p>
	{/if}

	{#if error}<p class="text-sm text-bad">{error}</p>{/if}

	{#snippet footer()}
		<Button variant="quiet" onclick={() => (open = false)}>Cancel</Button>
		<Button onclick={save} disabled={busy}>{busy ? 'Saving…' : treatment === 'applies' ? 'Close' : 'Create a new version and approve'}</Button>
	{/snippet}
</Dialog>
