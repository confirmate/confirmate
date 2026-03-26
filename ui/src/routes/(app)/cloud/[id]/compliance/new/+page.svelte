<script lang="ts">
	import { goto, invalidate } from '$app/navigation';
	import { startEvaluation } from '$lib/api/evaluation';
	import { createAuditScope } from '$lib/api/orchestrator';
	import type { WizardData } from '$lib/components/Wizard.svelte';
	import WizardStepCatalog from '$lib/components/WizardStepCatalog.svelte';
	import WizardStepSave from '$lib/components/WizardStepSave.svelte';
	import type { PageData } from './$types';

	interface Props {
		data: PageData;
	}

	let { data }: Props = $props();

	// Create a new wizard data object to store the catalogs that will be enabled
	let wizard = $state({
		service: data.service,
		catalogs: data.leftOverCatalogs,
		auditScopes: [],
		mode: 'edit'
	} satisfies WizardData);

	async function save(event: CustomEvent<WizardData>) {
		// Afterwards, create the Targets of Evaluation
		let auditScopes = await Promise.all(
			event.detail.auditScopes.map((auditScope) => {
				auditScope.name = 'UI generated Audit Scope';
				return createAuditScope(auditScope);
			})
		);

		await Promise.all(auditScopes.map((auditScope) => startEvaluation(auditScope)));
		await invalidate((url) => url.pathname === `/v1/orchestrator/audit_scopes`);
		goto('../');
	}
</script>

<WizardStepCatalog bind:data={wizard} />
<div class="mt-4"></div>
<WizardStepSave bind:data={wizard} on:cancel={() => goto('../')} on:save={save} />
