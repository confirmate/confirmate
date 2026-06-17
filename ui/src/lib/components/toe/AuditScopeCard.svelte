<script lang="ts">
	import { ChevronRight } from '@steeze-ui/heroicons';
	import { Icon } from '@steeze-ui/svelte-icon';
	import type { SchemaAuditScope, SchemaCatalog } from '$lib/api/openapi/orchestrator';

	let {
		auditScope,
		catalog
	}: {
		auditScope: SchemaAuditScope;
		catalog: SchemaCatalog | undefined;
	} = $props();
</script>

<a
	href="/toe/{auditScope.targetOfEvaluationId}/audit-scopes/{auditScope.id}/"
	class="flex items-center justify-between rounded-lg border border-gray-200 bg-white p-4 hover:border-confirmate hover:bg-blue-50"
>
	<div>
		<div class="font-medium text-gray-900">{auditScope.name}</div>
		{#if catalog}
			<div class="mt-0.5 text-sm text-gray-500">{catalog.name}</div>
		{/if}
		{#if auditScope.assuranceLevel}
			<div class="mt-1">
				<span class="inline-flex items-center rounded-full bg-blue-50 px-2 py-0.5 text-xs font-medium text-blue-700">
					{auditScope.assuranceLevel}
				</span>
			</div>
		{/if}
	</div>
	<Icon src={ChevronRight} class="h-5 w-5 shrink-0 text-gray-400" />
</a>
