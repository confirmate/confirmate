<script lang="ts">
	import type { FrameworkProfile, ProfileStep } from '../profile';
	import { wizard } from '../wizard.svelte';
	import ChoiceResult from './steps/ChoiceResult.svelte';
	import ChoiceStep from './steps/ChoiceStep.svelte';
	import DetailsStep from './steps/DetailsStep.svelte';
	import NextSteps from './steps/NextSteps.svelte';
	import QuestionnaireStep from './steps/QuestionnaireStep.svelte';
	import ScopePreview from './steps/ScopePreview.svelte';
	import ScopeStep from './steps/ScopeStep.svelte';
	import SourcesStep from './steps/SourcesStep.svelte';
	import WizardShell from './WizardShell.svelte';

	let { profile, step }: { profile: FrameworkProfile; step: ProfileStep } = $props();
</script>

{#snippet aside()}
	{#if step.type === 'choice'}
		<ChoiceResult {profile} />
	{:else if step.type === 'questionnaire'}
		<ScopePreview {profile} />
	{:else if step.type === 'details'}
		<NextSteps />
	{/if}
{/snippet}

<WizardShell
	{profile}
	stepId={step.id}
	nextDisabled={step.type === 'scope' && !wizard.approved}
	aside={step.type === 'sources' || step.type === 'scope' ? undefined : aside}
>
	{#if step.type === 'details'}
		<DetailsStep {profile} />
	{:else if step.type === 'choice'}
		<ChoiceStep {profile} />
	{:else if step.type === 'questionnaire'}
		<QuestionnaireStep {profile} />
	{:else if step.type === 'scope'}
		<ScopeStep {profile} />
	{:else if step.type === 'sources'}
		<SourcesStep {profile} />
	{/if}
</WizardShell>
