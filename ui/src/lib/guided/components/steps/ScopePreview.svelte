<script lang="ts">
	import type { FrameworkProfile } from '../../profile';
	import { computeScope, summarize } from '../../scope';
	import { wizard } from '../../wizard.svelte';

	let { profile }: { profile: FrameworkProfile } = $props();

	// What the risk profile and the answers do, before a person decides anything.
	const groups = $derived(computeScope(wizard.groups, wizard.risk, wizard.answers, { decided: {} }));
	const summary = $derived(summarize(groups));
	const taken = $derived(
		groups
			.map((g) => ({
				title: g.group.title,
				count: g.rows.filter((r) => r.treatment === 'na').length,
				total: g.rows.length
			}))
			.filter((g) => g.count > 0)
	);
</script>

<div class="sticky top-6 flex flex-col gap-3.5 rounded-card border border-line bg-surface p-(--cf-card-pad)">
	<div class="flex flex-col gap-1">
		<span class="text-[0.8125rem] uppercase tracking-wider text-ink-muted">{profile.risk.heading}</span>
		<span class="font-display text-heading font-(--cf-display-weight)">{summary.total - summary.na} of {summary.total} apply</span>
		<span class="text-[0.8125rem] text-ink-muted">
			{summary.na} drop out with this risk profile. You confirm each one in the next step. An answer of "don't know" keeps a
			requirement in.
		</span>
	</div>
	<ul class="m-0 flex list-none flex-col p-0">
		{#each taken as g (g.title)}
			<li class="flex items-start justify-between gap-3 border-t border-line py-2 text-sm">
				<span>{g.title}</span>
				<span class="whitespace-nowrap text-warn-ink">{g.count} of {g.total}</span>
			</li>
		{:else}
			<li class="border-t border-line py-2 text-sm text-ink-muted">Every requirement applies.</li>
		{/each}
	</ul>
</div>
