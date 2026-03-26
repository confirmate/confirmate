<script lang="ts">
	import type { ComplianceStatus } from '$lib/api/evaluation';
	import type { Catalog, AuditScope } from '$lib/api/orchestrator';
	import { Trash } from '@steeze-ui/heroicons';
	import { Icon } from '@steeze-ui/svelte-icon';
	import Button from './Button.svelte';
	import ComplianceChart from './ComplianceChart.svelte';

	interface Props {
		catalog: Catalog;
		auditScope: AuditScope;
		compliance: Map<string, ComplianceStatus>;
		onremove?: (detail: { auditScope: AuditScope }) => void;
	}

	let { catalog, auditScope, compliance, onremove }: Props = $props();

	function remove() {
		onremove?.({ auditScope });
	}
</script>

<li class="overflow-hidden rounded-xl border border-gray-200">
	<div class="flex items-center justify-between gap-x-4 border-b border-gray-900/5 bg-gray-50 p-6">
		<a href={'/cloud/' + auditScope.targetOfEvaluationId + '/compliance/' + catalog.id}>
			<div class="text-sm leading-6 font-medium text-gray-900">{catalog.name}</div>
			<div class="text-sm text-gray-500">{catalog.description}</div>
		</a>
		<div class="flex gap-x-1.5">
			<Button onclick={remove} class="bg-red-800 hover:bg-red-700">
				<Icon src={Trash} class="h-4 w-4" />
			</Button>
		</div>
	</div>

	<dl class="-my-3 divide-y divide-gray-100 px-6 py-4 text-sm leading-6">
		<ComplianceChart {compliance} {auditScope} />
		<div class="flex justify-between gap-x-4 py-3">
			<dt class="text-gray-500">Controls in Scope</dt>
			<dd class="flex items-start gap-x-2">
				<div class="font-medium text-gray-900">{compliance.size}</div>
			</dd>
		</div>
		<div class="flex justify-between gap-x-4 py-3">
			<dt class="text-gray-500">Assurance Level</dt>
			<dd class="flex items-start gap-x-2">
				<div class="font-medium text-gray-900">
					{auditScope.assuranceLevel ?? '-'}
				</div>
			</dd>
		</div>
	</dl>
</li>
