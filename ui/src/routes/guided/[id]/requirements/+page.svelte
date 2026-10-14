<script lang="ts">
	import { resolve } from '$app/paths';
	import { Card } from '$lib/ds';
	import StatusTag from '$lib/guided/components/StatusTag.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const target = $derived(data.target!);
	const parts = $derived(
		target.parts.map((p) => ({ ...p, items: target.requirements.filter((r) => r.partId === p.id) })).filter((p) => p.items.length)
	);
</script>

<div class="flex max-w-4xl flex-col gap-6">
	<div class="flex flex-col gap-2">
		<h1 class="font-display text-4xl leading-tight">Requirements</h1>
		<p class="text-ink-muted">
			Grouped by the parts of the Cyber Resilience Act. The preview lists a sample of the {target.progress.total}
			requirements in scope.
		</p>
	</div>

	{#each parts as part (part.id)}
		<Card>
			<div class="flex flex-wrap items-baseline justify-between gap-3">
				<h2 class="text-lg font-semibold">{part.label}</h2>
				<span class="text-sm text-ink-muted">{part.met} of {part.total} met</span>
			</div>
			<ul class="m-0 flex list-none flex-col p-0">
				{#each part.items as r (r.id)}
					<li class="flex items-center gap-4 border-t border-line py-3">
						<span class="flex flex-1 flex-col">
							<a href={resolve(`/guided/${target.id}/requirements/${r.id}/`)} class="font-medium no-underline">{r.title}</a>
							<span class="text-[0.8125rem] text-ink-muted">{r.reference}</span>
						</span>
						<StatusTag status={r.status} />
					</li>
				{/each}
			</ul>
		</Card>
	{/each}
</div>
