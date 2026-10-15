<script lang="ts">
	import { Button, Chip, Segmented, Select, TextArea, type Tone } from '$lib/ds';
	import type { ApplicabilityKind } from '../../profile';
	import { describeScenario } from '../../risk';
	import { notApplicableReasons, type NotApplicableReason, type ScopeRow, type Treatment } from '../../scope';

	let {
		row,
		titleOf,
		ondecide,
		onreason,
		onreset
	}: {
		row: ScopeRow;
		/** The title of a requirement by its ID, for the requirements that can replace this one. */
		titleOf: (id: string) => string;
		ondecide: (treatment: Treatment, reason: string, naReason?: NotApplicableReason, compensatedBy?: string) => void;
		onreason: (reason: string) => void;
		onreset: () => void;
	} = $props();

	const tags: Record<ApplicabilityKind, { label: string; tone: Tone }> = {
		always: { label: 'Always applies', tone: 'ok' },
		conditional: { label: 'Depends on a condition', tone: 'neutral' },
		risk: { label: 'Depends on the risk', tone: 'brand' }
	};
	const options = $derived([
		{ value: 'applies', label: 'Applies' },
		{
			value: 'na',
			label: 'Not applicable',
			disabled: row.naBlocked && row.treatment !== 'na',
			title: 'At least one requirement of each point has to remain.'
		},
		{ value: 'risk', label: 'Accepted risk' }
	]);

	const id = $derived(row.requirement.id);
	const applicability = $derived(row.requirement.applicability);
	const legal = $derived(row.requirement.references.find((r) => r.relationship === 'legal-basis'));
	const when = $derived(
		applicability.kind === 'conditional'
			? `Applies if this is true: ${applicability.condition}.`
			: applicability.kind === 'risk' && applicability.scenario
				? describeScenario(applicability.scenario)
				: ''
	);
	// Reasons that need a requirement that replaces this one are only offered if the catalog names one.
	const reasons = $derived(
		notApplicableReasons
			.filter((r) => r.value !== 'COMPENSATED' || applicability.compensatedBy.length > 0)
			.map((r) => ({ value: r.value, label: r.label }))
	);
	const compensations = $derived(applicability.compensatedBy.map((c) => ({ value: c, label: titleOf(c) })));
</script>

<div class="flex flex-col gap-3 border-t border-line px-6 py-4 first:border-t-0 {row.missingReason || row.missingCompensation ? 'bg-warn-soft/40' : ''}">
	<div class="flex flex-wrap items-start justify-between gap-x-6 gap-y-3">
		<div class="flex min-w-0 flex-[1_1_22rem] flex-col gap-1">
			<div class="flex flex-wrap items-center gap-2">
				<strong class="font-medium">{row.requirement.title}</strong>
				<Chip size="sm" tone={tags[applicability.kind].tone}>{tags[applicability.kind].label}</Chip>
				<Chip size="sm">{row.requirement.type}</Chip>
			</div>
			<span class="text-sm text-ink-soft">{row.requirement.text}</span>
			{#if when}<span class="text-[0.8125rem] text-ink-muted">{when}</span>{/if}
			<span class="text-xs text-ink-muted">
				{#if legal}<span>{legal.framework} {legal.identifier}</span><span>, </span>{/if}<span class="font-code">{id}</span>{#if row.requirement.checks > 0}<span>, </span><span>{row.requirement.checks === 1 ? '1 check' : `${row.requirement.checks} checks`}</span>{/if}
			</span>
		</div>

		<div class="flex flex-col items-end gap-1.5">
			{#if row.locked}
				<span class="flex min-h-9 items-center text-[0.8125rem] text-ink-muted">Always applies, cannot be taken out</span>
			{:else}
				<Segmented
					label="Treatment of {row.requirement.title}"
					{options}
					value={row.treatment}
					onchange={(v) =>
						ondecide(
							v as Treatment,
							row.treatment === 'applies' ? row.rule.reason : row.reason,
							v === 'na' ? (row.rule.naReason ?? 'TARGET_ABSENT') : undefined
						)}
				/>
			{/if}
			{#if row.naBlocked && !row.locked}
				<span class="max-w-60 text-right text-xs text-ink-muted">
					The last requirement of its point. If you do not implement it, accept the residual risk.
				</span>
			{/if}
		</div>
	</div>

	{#if row.source === 'rule' && row.treatment === 'na'}
		<p class="text-[0.8125rem] text-ink-muted">Derived from your risk profile and answers. Change it if that is not true.</p>
	{/if}

	{#if row.source === 'person' && row.rule.treatment === 'na' && row.treatment === 'applies'}
		<p class="flex flex-wrap items-center gap-x-3 text-[0.8125rem] text-ink-muted">
			Your risk profile or answers would take it out. You decided to keep it.
			<Button size="sm" variant="quiet" onclick={onreset}>Use the rules again</Button>
		</p>
	{/if}

	{#if row.treatment === 'na'}
		<div class="flex flex-col gap-1.5">
			<label for="reason-type-{id}" class="text-sm font-medium">Why is it not applicable?</label>
			<div class="max-w-lg">
				<Select
					id="reason-type-{id}"
					options={reasons}
					value={row.naReason ?? ''}
					onchange={(v) => ondecide('na', row.reason, v as NotApplicableReason, v === 'COMPENSATED' ? row.compensatedBy : undefined)}
				/>
			</div>
			{#if row.naReason === 'COMPENSATED'}
				<label for="compensation-{id}" class="mt-1 text-sm font-medium">Which requirement is fulfilled instead?</label>
				<div class="max-w-lg">
					<Select
						id="compensation-{id}"
						options={compensations}
						value={row.compensatedBy ?? ''}
						onchange={(v) => ondecide('na', row.reason, 'COMPENSATED', v)}
					/>
				</div>
				{#if row.missingCompensation}<span class="text-[0.8125rem] text-bad">Choose the requirement that replaces this one.</span>{/if}
			{/if}
		</div>
	{/if}

	{#if row.treatment !== 'applies'}
		<div class="flex flex-col gap-1.5">
			<label for="reason-{id}" class="text-sm font-medium">
				{row.treatment === 'risk' ? 'Residual risk that you accept' : 'Explain it in a sentence'}
			</label>
			<TextArea
				id="reason-{id}"
				rows={2}
				value={row.reason}
				placeholder={row.treatment === 'risk' ? 'Describe the risk and why it is acceptable.' : 'Say why, so that an auditor understands it.'}
				oninput={onreason}
			/>
			{#if row.missingReason}<span class="text-[0.8125rem] text-bad">A reason is required.</span>{/if}
			{#if row.treatment === 'risk'}
				<span class="text-[0.8125rem] text-ink-muted">Shown to users in the section on known limitations of the user information.</span>
			{/if}
		</div>
	{/if}
</div>
