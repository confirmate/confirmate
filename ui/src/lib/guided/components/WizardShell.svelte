<script lang="ts">
	import type { Snippet } from 'svelte';
	import { resolve } from '$app/paths';
	import { Button, Stepper } from '$lib/ds';
	import type { StepState } from '$lib/ds/Stepper.svelte';
	import type { FrameworkProfile } from '../profile';

	let {
		profile,
		stepId,
		nextDisabled = false,
		children,
		aside
	}: {
		profile: FrameworkProfile;
		stepId: string;
		nextDisabled?: boolean;
		children: Snippet;
		aside?: Snippet;
	} = $props();

	const index = $derived(profile.steps.findIndex((s) => s.id === stepId));
	const step = $derived(profile.steps[index]);
	const isLast = $derived(index === profile.steps.length - 1);

	const stepsView = $derived(
		profile.steps.map((s, i) => ({
			id: s.id,
			label: s.label,
			state: (i < index ? 'done' : i === index ? 'current' : 'todo') as StepState,
			href: resolve(`/guided/new/${s.id}/`)
		}))
	);

	const backHref = $derived(
		index <= 0 ? resolve('/guided/new/') : resolve(`/guided/new/${profile.steps[index - 1].id}/`)
	);
	const nextHref = $derived(
		isLast ? resolve('/guided/') : resolve(`/guided/new/${profile.steps[index + 1].id}/`)
	);
</script>

<div class="flex min-h-[calc(100vh-4rem)] flex-wrap font-body text-ink">
	<aside class="box-border w-75 shrink-0 border-r border-line bg-surface px-8 py-12">
		<Stepper steps={stepsView} heading="Setup, about 10 minutes" />
	</aside>

	<main class="box-border flex min-w-0 max-w-3xl flex-[999_1_34rem] flex-col gap-7 px-14 py-14">
		<div class="flex flex-col gap-2">
			<h1 class="font-display text-title font-(--cf-display-weight) leading-tight">{step.title}</h1>
			<p class="text-[1.0625rem] text-ink-muted">{step.intro}</p>
		</div>

		{@render children()}

		<div class="flex flex-wrap items-center gap-3">
			<Button href={nextHref} disabled={nextDisabled}>{isLast ? 'Finish setup' : 'Continue'}</Button>
			<Button variant="quiet" href={backHref}>Back</Button>
		</div>
	</main>

	{#if aside}
		<aside class="box-border max-w-md flex-[1_1_20rem] px-10 py-14">
			{@render aside()}
		</aside>
	{/if}
</div>
