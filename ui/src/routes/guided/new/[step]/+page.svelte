<script lang="ts">
	import { wizard } from '$lib/guided/wizard.svelte';
	import ChoiceResult from '$lib/guided/components/steps/ChoiceResult.svelte';
	import ChoiceStep from '$lib/guided/components/steps/ChoiceStep.svelte';
	import DetailsStep from '$lib/guided/components/steps/DetailsStep.svelte';
	import QuestionnaireStep from '$lib/guided/components/steps/QuestionnaireStep.svelte';
	import ScopePreview from '$lib/guided/components/steps/ScopePreview.svelte';
	import ScopeStep from '$lib/guided/components/steps/ScopeStep.svelte';
	import SourcesStep from '$lib/guided/components/steps/SourcesStep.svelte';
	import WizardShell from '$lib/guided/components/WizardShell.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const profile = $derived(data.profile);
	const step = $derived(data.step);
</script>

{#snippet aside()}
	{#if step.type === 'choice'}
		<ChoiceResult {profile} />
	{:else if step.type === 'questionnaire'}
		<ScopePreview {profile} />
	{:else if step.type === 'details'}
		<div class="flex flex-col gap-2.5 rounded-card bg-brand-soft p-6">
			<h2 class="text-base font-semibold text-brand-ink">What happens next</h2>
			<p class="text-ink-soft">
				You answer a few questions about the product. We work out its class, ask for a short risk assessment and
				then show which requirements apply. Nothing is evaluated before you confirm the scope.
			</p>
		</div>
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
