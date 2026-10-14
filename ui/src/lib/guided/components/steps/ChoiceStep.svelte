<script lang="ts">
	import { OptionCard } from '$lib/ds';
	import type { FrameworkProfile } from '../../profile';
	import { wizard } from '../../wizard.svelte';

	let { profile }: { profile: FrameworkProfile } = $props();
</script>

<div role="radiogroup" aria-label={profile.steps.find((s) => s.type === 'choice')?.title} class="flex flex-col gap-2.5">
	{#each profile.choices as choice (choice.id)}
		<OptionCard
			label={choice.label}
			hint={choice.hint}
			selected={wizard.choiceId === choice.id}
			onselect={() => (wizard.choiceId = choice.id)}
		/>
	{/each}
</div>
