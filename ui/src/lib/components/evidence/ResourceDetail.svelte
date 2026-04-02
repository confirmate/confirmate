<script lang="ts">
	import type { SchemaResourceSnapshot as Resource, SchemaEvidence as Evidence } from '$lib/api/openapi/evidence';
	import type { SchemaAssessmentResult as AssessmentResult } from '$lib/api/openapi/orchestrator';
	import { evidenceClient } from '$lib/api/client';
	import EvidenceDetail from './EvidenceDetail.svelte';
	import { Icon } from '@steeze-ui/svelte-icon';
	import { XMark, CheckCircle, XCircle, MinusCircle, ChevronDown, ChevronUp } from '@steeze-ui/heroicons';

	interface Props {
		resource: Resource;
		results: AssessmentResult[];
		toeId: string;
		onclose: () => void;
	}
	let { resource, results, toeId, onclose }: Props = $props();

	// Derived assessment results for this resource
	let resourceResults = $derived(results.filter((r) => r.resourceId === resource.id));

	// Extract a field from the ontology union (Evidence.resource or ResourceSnapshot.resource)
	function getResourceProp(res: Evidence['resource'], prop: string): unknown {
		if (!res) return undefined;
		for (const val of Object.values(res)) {
			if (val && typeof val === 'object' && prop in val) {
				return (val as Record<string, unknown>)[prop];
			}
		}
		return undefined;
	}

	// Evidence loading (lazy, client-side filtered by resource.id)
	let evidences = $state<Evidence[]>([]);
	let loading = $state(false);
	let loadError = $state<string | null>(null);
	let nextPageToken = $state<string | undefined>(undefined);
	let hasMore = $state(false);

	let expandedId = $state<string | null>(null);

	// Re-load when resource changes
	$effect(() => {
		void resource.id; // track
		evidences = [];
		nextPageToken = undefined;
		hasMore = false;
		expandedId = null;
		loadEvidences();
	});

	async function loadEvidences() {
		loading = true;
		loadError = null;
		try {
			const res = await evidenceClient().GET('/v1/evidence_store/evidences', {
				params: {
					query: {
						'filter.targetOfEvaluationId': toeId,
						pageSize: 200,
						orderBy: 'timestamp',
						asc: false,
						...(nextPageToken ? { pageToken: nextPageToken } : {})
					}
				}
			});
			const page = (res.data?.evidences ?? []) as Evidence[];
			// Match by the id nested inside the evidence's ontology union resource
			const filtered = page.filter(
				(e) => getResourceProp(e.resource, 'id') === resource.id
			);
			evidences = [...evidences, ...filtered];
			nextPageToken = res.data?.nextPageToken ?? undefined;
			hasMore = !!nextPageToken;
		} catch (e) {
			loadError = String(e);
		} finally {
			loading = false;
		}
	}

	function formatRelative(iso: string): string {
		const diff = Date.now() - new Date(iso).getTime();
		const s = Math.floor(diff / 1000);
		if (s < 60) return `${s}s ago`;
		const m = Math.floor(s / 60);
		if (m < 60) return `${m}m ago`;
		const h = Math.floor(m / 60);
		if (h < 24) return `${h}h ago`;
		return `${Math.floor(h / 24)}d ago`;
	}

	let shortName = $derived(
		(getResourceProp(resource.resource, 'name') as string | undefined) ??
		resource.id.split('/').pop() ??
		resource.id
	);
	let resourceType = $derived(resource.resourceType.split(',')[0].trim());
</script>

<!-- Slide-over panel -->
<div class="flex h-full flex-col overflow-hidden border-l border-gray-200 bg-white shadow-xl">
	<!-- Header -->
	<div class="flex items-start justify-between gap-3 border-b border-gray-100 px-5 py-4">
		<div class="min-w-0">
			<p class="truncate font-semibold text-gray-900">{shortName}</p>
			<p class="mt-0.5 text-xs text-gray-400">{resourceType}</p>
			<p class="mt-0.5 max-w-xs truncate font-mono text-xs text-gray-300">{resource.id}</p>
		</div>
		<button onclick={onclose} class="shrink-0 text-gray-400 hover:text-gray-600">
			<Icon src={XMark} class="h-5 w-5" />
		</button>
	</div>

	<!-- Metrics strip -->
	{#if resourceResults.length > 0}
		<div class="border-b border-gray-100 px-5 py-3">
			<p class="mb-2 text-xs font-semibold tracking-wide text-gray-400 uppercase">Assessment</p>
			<div class="flex flex-wrap gap-2">
				{#each resourceResults as r}
					<span
						class="{r.compliant
							? 'bg-green-50 text-green-700 ring-green-200'
							: 'bg-red-50 text-red-700 ring-red-200'} inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium ring-1"
					>
						{#if r.compliant}
							<Icon src={CheckCircle} class="h-3 w-3" />
						{:else}
							<Icon src={XCircle} class="h-3 w-3" />
						{/if}
						{r.metricId}
					</span>
				{/each}
			</div>
		</div>
	{:else}
		<div class="border-b border-gray-100 px-5 py-3">
			<span class="inline-flex items-center gap-1 text-xs text-gray-400">
				<Icon src={MinusCircle} class="h-3.5 w-3.5" />
				No assessment results
			</span>
		</div>
	{/if}

	<!-- Evidence list -->
	<div class="flex-1 overflow-y-auto px-5 py-4">
		<div class="mb-3 flex items-center justify-between">
			<p class="text-xs font-semibold tracking-wide text-gray-400 uppercase">
				Evidences{evidences.length > 0 ? ` (${evidences.length}${hasMore ? '+' : ''})` : ''}
			</p>
		</div>

		{#if loading && evidences.length === 0}
			<p class="text-sm text-gray-400">Loading...</p>
		{:else if loadError}
			<p class="text-sm text-red-500">{loadError}</p>
		{:else if evidences.length === 0}
			<p class="text-sm text-gray-400">No evidences collected yet for this resource.</p>
		{:else}
			<ul class="space-y-1">
				{#each evidences as ev}
					{@const open = expandedId === ev.id}
					<li class="rounded-lg border border-gray-100 bg-gray-50">
						<button
							onclick={() => (expandedId = open ? null : (ev.id ?? null))}
							class="flex w-full items-center justify-between gap-3 px-3 py-2 text-left"
						>
							<div class="min-w-0">
								<p class="text-xs font-medium text-gray-700">{formatRelative(ev.timestamp ?? '')}</p>
								<p class="truncate font-mono text-xs text-gray-400">{ev.toolId}</p>
							</div>
							<Icon src={open ? ChevronUp : ChevronDown} class="h-4 w-4 shrink-0 text-gray-400" />
						</button>
						{#if open}
							<div class="border-t border-gray-100 px-3">
								<EvidenceDetail evidence={ev} {results} />
							</div>
						{/if}
					</li>
				{/each}
			</ul>

			{#if hasMore}
				<button
					onclick={loadEvidences}
					disabled={loading}
					class="mt-3 w-full rounded-lg border border-gray-200 py-2 text-sm text-gray-500 hover:bg-gray-50 disabled:opacity-40"
				>
					{loading ? 'Loading...' : 'Load more'}
				</button>
			{/if}
		{/if}
	</div>
</div>
