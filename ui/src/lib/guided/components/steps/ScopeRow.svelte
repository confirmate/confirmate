<script lang="ts">
	import { Button, Chip, Segmented, TextArea, type Tone } from '$lib/ds';
	import type { Applicability } from '../../profile';
	import type { ScopeRow, Treatment } from '../../scope';

	let {
		row,
		ondecide,
		onreason,
		ondismiss
	}: {
		row: ScopeRow;
		ondecide: (treatment: Treatment, reason: string) => void;
		onreason: (reason: string) => void;
		ondismiss: () => void;
	} = $props();

	const tags: Record<Applicability, { label: string; tone: Tone }> = {
		mandatory: { label: 'Mandatory', tone: 'ok' },
		capability: { label: 'Only if the capability exists', tone: 'neutral' },
		risk: { label: 'Risk based', tone: 'brand' }
	};
	const options = [
		{ value: 'applies', label: 'Applies' },
		{ value: 'na', label: 'Not applicable' },
		{ value: 'risk', label: 'Accepted risk' }
	];
	const suggestionLabel = { na: 'not applicable', risk: 'accepted residual risk' };

	const id = $derived(row.requirement.id);
</script>

<div class="flex flex-col gap-3 border-t border-line px-6 py-4 first:border-t-0 {row.suggestion || row.missingReason ? 'bg-warn-soft/40' : ''}">
	<div class="flex flex-wrap items-start justify-between gap-x-6 gap-y-3">
		<div class="flex min-w-0 flex-[1_1_22rem] flex-col gap-1">
			<div class="flex flex-wrap items-center gap-2">
				<strong class="font-medium">{row.requirement.title}</strong>
				<Chip size="sm" tone={tags[row.requirement.applicability].tone}>{tags[row.requirement.applicability].label}</Chip>
			</div>
			<span class="text-sm text-ink-soft">{row.requirement.text}</span>
			<span class="text-xs text-ink-muted">
				<span class="font-code">{id}</span>, {row.requirement.checks === 1 ? '1 check' : `${row.requirement.checks} checks`}
			</span>
		</div>

		<div class="flex flex-col items-end gap-1.5">
			{#if row.locked}
				<span class="flex min-h-9 items-center text-[0.8125rem] text-ink-muted">Mandatory, cannot be taken out</span>
			{:else}
				<Segmented
					label="Treatment of {row.requirement.title}"
					{options}
					value={row.treatment}
					onchange={(v) => ondecide(v as Treatment, row.reason)}
				/>
			{/if}
		</div>
	</div>

	{#if row.suggestion}
		<div class="flex flex-wrap items-center gap-x-4 gap-y-2 rounded-control bg-brand-soft px-3.5 py-2.5 text-sm text-brand-ink">
			<span class="min-w-[18rem] flex-1">
				<strong class="font-semibold">Suggested: {suggestionLabel[row.suggestion.treatment]}.</strong>
				{row.suggestion.reason}
			</span>
			<span class="flex gap-2">
				<Button size="sm" onclick={() => ondecide(row.suggestion!.treatment, row.suggestion!.reason)}>Accept</Button>
				<Button size="sm" variant="quiet" onclick={ondismiss}>Keep it</Button>
			</span>
		</div>
	{/if}

	{#if row.source === 'rule' && row.treatment === 'na'}
		<p class="text-[0.8125rem] text-ink-muted">
			Derived from your answers. {row.reason} Change it if that is not true.
		</p>
	{/if}

	{#if row.treatment !== 'applies'}
		<div class="flex flex-col gap-1.5">
			<label for="reason-{id}" class="text-sm font-medium">
				{row.treatment === 'risk' ? 'Residual risk that you accept' : 'Why is it not applicable?'}
			</label>
			<TextArea
				id="reason-{id}"
				rows={2}
				value={row.reason}
				placeholder={row.treatment === 'risk' ? 'Describe the risk and why it is acceptable.' : 'Give the reason in a sentence.'}
				oninput={onreason}
			/>
			{#if row.missingReason}<span class="text-[0.8125rem] text-bad">A reason is required.</span>{/if}
			{#if row.treatment === 'risk'}
				<span class="text-[0.8125rem] text-ink-muted">Shown to users in the section on known limitations of the user information.</span>
			{/if}
		</div>
	{/if}
</div>
