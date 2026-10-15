<script lang="ts">
	import { Button, OptionCard, Tabs } from '$lib/ds';
	import { conditionsOf } from '../../catalog';
	import type { FrameworkProfile } from '../../profile';
	import { wizard } from '../../wizard.svelte';
	import AssetsEditor from './AssetsEditor.svelte';
	import ConditionList from './ConditionList.svelte';
	import EnvironmentFields from './EnvironmentFields.svelte';

	let { profile }: { profile: FrameworkProfile } = $props();

	let tab = $state('assets');

	// What the person has said in each part, shown on the tab so that nothing is missed
	const environmentSet = $derived(
		[wizard.risk.physicalAccess, wizard.risk.connectivity, wizard.risk.operatorSkill].filter(Boolean).length
	);
	const tabs = $derived([
		{ id: 'assets', label: 'Assets', badge: String(wizard.risk.assets.length) },
		{ id: 'environment', label: 'Environment', badge: `${environmentSet} of 3` },
		{ id: 'questions', label: 'Questions', badge: String(conditionsOf(wizard.groups).length) }
	]);
	const next = $derived(tabs[tabs.findIndex((t) => t.id === tab) + 1]);
</script>

<div class="flex flex-col gap-6">
	<section class="flex flex-col gap-3" aria-labelledby="preset-heading">
		<div class="flex flex-col gap-1">
			<h2 id="preset-heading" class="text-[1.0625rem] font-semibold">Start from what looks like your product</h2>
			<p class="text-sm text-ink-muted">You can change everything in the tabs below. The risk assessment is yours.</p>
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

	<div class="flex flex-col gap-5">
		<Tabs label="Parts of the risk assessment" {tabs} bind:value={tab} />

		<div role="tabpanel" id="panel-assets" aria-labelledby="tab-assets" hidden={tab !== 'assets'}>
			{#if tab === 'assets'}<AssetsEditor {profile} />{/if}
		</div>
		<div role="tabpanel" id="panel-environment" aria-labelledby="tab-environment" hidden={tab !== 'environment'}>
			{#if tab === 'environment'}<EnvironmentFields />{/if}
		</div>
		<div role="tabpanel" id="panel-questions" aria-labelledby="tab-questions" hidden={tab !== 'questions'}>
			{#if tab === 'questions'}<ConditionList />{/if}
		</div>

		{#if next}
			<div><Button variant="quiet" onclick={() => (tab = next.id)}>Next: {next.label}</Button></div>
		{/if}
	</div>
</div>
