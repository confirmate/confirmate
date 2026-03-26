<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import Button from './Button.svelte';
	import type { WizardData } from './Wizard.svelte';

	interface Props {
		data: WizardData;
	}

	let { data }: Props = $props();

	const dispatch = createEventDispatcher<{ save: WizardData; cancel: {} }>();

	function save() {
		dispatch('save', data);
	}

	function cancel() {
		dispatch('cancel');
	}
</script>

<div class="mb-5 text-sm">
	{#if data.service.name.length == 0}
		Please provide at least a name for the new Target of Evaluation.
	{:else}
		{#if data.mode == 'create'}
			This will create a new Target of Evaluation called <b>{data.service.name}</b> with
		{:else}
			This will change the Target of Evaluation called <b>{data.service.name}</b> with
		{/if}
		{#if data.auditScopes.length > 0}
			the following Targets of Evaluation:
			<ul class="mt-1 ml-4 list-disc">
				{#each data.auditScopes as auditScope}
					<li>
						{auditScope.catalogId}
						{#if auditScope.assuranceLevel !== undefined}
							({auditScope.assuranceLevel})
						{/if}
					</li>
				{/each}
			</ul>
		{:else}
			no configured Target of Evaluation.
		{/if}
	{/if}
</div>

<Button on:click={save} disabled={data.service.name.length == 0}>
	{#if data.mode == 'create'}
		Create
	{:else}
		Save
	{/if}
</Button>
<Button on:click={cancel} class="ml-2 bg-red-800 hover:bg-red-700">Cancel</Button>
