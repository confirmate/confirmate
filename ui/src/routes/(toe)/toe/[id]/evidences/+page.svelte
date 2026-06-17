<script lang="ts">
	import type { SchemaEvidence } from '$lib/api/openapi/evidence';
	import SectionHeader from '$lib/components/ui/SectionHeader.svelte';
	import { Icon } from '@steeze-ui/svelte-icon';
	import { DocumentText, Beaker, Clock, ServerStack } from '@steeze-ui/heroicons';
	import { page } from '$app/state';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const toeId = $derived(page.params.id);
</script>

<div>
	<SectionHeader
		title="Evidence List"
		description="All evidence collected for this target of evaluation."
	/>

	<div class="mt-6 grid grid-cols-3 gap-4">
		<div class="rounded-lg border border-gray-200 bg-white px-5 py-4">
			<div class="text-sm font-medium text-gray-500">Total</div>
			<div class="mt-1 text-2xl font-bold text-confirmate">{data.evidences.length}</div>
		</div>
	</div>

	<div class="mt-6 overflow-hidden rounded-lg border border-gray-200">
		<table class="min-w-full divide-y divide-gray-200">
			<thead class="bg-gray-50">
				<tr>
					<th scope="col" class="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500">
						Evidence ID
					</th>
					<th scope="col" class="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500">
						Resource
					</th>
					<th scope="col" class="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500">
						Tool
					</th>
					<th scope="col" class="px-4 py-3 text-left text-xs font-medium uppercase text-gray-500">
						Timestamp
					</th>
				</tr>
			</thead>
			<tbody class="divide-y divide-gray-200 bg-white">
				{#if data.evidences.length === 0}
					<tr>
						<td colspan="4" class="px-4 py-8 text-center text-sm text-gray-500">
							No evidence found.
						</td>
					</tr>
				{:else}
					{#each data.evidences as evidence}
						{@const resourceType = evidence.resource?.account ? 'Account' :
							evidence.resource?.application ? 'Application' :
							evidence.resource?.container ? 'Container' :
							evidence.resource?.virtualMachine ? 'Virtual Machine' :
							evidence.resource?.codeRepository ? 'Code Repository' :
							evidence.resource?.certificate ? 'Certificate' :
							evidence.resource?.secret ? 'Secret' :
							evidence.resource?.key ? 'Key' :
							evidence.resource?.identity ? 'Identity' :
							evidence.resource?.function ? 'Function' : 'Resource'}
						{@const resourceName = evidence.resource?.account?.name ??
							evidence.resource?.application?.name ??
							evidence.resource?.container?.name ??
							evidence.resource?.virtualMachine?.name ??
							evidence.resource?.codeRepository?.name ??
							evidence.resource?.certificate?.name ??
							evidence.resource?.secret?.name ??
							evidence.resource?.key?.name ??
							evidence.resource?.identity?.name ??
							evidence.resource?.function?.name ??
							'Unknown'}
						<tr class="hover:bg-gray-50">
							<td class="whitespace-nowrap px-4 py-3">
								<a
									href="/toe/{toeId}/evidences/{evidence.id}"
									class="flex items-center gap-2 text-sm font-mono text-gray-900 hover:text-confirmate"
								>
									<Icon src={DocumentText} class="h-4 w-4 text-gray-400" />
									{evidence.id}
								</a>
							</td>
							<td class="whitespace-nowrap px-4 py-3">
								<div class="flex items-center gap-2 text-sm text-gray-500">
									<Icon src={ServerStack} class="h-4 w-4 text-gray-400" />
									{resourceType}
								</div>
								<div class="text-xs text-gray-400">{resourceName}</div>
							</td>
							<td class="whitespace-nowrap px-4 py-3">
								<div class="flex items-center gap-2 text-sm text-gray-500">
									<Icon src={Beaker} class="h-4 w-4 text-gray-400" />
									{evidence.toolId ?? 'Unknown'}
								</div>
							</td>
							<td class="whitespace-nowrap px-4 py-3">
								<div class="flex items-center gap-2 text-sm text-gray-500">
									<Icon src={Clock} class="h-4 w-4 text-gray-400" />
									{evidence.timestamp ? new Date(evidence.timestamp).toLocaleString() : 'Unknown'}
								</div>
							</td>
						</tr>
					{/each}
				{/if}
			</tbody>
		</table>
	</div>
</div>