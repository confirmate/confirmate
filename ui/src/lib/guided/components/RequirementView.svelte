<script lang="ts">
	import { resolve } from '$app/paths';
	import { Card, Chip, PageHeading } from '$lib/ds';
	import type { RequirementDetail, TargetDetail } from '../source';
	import CheckCard from './CheckCard.svelte';
	import ScopeDialog from './ScopeDialog.svelte';
	import StatementDialog from './StatementDialog.svelte';
	import StatusTag from './StatusTag.svelte';

	import type { Check } from '../source';

	let { target, requirement: r }: { target: TargetDetail; requirement: RequirementDetail } = $props();

	let scopeOpen = $state(false);
	let statementOpen = $state(false);
	let statementCheck = $state<Check | undefined>();

	function onaction(check: Check, actionId: string) {
		if (actionId !== 'statement') return;
		statementCheck = check;
		statementOpen = true;
	}
</script>

<div class="flex flex-col gap-7">
	<PageHeading title={r.title} intro={r.plain}>
		{#snippet above()}
			<a href={resolve(`/guided/${target.id}/requirements/`)} class="text-sm no-underline">Requirements</a>
		{/snippet}
	</PageHeading>

	<div class="-mt-4 flex flex-wrap gap-2">
		<StatusTag status={r.status} />
		<Chip>{r.reference}</Chip>
		<Chip>{r.scopeReason}</Chip>
	</div>

	<div class="flex flex-wrap items-start gap-6">
		<div class="flex min-w-0 flex-[3_1_34rem] flex-col gap-4">
			{#each r.checks as check (check.id)}
				<CheckCard {check} {onaction} />
			{/each}
			<p class="text-sm text-ink-muted">
				Think this requirement does not apply, or does a risk stay? You can take it out of scope with a reason. That creates a new
				version of the risk assessment. <button type="button" class="cursor-pointer text-brand underline" onclick={() => (scopeOpen = true)}>Change the scope</button>.
			</p>
		</div>

		<aside class="flex min-w-0 flex-[2_1_18rem] flex-col gap-6">
			{#if r.history.length}
			<Card>
				<h2 class="text-base font-semibold">History</h2>
				<ul class="m-0 flex list-none flex-col p-0 text-sm">
					{#each r.history as h (h.when + h.what)}
						<li class="border-t border-line py-2.5"><span class="text-ink-muted">{h.when}</span><br />{h.what}</li>
					{/each}
				</ul>
			</Card>
			{/if}
			{#if r.comments.length}
			<Card>
				<h2 class="text-base font-semibold">Comments</h2>
				{#each r.comments as c (c.who + c.text)}
					<p class="text-sm text-ink-soft"><strong class="font-medium">{c.who}:</strong> {c.text}</p>
				{/each}
			</Card>
			{/if}
		</aside>
	</div>
</div>

<StatementDialog bind:open={statementOpen} targetId={target.id} check={statementCheck} />
<ScopeDialog bind:open={scopeOpen} targetId={target.id} requirement={r} />
