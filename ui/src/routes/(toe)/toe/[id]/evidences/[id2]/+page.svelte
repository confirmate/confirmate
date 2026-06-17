<script lang="ts">
	import type { SchemaEvidence, SchemaResource } from '$lib/api/openapi/evidence';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import { Icon } from '@steeze-ui/svelte-icon';
	import {
		ArrowLeft,
		DocumentText,
		Clock,
		User,
		ServerStack,
		Beaker,
		ExclamationTriangle
	} from '@steeze-ui/heroicons';
	import { page } from '$app/state';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const evidence: SchemaEvidence | null = $derived(data.evidence);
	const notFound = $derived(data.notFound);
	const toeId = $derived(page.params.id);

	function getResourceType(resource: SchemaResource | undefined): string {
		if (!resource) return 'Unknown';
		if (resource.account) return 'Account';
		if (resource.application) return 'Application';
		if (resource.job) return 'Job';
		if (resource.workflow) return 'Workflow';
		if (resource.codeRepository) return 'Code Repository';
		if (resource.container) return 'Container';
		if (resource.function) return 'Function';
		if (resource.virtualMachine) return 'Virtual Machine';
		if (resource.containerOrchestration) return 'Container Orchestration';
		if (resource.containerRegistry) return 'Container Registry';
		if (resource.certificate) return 'Certificate';
		if (resource.key) return 'Key';
		if (resource.secret) return 'Secret';
		if (resource.identity) return 'Identity';
		if (resource.roleAssignment) return 'Role Assignment';
		if (resource.containerImage) return 'Container Image';
		if (resource.vmImage) return 'VM Image';
		if (resource.keyVault) return 'Key Vault';
		if (resource.networkSecurityGroup) return 'Network Security Group';
		if (resource.functionService) return 'Function Service';
		if (resource.loadBalancer) return 'Load Balancer';
		if (resource.loggingService) return 'Logging Service';
		if (resource.relationalDatabaseService) return 'Relational Database';
		if (resource.objectStorageService) return 'Object Storage';
		if (resource.virtualNetwork) return 'Virtual Network';
		if (resource.blockStorage) return 'Block Storage';
		return 'Resource';
	}

function getResourceName(resource: SchemaResource | undefined): string {
		if (!resource) return 'Unknown';
		if (resource.account) return resource.account.name ?? resource.account.id ?? 'Unknown';
		if (resource.application) return resource.application.name ?? resource.application.id ?? 'Unknown';
		if (resource.job) return resource.job.name ?? resource.job.id ?? 'Unknown';
		if (resource.workflow) return resource.workflow.name ?? resource.workflow.id ?? 'Unknown';
		if (resource.codeRepository) return resource.codeRepository.name ?? resource.codeRepository.id ?? 'Unknown';
		if (resource.container) return resource.container.name ?? resource.container.id ?? 'Unknown';
		if (resource.function) return resource.function.name ?? resource.function.id ?? 'Unknown';
		if (resource.virtualMachine) return resource.virtualMachine.name ?? resource.virtualMachine.id ?? 'Unknown';
		if (resource.containerOrchestration) return resource.containerOrchestration.name ?? resource.containerOrchestration.id ?? 'Unknown';
		if (resource.containerRegistry) return resource.containerRegistry.name ?? resource.containerRegistry.id ?? 'Unknown';
		if (resource.certificate) return resource.certificate.name ?? resource.certificate.id ?? 'Unknown';
		if (resource.key) return resource.key.name ?? resource.key.id ?? 'Unknown';
		if (resource.secret) return resource.secret.name ?? resource.secret.id ?? 'Unknown';
		if (resource.identity) return resource.identity.name ?? resource.identity.id ?? 'Unknown';
		if (resource.roleAssignment) return resource.roleAssignment.id ?? 'Unknown';
		if (resource.containerImage) return resource.containerImage.name ?? resource.containerImage.id ?? 'Unknown';
		if (resource.vmImage) return resource.vmImage.name ?? resource.vmImage.id ?? 'Unknown';
		if (resource.keyVault) return resource.keyVault.name ?? resource.keyVault.id ?? 'Unknown';
		if (resource.networkSecurityGroup) return resource.networkSecurityGroup.name ?? resource.networkSecurityGroup.id ?? 'Unknown';
		if (resource.functionService) return resource.functionService.name ?? resource.functionService.id ?? 'Unknown';
		if (resource.loadBalancer) return resource.loadBalancer.name ?? resource.loadBalancer.id ?? 'Unknown';
		if (resource.loggingService) return resource.loggingService.name ?? resource.loggingService.id ?? 'Unknown';
		if (resource.relationalDatabaseService) return resource.relationalDatabaseService.name ?? resource.relationalDatabaseService.id ?? 'Unknown';
		if (resource.objectStorageService) return resource.objectStorageService.name ?? resource.objectStorageService.id ?? 'Unknown';
		if (resource.virtualNetwork) return resource.virtualNetwork.name ?? resource.virtualNetwork.id ?? 'Unknown';
		if (resource.blockStorage) return resource.blockStorage.name ?? resource.blockStorage.id ?? 'Unknown';
		return 'Unknown';
	}

	function getResourceId(resource: SchemaResource | undefined): string {
		if (!resource) return 'Unknown';
		if (resource.account) return resource.account.id ?? '';
		if (resource.application) return resource.application.id ?? '';
		if (resource.job) return resource.job.id ?? '';
		if (resource.workflow) return resource.workflow.id ?? '';
		if (resource.codeRepository) return resource.codeRepository.id ?? '';
		if (resource.container) return resource.container.id ?? '';
		if (resource.function) return resource.function.id ?? '';
		if (resource.virtualMachine) return resource.virtualMachine.id ?? '';
		if (resource.containerOrchestration) return resource.containerOrchestration.id ?? '';
		if (resource.containerRegistry) return resource.containerRegistry.id ?? '';
		if (resource.certificate) return resource.certificate.id ?? '';
		if (resource.key) return resource.key.id ?? '';
		if (resource.secret) return resource.secret.id ?? '';
		if (resource.identity) return resource.identity.id ?? '';
		if (resource.roleAssignment) return resource.roleAssignment.id ?? '';
		if (resource.containerImage) return resource.containerImage.id ?? '';
		if (resource.vmImage) return resource.vmImage.id ?? '';
		if (resource.keyVault) return resource.keyVault.id ?? '';
		if (resource.networkSecurityGroup) return resource.networkSecurityGroup.id ?? '';
		if (resource.functionService) return resource.functionService.id ?? '';
		if (resource.loadBalancer) return resource.loadBalancer.id ?? '';
		if (resource.loggingService) return resource.loggingService.id ?? '';
		if (resource.relationalDatabaseService) return resource.relationalDatabaseService.id ?? '';
		if (resource.objectStorageService) return resource.objectStorageService.id ?? '';
		if (resource.virtualNetwork) return resource.virtualNetwork.id ?? '';
		if (resource.blockStorage) return resource.blockStorage.id ?? '';
		return '';
	}
</script>

<div>
	<div class="mb-6">
		<a
			href="/toe/{toeId}/evidences/"
			class="inline-flex items-center gap-1.5 text-sm text-gray-500 hover:text-confirmate"
		>
			<Icon src={ArrowLeft} class="h-4 w-4" />
			Back to Evidences
		</a>
	</div>

	{#if notFound}
		<div class="rounded-lg border border-amber-200 bg-amber-50 px-5 py-4">
			<div class="flex items-center gap-2 text-amber-800">
				<Icon src={ExclamationTriangle} class="h-5 w-5" />
				<span class="font-medium">Evidence Not Found</span>
			</div>
			<p class="mt-2 text-sm text-amber-700">
				The evidence with ID <code class="font-mono">{data.evidenceId}</code> was not found.
				This may happen when evidence has been deduplicated or removed.
			</p>
			<p class="mt-1 text-sm text-amber-600">
				The assessment result may have been updated to reference a newer evidence with the same content.
			</p>
		</div>
	{:else if evidence}
	<PageHeader
		title={evidence.id ?? 'Evidence'}
		description="Evidence details"
	/>

	<div class="mt-6 grid grid-cols-3 gap-4">
		<div class="rounded-lg border border-gray-200 bg-white px-5 py-4">
			<div class="flex items-center gap-2 text-sm font-medium text-gray-500">
				<Icon src={DocumentText} class="h-4 w-4" />
				Evidence ID
			</div>
			<div class="mt-2 font-mono text-sm text-gray-900">{evidence.id}</div>
		</div>

		<div class="rounded-lg border border-gray-200 bg-white px-5 py-4">
			<div class="flex items-center gap-2 text-sm font-medium text-gray-500">
				<Icon src={Beaker} class="h-4 w-4" />
				Tool
			</div>
			<div class="mt-2 text-sm text-gray-900">{evidence.toolId ?? 'Unknown'}</div>
		</div>

		<div class="rounded-lg border border-gray-200 bg-white px-5 py-4">
			<div class="flex items-center gap-2 text-sm font-medium text-gray-500">
				<Icon src={Clock} class="h-4 w-4" />
				Timestamp
			</div>
			<div class="mt-2 text-sm text-gray-900">
				{evidence.timestamp ? new Date(evidence.timestamp).toLocaleString() : 'Unknown'}
			</div>
		</div>
	</div>

	{#if evidence.resource}
		<div class="mt-6 rounded-lg border border-gray-200 bg-white">
			<div class="border-b border-gray-200 px-5 py-4">
				<div class="flex items-center gap-2 text-sm font-medium text-gray-500">
					<Icon src={ServerStack} class="h-4 w-4" />
					Resource
				</div>
			</div>
			<div class="px-5 py-4">
				<div class="grid grid-cols-2 gap-4">
					<div>
						<div class="text-xs font-medium uppercase text-gray-400">Type</div>
						<div class="mt-1 text-sm text-gray-900">{getResourceType(evidence.resource)}</div>
					</div>
					<div>
						<div class="text-xs font-medium uppercase text-gray-400">Name</div>
						<div class="mt-1 text-sm text-gray-900">{getResourceName(evidence.resource)}</div>
					</div>
					<div class="col-span-2">
						<div class="text-xs font-medium uppercase text-gray-400">ID</div>
						<div class="mt-1 font-mono text-sm text-gray-900">{getResourceId(evidence.resource)}</div>
					</div>
				</div>
			</div>
		</div>
	{/if}

	{#if evidence.targetOfEvaluationId}
		<div class="mt-6 rounded-lg border border-gray-200 bg-white px-5 py-4">
			<div class="flex items-center gap-2 text-sm font-medium text-gray-500">
				<Icon src={User} class="h-4 w-4" />
				Target of Evaluation
			</div>
			<div class="mt-2 text-sm text-gray-900">{evidence.targetOfEvaluationId}</div>
		</div>
	{/if}

	{#if evidence.experimentalRelatedResourceIds && evidence.experimentalRelatedResourceIds.length > 0}
		<div class="mt-6 rounded-lg border border-gray-200 bg-white">
			<div class="border-b border-gray-200 px-5 py-4">
				<div class="text-sm font-medium text-gray-500">Related Resource IDs</div>
			</div>
			<div class="px-5 py-4">
				<ul class="space-y-1">
					{#each evidence.experimentalRelatedResourceIds as relatedId}
						<li class="font-mono text-sm text-gray-600">{relatedId}</li>
					{/each}
				</ul>
			</div>
		</div>
	{/if}
	{/if}
</div>