<script lang="ts">
	import { Chip, FilterChips } from '$lib/ds';
	import type { FrameworkProfile } from '../../profile';
	import { computeScope, summarize, type ScopeGroup } from '../../scope';
	import { wizard } from '../../wizard.svelte';
	import ScopeRow from './ScopeRow.svelte';

	let { profile }: { profile: FrameworkProfile } = $props();

	const groups = $derived(computeScope(wizard.groups, wizard.risk, wizard.answers, wizard.decisions));
	const summary = $derived(summarize(groups));
	const titles = $derived(new Map(wizard.groups.flatMap((g) => g.requirements.map((r) => [r.id, r.title] as const))));

	const depends = 'Depends on the product';
	const filterNames = [depends, 'Need a decision', 'Not applicable', 'Accepted risk', 'All'];
	let filter = $state(depends);

	function visible(group: ScopeGroup) {
		return group.rows.filter(
			(r) =>
				filter === 'All' ||
				(filter === depends && r.requirement.applicability.kind !== 'always') ||
				(filter === 'Need a decision' && (r.missingReason || r.missingCompensation || group.empty)) ||
				(filter === 'Not applicable' && r.treatment === 'na') ||
				(filter === 'Accepted risk' && r.treatment === 'risk')
		);
	}

	const shown = $derived(groups.map((g) => ({ ...g, visible: visible(g) })).filter((g) => g.visible.length > 0));
	const stats = $derived([
		{ value: summary.applies, label: 'apply' },
		{ value: summary.na, label: 'not applicable' },
		{ value: summary.risk, label: 'accepted residual risk' },
		{ value: summary.review, label: 'need your decision' }
	]);
</script>

<div class="flex flex-col gap-6">
	<section class="flex flex-wrap items-center gap-x-7 gap-y-3 rounded-card border border-line bg-surface p-(--cf-card-pad)">
		{#each stats as s (s.label)}
			<div class="flex flex-col">
				<span class="font-display text-figure font-(--cf-display-weight) leading-tight {s.label === 'need your decision' && s.value > 0 ? 'text-warn-ink' : ''}">{s.value}</span>
				<span class="text-sm text-ink-muted">{s.label}</span>
			</div>
		{/each}
		<span class="ml-auto max-w-sm text-sm text-ink-muted">
			of {summary.total} requirements of the {profile.name}. {summary.total - summary.depends} always apply and are listed
			under "All". The others depend on your product.
		</span>
	</section>

	<FilterChips label="Filter" options={filterNames} bind:value={filter} />

	<div class="flex flex-col gap-5">
		{#each shown as g (g.group.id)}
			<section class="overflow-hidden rounded-card border border-line bg-surface">
				<div class="flex flex-wrap items-baseline justify-between gap-x-4 gap-y-1 border-b border-line bg-sunken px-6 py-3.5">
					<div class="flex flex-col">
						<h2 class="text-[1.0625rem] font-semibold">{g.group.title}</h2>
						<span class="text-[0.8125rem] text-ink-muted">{g.group.reference}</span>
					</div>
					<span class="text-[0.8125rem] text-ink-soft">
						{g.rows.filter((r) => r.treatment === 'applies').length} apply,
						{g.rows.filter((r) => r.treatment === 'na').length} not applicable,
						{g.rows.filter((r) => r.treatment === 'risk').length} accepted risk
					</span>
				</div>
				{#if g.empty}
					<div class="bg-bad-soft px-6 py-2.5 text-sm text-bad-ink" role="alert">
						Your answers took out every requirement of this point. A point cannot drop out as a whole: keep at least one
						requirement, and if you do not implement it, accept the residual risk.
					</div>
				{/if}
				{#each g.visible as row (row.requirement.id)}
					<ScopeRow
						{row}
						titleOf={(id) => titles.get(id) ?? id}
						ondecide={(t, reason, naReason, compensatedBy) => wizard.decide(row.requirement.id, t, reason, naReason, compensatedBy)}
						onreason={(reason) => wizard.setReason(row.requirement.id, reason)}
						onreset={() => wizard.reset(row.requirement.id)}
					/>
				{/each}
			</section>
		{:else}
			<p class="rounded-card border border-line bg-surface px-6 py-5 text-sm text-ink-muted">Nothing to show with this filter.</p>
		{/each}
	</div>

	<label class="flex cursor-pointer items-start gap-3 {summary.review > 0 ? 'opacity-60' : ''}">
		<input
			type="checkbox"
			bind:checked={wizard.approved}
			disabled={summary.review > 0}
			class="mt-0.5 size-5.5 accent-brand"
		/>
		<span>
			I have read the decisions and the reasons and take responsibility for them.
			{#if summary.review > 0}
				<Chip size="sm" tone="warn">{summary.review} still need a decision or a reason</Chip>
			{/if}
		</span>
	</label>
</div>
