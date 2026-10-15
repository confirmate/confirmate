<script lang="ts">
	import type { FrameworkProfile } from '../../profile';
	import { computeScope, summarize } from '../../scope';
	import { wizard } from '../../wizard.svelte';

	let { profile }: { profile: FrameworkProfile } = $props();

	const groups = $derived(computeScope(profile.questionnaire, wizard.answers, { decided: {}, dismissed: {} }));
	const summary = $derived(summarize(groups));
	const candidates = $derived(
		groups
			.map((g) => ({
				title: g.group.title,
				count: g.rows.filter((r) => r.treatment === 'na' || r.suggestion).length,
				total: g.rows.length
			}))
			.filter((g) => g.count > 0)
	);
	const total = $derived(candidates.reduce((n, g) => n + g.count, 0));
</script>

<div class="sticky top-6 flex flex-col gap-3.5 rounded-card border border-line bg-surface p-(--cf-card-pad)">
	<div class="flex flex-col gap-1">
		<span class="text-[0.8125rem] uppercase tracking-wider text-ink-muted">{profile.questionnaire.heading}</span>
		<span class="font-display text-heading font-(--cf-display-weight)">{summary.total - total} of {summary.total} stay in</span>
		<span class="text-[0.8125rem] text-ink-muted">
			{total} could be taken out. You decide for each requirement in the next step. Unknown answers keep a requirement in.
		</span>
	</div>
	<ul class="m-0 flex list-none flex-col p-0">
		{#each candidates as g (g.title)}
			<li class="flex items-start justify-between gap-3 border-t border-line py-2 text-sm">
				<span>{g.title}</span>
				<span class="whitespace-nowrap text-warn-ink">{g.count} of {g.total}</span>
			</li>
		{:else}
			<li class="border-t border-line py-2 text-sm text-ink-muted">Nothing is suggested to be taken out.</li>
		{/each}
	</ul>
</div>
