<script lang="ts">
	import { OptionCard } from '$lib/ds';
	import type { FrameworkProfile } from '../../profile';
	import { wizard } from '../../wizard.svelte';
	import AssetsEditor from './AssetsEditor.svelte';
	import ConditionList from './ConditionList.svelte';
	import EnvironmentFields from './EnvironmentFields.svelte';

	let { profile }: { profile: FrameworkProfile } = $props();
</script>

<div class="flex flex-col gap-7">
	<section class="flex flex-col gap-3" aria-labelledby="preset-heading">
		<div class="flex flex-col gap-1">
			<h2 id="preset-heading" class="text-[1.0625rem] font-semibold">Start from what looks like your product</h2>
			<p class="text-sm text-ink-muted">You can change everything below. The risk assessment is yours.</p>
		</div>
		<div role="radiogroup" aria-labelledby="preset-heading" class="grid grid-cols-1 gap-2.5 md:grid-cols-2">
			{#each profile.risk.presets as preset (preset.id)}
				<OptionCard
					label={preset.label}
					hint={preset.hint}
					selected={wizard.presetId === preset.id}
					onselect={() => wizard.usePreset(profile, preset.id)}
				/>
			{/each}
		</div>
	</section>

	<AssetsEditor {profile} />
	<EnvironmentFields />
	<ConditionList />
</div>
