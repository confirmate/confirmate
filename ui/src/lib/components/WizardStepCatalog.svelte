<script lang="ts">
	import type { Catalog, AuditScope } from '$lib/api/orchestrator';
	import AssuranceLevelPopover, {
		type AssuranceLevelEvent
	} from '$lib/components/AssuranceLevelPopover.svelte';
	import type { WizardData } from './Wizard.svelte';

	interface Props {
		data: WizardData;
	}

	let { data = $bindable() }: Props = $props();

	// Reactive property for the selection status of all catalogs
	let selected = $derived(
		new Map(
			data.catalogs.map((catalog) => [
				catalog.id,
				data.auditScopes.find((auditScope) => auditScope.catalogId == catalog.id) !== undefined
			])
		)
	);

	function assuranceLevelSelected(e: CustomEvent<AssuranceLevelEvent>) {
		toggle(e.detail.catalog, e.detail.assuranceLevel);
	}

	/**
	 * This function selects a catalog if its not already selected or de-selects
	 * an already selected catalog.
	 *
	 * If the catalog needs an assurance level, this must be specified as an extra
	 * parameter.
	 *
	 * @param catalog the catalog to (de)-select
	 * @param assuranceLevel the assurance level, if the catalog needs it
	 */
	function toggle(catalog: Catalog, assuranceLevel?: string) {
		// Check, if catalog already exists in the Audit Scope
		if (!selected.get(catalog.id)) {
			// Does not exist yet -> create new Audit Scope
			const auditScope: AuditScope = {
				catalogId: catalog.id,
				// This will not be the final ID, since we do not know it at this point.
				// This needs to be set by the caller of save()
				targetOfEvaluationId: data.service.id,
				assuranceLevel: assuranceLevel
			};

			data.auditScopes = [...data.auditScopes, auditScope];
		} else {
			// Already exists -> remove it from the Audit Scope list
			data.auditScopes = data.auditScopes.filter(
				(auditScope) => auditScope.catalogId != catalog.id
			);
		}
	}
</script>

<ul class="mt-5 grid grid-cols-1 gap-5 sm:grid-cols-2 sm:gap-6 lg:grid-cols-2">
	{#each data.catalogs as catalog (catalog.id)}
		<li class="col-span-1 flex rounded-md shadow-xs">
			{#if catalog.assuranceLevels.length > 0 && !selected.get(catalog.id)}
				<AssuranceLevelPopover {catalog} on:select={assuranceLevelSelected}>
					<button
						class={selected.get(catalog.id)
							? ''
							: 'flex h-full w-18 shrink-0 items-center justify-center rounded-l-md bg-gray-400 text-sm text-white'}
						style={selected.get(catalog.id) ? 'background-color: ' + catalog.metadata.color : ''}
					>
						{catalog.shortName}
					</button>
				</AssuranceLevelPopover>
			{:else}
				<button
					onclick={() => toggle(catalog)}
					class="{selected.get(catalog.id) ? '' : 'bg-gray-400'}
       flex w-18 shrink-0 items-center justify-center rounded-l-md text-sm font-medium text-white"
					style={selected.get(catalog.id) ? 'background-color: ' + catalog.metadata.color : ''}
				>
					{catalog.shortName}
				</button>
			{/if}
			<div
				class="flex flex-1 items-center justify-between truncate rounded-r-md border-t border-r border-b border-gray-200 bg-white"
			>
				<div class="flex-1 truncate px-4 py-2 text-sm">
					<p class="font-medium text-gray-900 hover:text-gray-600">
						{catalog.name}
					</p>
					<p class="text-xs whitespace-normal text-gray-500">
						{catalog.description}
					</p>
				</div>
			</div>
		</li>
	{/each}
</ul>
