<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { Button, Dialog, Field, Segmented, TextArea, TextInput } from '$lib/ds';
	import { guidedSource } from '../select';
	import type { Check } from '../source';

	let {
		open = $bindable(false),
		targetId,
		check
	}: { open: boolean; targetId: string; check?: Check } = $props();

	let answer = $state('yes');
	let comment = $state('');
	let validUntil = $state('');
	let busy = $state(false);
	let error = $state<string | null>(null);

	async function save() {
		if (!check?.metricId) return;
		if (!comment.trim()) {
			error = 'Please give a reason.';
			return;
		}
		busy = true;
		error = null;
		try {
			await guidedSource.addStatement({
				targetId,
				metricId: check.metricId,
				compliant: answer === 'yes',
				comment: comment.trim(),
				validUntil: validUntil || undefined
			});
			await invalidateAll();
			comment = '';
			validUntil = '';
			answer = 'yes';
			open = false;
		} catch (e) {
			error = e instanceof Error ? e.message : 'The statement could not be saved.';
		} finally {
			busy = false;
		}
	}
</script>

<Dialog bind:open title={check?.title ?? 'Statement'}>
	<p class="text-sm text-ink-muted">
		A statement is additional information. An automatic result of the same check always wins over it.
	</p>
	<Field label="Is this met?" id="statement-answer">
		<Segmented
			label="Is this met?"
			options={[
				{ value: 'yes', label: 'Yes' },
				{ value: 'no', label: 'No' }
			]}
			bind:value={answer}
		/>
	</Field>
	<Field label="Reason" id="statement-comment" hint="Why is it met or not met? Name the evidence if you have one.">
		<TextArea id="statement-comment" bind:value={comment} />
	</Field>
	<Field label="Valid until (optional)" id="statement-valid-until">
		<TextInput id="statement-valid-until" type="date" bind:value={validUntil} />
	</Field>
	{#if error}<p class="text-sm text-bad">{error}</p>{/if}
	{#snippet footer()}
		<Button variant="quiet" onclick={() => (open = false)}>Cancel</Button>
		<Button onclick={save} disabled={busy}>{busy ? 'Saving…' : 'Save'}</Button>
	{/snippet}
</Dialog>
