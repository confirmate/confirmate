<script lang="ts">
	import { Segmented } from '$lib/ds';
	import { conditionsOf } from '../../catalog';
	import type { Answer } from '../../profile';
	import { wizard } from '../../wizard.svelte';

	const conditions = $derived(conditionsOf(wizard.groups));
	const options = [
		{ value: 'yes', label: 'Yes' },
		{ value: 'no', label: 'No' },
		{ value: 'unknown', label: "Don't know" }
	];
</script>

{#if conditions.length > 0}
	<section class="flex flex-col rounded-card border border-line bg-surface px-7 py-2" aria-labelledby="conditions-heading">
		<div class="flex flex-col gap-1 pt-4">
			<h2 id="conditions-heading" class="text-[1.0625rem] font-semibold">What is true for the product?</h2>
			<p class="text-sm text-ink-muted">
				Some requirements only apply in a certain case. A requirement is only taken out if you answer no. "Don't know"
				keeps it.
			</p>
		</div>
		{#each conditions as condition (condition.id)}
			<div class="flex flex-wrap items-center justify-between gap-5 border-b border-line py-5 last:border-b-0">
				<div class="flex min-w-[20rem] flex-1 flex-col gap-0.5">
					<strong class="font-medium">{condition.text}.</strong>
					<span class="text-[0.8125rem] text-ink-muted">
						{condition.requirements.length === 1 ? '1 requirement depends' : `${condition.requirements.length} requirements depend`} on it.
						{#if wizard.fromChoice[condition.id]}Suggested by the kind of product you chose.{/if}
					</span>
				</div>
				<Segmented
					label={condition.text}
					{options}
					value={wizard.answers[condition.id]}
					onchange={(v) => wizard.answer(condition.id, v as Answer)}
				/>
			</div>
		{/each}
	</section>
{/if}
