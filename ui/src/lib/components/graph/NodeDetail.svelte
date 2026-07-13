<script lang="ts">
	import type { SchemaResourceSnapshot as Resource } from '$lib/api/openapi/evidence';
	import type { SchemaAssessmentResult as AssessmentResult } from '$lib/api/openapi/orchestrator';
	import { XMark, CheckCircle, XCircle, ShieldExclamation } from '@steeze-ui/heroicons';
	import { Icon } from '@steeze-ui/svelte-icon';

	interface Props {
		resource: Resource;
		results: AssessmentResult[];
		onclose: () => void;
		ondetail?: (resource: Resource) => void;
	}

	let { resource, results, onclose, ondetail }: Props = $props();

	let tab = $state<'results' | 'vulnerabilities'>('results');

	function shortName(id: string | undefined) {
		if (!id) return '—';
		const parts = id.split('/');
		return parts[parts.length - 1];
	}

	// Extract concrete resource data from the ontology union
	let concreteResource = $derived.by(() => {
		if (!resource.resource) return {};
		for (const val of Object.values(resource.resource)) {
			if (val && typeof val === 'object') return val as Record<string, unknown>;
		}
		return {};
	});

	let resourceType = $derived(resource.resourceType.split(',')[0].trim());
	let resourceName = $derived(
		(concreteResource.name as string) ?? shortName(resource.id)
	);

	// Extract vulnerabilities from the resource
	interface Vulnerability {
		cve?: string;
		criticality?: string;
		description?: string;
		cwe?: string[];
		url?: string;
	}
	let vulnerabilities = $derived.by(() => {
		const v = concreteResource.vulnerabilities;
		if (!Array.isArray(v)) return [];
		return (v as Vulnerability[]).filter((vuln) => vuln.cve);
	});
	let hasVulnerabilities = $derived(vulnerabilities.length > 0);

	// Auto-switch to vulnerabilities tab if there are any and no assessment results
	$effect(() => {
		if (hasVulnerabilities && results.length === 0 && tab === 'results') {
			tab = 'vulnerabilities';
		}
	});

	const CRITICALITY_COLORS: Record<string, string> = {
		critical: 'text-red-600 bg-red-50',
		high: 'text-red-600 bg-red-50',
		medium: 'text-amber-600 bg-amber-50',
		low: 'text-blue-600 bg-blue-50',
	};
</script>

<div class="flex h-full flex-col overflow-hidden rounded-xl border border-gray-200 bg-white shadow-sm">
	<!-- Header -->
	<div class="border-b border-gray-100 px-4 py-3">
		<div class="flex items-start justify-between">
			<div class="min-w-0">
				<p class="truncate text-sm font-semibold text-gray-900">{resourceName}</p>
				<p class="mt-0.5 text-xs text-gray-400">
					{resourceType}
					{#if results.length > 0}
						&middot; {results.length} assessment{results.length !== 1 ? 's' : ''}
					{/if}
					{#if hasVulnerabilities}
						&middot; <span class="text-red-500">{vulnerabilities.length} vulnerabilit{vulnerabilities.length !== 1 ? 'ies' : 'y'}</span>
					{/if}
				</p>
			</div>
			<div class="flex items-center gap-1">
				{#if ondetail}
					<button
						onclick={() => ondetail?.(resource)}
						class="rounded-md px-2 py-1 text-xs text-confirmate hover:bg-blue-50"
					>
						Details
					</button>
				{/if}
				<button onclick={onclose} class="text-gray-300 hover:text-gray-500">
					<Icon src={XMark} class="h-5 w-5" />
				</button>
			</div>
		</div>
	</div>

	<!-- Tabs -->
	<div class="border-b border-gray-100 px-4">
		<nav class="-mb-px flex space-x-4">
			{#if results.length > 0}
				<button
					onclick={() => (tab = 'results')}
					class="{tab === 'results'
						? 'border-confirmate text-confirmate'
						: 'border-transparent text-gray-400 hover:text-gray-600'} flex items-center gap-1.5 border-b-2 py-2.5 text-xs font-medium whitespace-nowrap"
				>
					Assessment
					<span class="bg-gray-100 text-gray-500 rounded-full px-1.5 py-0.5 text-xs">{results.length}</span>
				</button>
			{/if}
			{#if hasVulnerabilities}
				<button
					onclick={() => (tab = 'vulnerabilities')}
					class="{tab === 'vulnerabilities'
						? 'border-confirmate text-confirmate'
						: 'border-transparent text-gray-400 hover:text-gray-600'} flex items-center gap-1.5 border-b-2 py-2.5 text-xs font-medium whitespace-nowrap"
				>
					Vulnerabilities
					<span class="bg-red-50 text-red-600 rounded-full px-1.5 py-0.5 text-xs">{vulnerabilities.length}</span>
				</button>
			{/if}
		</nav>
	</div>

	<!-- Content -->
	<div class="flex-1 overflow-y-auto px-4 py-4">
		{#if tab === 'results'}
			{#if results.length === 0}
				<p class="text-sm text-gray-400">No assessment results for this resource.</p>
			{:else}
				<ul class="space-y-3">
					{#each results as r}
						<li class="flex items-start gap-x-3">
							{#if r.compliant}
								<Icon src={CheckCircle} class="mt-0.5 h-5 w-5 shrink-0 text-green-600" />
							{:else}
								<Icon src={XCircle} class="mt-0.5 h-5 w-5 shrink-0 text-red-600" />
							{/if}
							<div class="min-w-0">
								<a
									href="/toe/{resource.targetOfEvaluationId}/assessment-results/{r.id}"
									class="text-sm font-medium text-blue-600 hover:underline"
								>
									{r.metricId}
								</a>
							</div>
						</li>
					{/each}
				</ul>
			{/if}
		{:else if tab === 'vulnerabilities'}
			{#if vulnerabilities.length === 0}
				<p class="text-sm text-gray-400">No known vulnerabilities.</p>
			{:else}
				<ul class="space-y-3">
					{#each vulnerabilities as vuln}
						<li class="rounded-lg border border-gray-100 p-3">
							<div class="flex items-center gap-2">
								<Icon src={ShieldExclamation} class="h-4 w-4 shrink-0 text-red-500" />
								<span class="font-mono text-sm font-medium text-gray-900">{vuln.cve}</span>
								{#if vuln.criticality}
									<span class="ml-auto rounded-full px-2 py-0.5 text-xs font-medium {CRITICALITY_COLORS[vuln.criticality.toLowerCase()] ?? 'text-gray-600 bg-gray-50'}">
										{vuln.criticality}
									</span>
								{/if}
							</div>
							{#if vuln.description}
								<p class="mt-1.5 text-xs text-gray-500">{vuln.description}</p>
							{/if}
							{#if vuln.cwe && vuln.cwe.length > 0}
								<div class="mt-1.5 flex flex-wrap gap-1">
									{#each vuln.cwe as cwe}
										<span class="rounded bg-gray-100 px-1.5 py-0.5 font-mono text-xs text-gray-500">{cwe}</span>
									{/each}
								</div>
							{/if}
							{#if vuln.url}
								<a href={vuln.url} target="_blank" rel="noopener" class="mt-1.5 inline-block text-xs text-blue-600 hover:underline">
									More info
								</a>
							{/if}
						</li>
					{/each}
				</ul>
			{/if}
		{/if}
	</div>
</div>
