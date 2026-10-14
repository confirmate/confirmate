<script lang="ts">
	import { resolve } from '$app/paths';
	import { Card, Chip } from '$lib/ds';
	import CheckCard from '$lib/guided/components/CheckCard.svelte';
	import StatusTag from '$lib/guided/components/StatusTag.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const target = $derived(data.target!);
	const r = $derived(data.requirement);
</script>

<div class="flex flex-col gap-7">
	<div class="flex max-w-3xl flex-col gap-2.5">
		<a href={resolve(`/guided/${target.id}/requirements/`)} class="text-sm no-underline">Requirements</a>
		<h1 class="font-display text-4xl leading-tight">{r.title}</h1>
		<div class="flex flex-wrap gap-2">
			<StatusTag status={r.status} />
			<Chip>{r.reference}</Chip>
			<Chip>{r.scopeReason}</Chip>
		</div>
		<p class="text-[1.0625rem] text-ink-soft">{r.plain}</p>
	</div>

	<div class="flex flex-wrap items-start gap-6">
		<div class="flex min-w-0 flex-[3_1_34rem] flex-col gap-4">
			{#each r.checks as check (check.id)}
				<CheckCard {check} />
			{/each}
			<p class="text-sm text-ink-muted">
				Think this requirement does not apply? <a href={resolve(`/guided/${target.id}/requirements/${r.id}/`)}>Take it out of scope</a>.
				That needs a reason and an approval, and it is added to the risk assessment.
			</p>
		</div>

		<aside class="flex min-w-0 flex-[2_1_18rem] flex-col gap-6">
			<Card>
				<h2 class="text-base font-semibold">History</h2>
				<ul class="m-0 flex list-none flex-col p-0 text-sm">
					{#each r.history as h (h.when + h.what)}
						<li class="border-t border-line py-2.5"><span class="text-ink-muted">{h.when}</span><br />{h.what}</li>
					{/each}
				</ul>
			</Card>
			<Card>
				<h2 class="text-base font-semibold">Comments</h2>
				{#each r.comments as c (c.who + c.text)}
					<p class="text-sm text-ink-soft"><strong class="font-medium">{c.who}:</strong> {c.text}</p>
				{/each}
			</Card>
		</aside>
	</div>
</div>
