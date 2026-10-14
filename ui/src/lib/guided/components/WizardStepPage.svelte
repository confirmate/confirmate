<script lang="ts">
	import { goto, invalidateAll } from '$app/navigation';
	import { resolve } from '$app/paths';
	import type { FrameworkProfile, ProfileStep } from '../profile';
	import { guidedSource } from '../select';
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

	let busy = $state(false);
	let error = $state<string | null>(null);

	/** Creates the Target of Evaluation from the answers of the wizard and opens it. */
	async function finish() {
		busy = true;
		error = null;
		try {
			const created = await guidedSource.createTarget({
				profile,
				details: { ...wizard.details },
				choiceId: wizard.choiceId,
				answers: { ...wizard.answers }
			});
			await invalidateAll();
			wizard.start(profile);
			await goto(resolve(`/guided/${created.id}/`));
		} catch (e) {
			error = e instanceof Error ? e.message : 'The setup failed.';
		} finally {
			busy = false;
		}
	}
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
	onfinish={step.type === 'sources' ? finish : undefined}
	{busy}
	{error}
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
