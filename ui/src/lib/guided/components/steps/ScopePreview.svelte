<script lang="ts">
	import { Chip } from '$lib/ds';
	import type { FrameworkProfile } from '../../profile';
	import { computeScope } from '../../scope';
	import { wizard } from '../../wizard.svelte';

	let { profile }: { profile: FrameworkProfile } = $props();

	const scope = $derived(computeScope(profile.questionnaire, wizard.answers));
</script>

<div class="sticky top-6 flex flex-col gap-3.5 rounded-card border border-line bg-surface p-6">
	<div class="flex flex-col gap-1">
		<span class="text-[0.8125rem] uppercase tracking-wider text-ink-muted">{profile.questionnaire.heading}</span>
		<span class="font-display text-heading font-(--cf-display-weight)">{scope.applyCount} of {scope.requirements.length} apply</span>
		<span class="text-[0.8125rem] text-ink-muted">
			You see why a requirement drops out, and you approve it in the next step. Unknown answers keep it in.
		</span>
	</div>
	<ul class="m-0 flex list-none flex-col p-0">
		{#each scope.requirements as r (r.id)}
			<li class="flex items-start gap-3 border-t border-line py-2">
				<span class="w-5.5 shrink-0 font-code text-ink-muted">{r.id}</span>
				<span class="flex flex-1 flex-col">
					<span class={r.out ? 'text-ink-muted line-through' : 'text-ink'}>{r.title}</span>
					{#if r.out}<span class="text-[0.8125rem] text-warn-ink">{r.reason}</span>{/if}
				</span>
				<Chip size="sm" tone={r.out ? 'warn' : 'ok'}>{r.out ? 'Out' : 'Applies'}</Chip>
			</li>
		{/each}
	</ul>
</div>
