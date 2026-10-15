<script lang="ts">
	import { resolve } from '$app/paths';
	import { Button, Card, Chip, PageHeading } from '$lib/ds';
	import type { AcceptedRisk, TargetDetail } from '../source';

	let { target, risks }: { target: TargetDetail; risks: AcceptedRisk[] } = $props();
</script>

<div class="flex max-w-4xl flex-col gap-6">
	<PageHeading
		title="Accepted residual risks"
		intro="Risks that you decided to accept for a requirement. The CRA expects users to be told about them, so they go into the user information."
	/>

	{#if risks.length === 0}
		<Card>
			<p class="text-ink-soft">
				No residual risk is accepted. If a requirement applies but a risk stays, choose "accepted risk" for it in the risk
				assessment.
			</p>
		</Card>
	{:else}
		<Card padding="sm">
			<ul class="m-0 flex list-none flex-col p-0">
				{#each risks as r (r.id)}
					<li class="flex flex-wrap gap-x-6 gap-y-2 border-t border-line px-3 py-4 first:border-t-0">
						<div class="flex min-w-[18rem] flex-1 flex-col gap-1">
							<strong class="font-medium">{r.requirement}</strong>
							<span class="text-[0.8125rem] text-ink-muted"><span class="font-code">{r.id}</span>{r.reference ? `, ${r.reference}` : ''}</span>
							<p class="mt-1.5 text-ink-soft">{r.text}</p>
						</div>
						<div class="flex w-56 flex-col text-sm text-ink-soft">
							<span>Accepted by {r.acceptedBy}</span>
							<span>{r.acceptedAt}</span>
							<span class="text-ink-muted">Review by {r.reviewBy}</span>
						</div>
					</li>
				{/each}
			</ul>
		</Card>

		<Card>
			<div class="flex flex-wrap items-center justify-between gap-3">
				<h2 class="text-lg font-semibold">Preview in the user information</h2>
				<Chip size="sm" tone="ok">Kept up to date automatically</Chip>
			</div>
			<div class="flex flex-col gap-2 border-l-[3px] border-line-strong py-1 pl-4.5 text-ink-soft">
				<strong class="font-semibold text-ink">Known limitations</strong>
				{#each risks as r (r.id)}<p>{r.text}</p>{/each}
			</div>
		</Card>
	{/if}

	<div>
		<Button variant="quiet" href={resolve(`/guided/${target.id}/requirements/`)}>Back to the requirements</Button>
	</div>
</div>
