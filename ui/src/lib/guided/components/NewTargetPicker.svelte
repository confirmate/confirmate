<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { Button, Chip, PageHeading } from '$lib/ds';
	import type { FrameworkProfile } from '../profile';
	import { plannedProfiles, profiles } from '../profiles';
	import { wizard } from '../wizard.svelte';

	function start(profile: FrameworkProfile) {
		wizard.start(profile);
		goto(resolve(`/guided/new/${profile.steps[0].id}/`));
	}
</script>

<main class="mx-auto flex w-full max-w-275 flex-col gap-9 px-10 pt-18 pb-20">
	<PageHeading
		title="What do you want to check?"
		intro="The type of Target of Evaluation (ToE) decides the questions we ask and the framework that fits. You can add more frameworks to a ToE later."
	/>

	<div class="grid grid-cols-[repeat(auto-fit,minmax(min(18.75rem,100%),1fr))] gap-5">
		{#each profiles as profile (profile.id)}
			<button
				type="button"
				onclick={() => start(profile)}
				class="flex cursor-pointer flex-col items-start gap-3 rounded-card border border-line-strong bg-surface p-6.5 text-left hover:border-brand hover:bg-brand-soft/40"
			>
				<Chip tone="brand">{profile.name}</Chip>
				<h2 class="font-display text-heading font-(--cf-display-weight) leading-tight">{profile.pitch.title}</h2>
				<p class="text-ink-soft">{profile.pitch.text}</p>
			</button>
		{/each}
		{#each plannedProfiles as planned (planned.id)}
			<div class="flex flex-col items-start gap-3 rounded-card border border-line bg-surface p-6.5 opacity-70">
				<span class="flex gap-2"><Chip tone="brand">{planned.name}</Chip><Chip size="sm">Not available yet</Chip></span>
				<h2 class="font-display text-heading font-(--cf-display-weight) leading-tight">{planned.title}</h2>
				<p class="text-ink-soft">{planned.text}</p>
			</div>
		{/each}
	</div>

	<div><Button variant="quiet" href={resolve('/guided/')}>Cancel</Button></div>
</main>
