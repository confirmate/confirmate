<script lang="ts">
	import { Chip } from '$lib/ds';
	import type { FrameworkProfile } from '../../profile';
	import { computeScope } from '../../scope';
	import { wizard } from '../../wizard.svelte';

	let { profile }: { profile: FrameworkProfile } = $props();

	const scope = $derived(computeScope(profile.questionnaire, wizard.answers));
	const outCount = $derived(scope.requirements.filter((r) => r.out).length);
</script>

<div class="flex flex-col gap-7">
	<p class="text-ink-soft">
		{scope.applyCount} of {scope.requirements.length} product requirements apply.
		{#if outCount === 0}None drops out.{:else}{outCount} drop{outCount === 1 ? 's' : ''} out, and the reason is recorded.{/if}
	</p>

	<section class="overflow-hidden rounded-card border border-line bg-surface">
		<ul class="m-0 flex list-none flex-col p-0">
			{#each scope.requirements as r (r.id)}
				<li class="flex items-start gap-4 border-t border-line px-6 py-4 first:border-t-0 {r.out ? 'bg-warn-soft/40' : ''}">
					<span class="w-5.5 shrink-0 font-code text-ink-muted">{r.id}</span>
					<span class="flex flex-1 flex-col gap-1.5">
						<strong class="font-medium">{r.title}</strong>
						{#if r.out}<span class="text-sm text-ink-soft">Reason: {r.reason}</span>{/if}
					</span>
					<Chip size="sm" tone={r.out ? 'warn' : 'ok'}>{r.out ? 'Out of scope' : 'Applies'}</Chip>
				</li>
			{/each}
		</ul>
	</section>

	<section class="flex flex-col gap-1.5 rounded-card border border-line bg-surface px-7 py-5">
		<h2 class="text-base font-semibold">These always apply and cannot be taken out</h2>
		<p class="text-ink-soft">{profile.alwaysApplies.text} Together {profile.alwaysApplies.count} requirements.</p>
	</section>

	<label class="flex cursor-pointer items-start gap-3">
		<input type="checkbox" bind:checked={wizard.approved} class="mt-0.5 size-5.5 accent-brand" />
		<span>
			I have read the reasons{outCount > 0 ? ` for the ${outCount} requirement${outCount === 1 ? '' : 's'} that drop out` : ''} and take
			responsibility for this decision.
		</span>
	</label>
</div>
