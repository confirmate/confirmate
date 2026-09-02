<script lang="ts">
	import type { SchemaEvidence as Evidence } from '$lib/api/openapi/evidence';
	import type { SchemaAssessmentResult as AssessmentResult } from '$lib/api/openapi/orchestrator';
	import { Icon } from '@steeze-ui/svelte-icon';
	import { CheckCircle, XCircle, ChevronDown, ChevronRight } from '@steeze-ui/heroicons';

	interface Props {
		evidence: Evidence;
		results: AssessmentResult[];
	}
	let { evidence, results }: Props = $props();

	// The evidence.resource is the ontology union — extract from first non-null field
	function getResourceProp(resource: Evidence['resource'], prop: string): unknown {
		if (!resource) return undefined;
		for (const val of Object.values(resource)) {
			if (val && typeof val === 'object' && prop in val) {
				return (val as Record<string, unknown>)[prop];
			}
		}
		return undefined;
	}

	let resourceId = $derived(getResourceProp(evidence.resource, 'id') as string | undefined);
	let resourceResults = $derived(results.filter((r) => r.resourceId === resourceId));

	function formatTimestamp(iso: string | undefined) {
		if (!iso) return '—';
		return new Intl.DateTimeFormat(navigator.language, {
			dateStyle: 'medium',
			timeStyle: 'medium'
		}).format(new Date(iso));
	}

	// --- Collapsible JSON tree ---

	interface TreeNode {
		key: string;
		value: unknown;
		depth: number;
		path: string;
		isObject: boolean;
	}

	function buildTree(obj: unknown, depth = 0, prefix = ''): TreeNode[] {
		if (typeof obj !== 'object' || obj === null) return [];
		const nodes: TreeNode[] = [];
		for (const [k, v] of Object.entries(obj as Record<string, unknown>)) {
			const path = prefix ? `${prefix}.${k}` : k;
			const isObj = typeof v === 'object' && v !== null && !Array.isArray(v);
			nodes.push({ key: k, value: v, depth, path, isObject: isObj });
			if (isObj) nodes.push(...buildTree(v, depth + 1, path));
		}
		return nodes;
	}

	let collapsed = $state(new Set<string>());

	function toggle(path: string) {
		const next = new Set(collapsed);
		next.has(path) ? next.delete(path) : next.add(path);
		collapsed = next;
	}

	function isVisible(node: TreeNode): boolean {
		const parts = node.path.split('.');
		for (let i = 1; i < parts.length; i++) {
			if (collapsed.has(parts.slice(0, i).join('.'))) return false;
		}
		return true;
	}

	// Build the property tree from the first populated field of the resource union
	let resourceData = $derived(() => {
		if (!evidence.resource) return {};
		for (const val of Object.values(evidence.resource)) {
			if (val && typeof val === 'object') return val as Record<string, unknown>;
		}
		return {};
	});

	let tree = $derived(buildTree(resourceData()));
</script>

<div class="space-y-4 py-2">
	<!-- Meta row -->
	<div class="flex items-start justify-between gap-4 text-xs text-gray-400">
		<span class="font-mono truncate">{evidence.id}</span>
		<span class="shrink-0">{formatTimestamp(evidence.timestamp)}</span>
	</div>
	<p class="text-xs text-gray-400">
		Tool: <span class="font-mono text-gray-600">{evidence.toolId ?? '—'}</span>
	</p>

	<!-- Properties tree -->
	<div>
		<p class="mb-1.5 text-xs font-semibold tracking-wide text-gray-400 uppercase">Properties</p>
		<div class="overflow-auto rounded-lg border border-gray-100 bg-gray-50 p-3 font-mono text-xs leading-5">
			{#if tree.length === 0}
				<span class="text-gray-400">—</span>
			{:else}
				{#each tree as node}
					{#if isVisible(node)}
						<div class="flex items-start gap-1" style="padding-left: {node.depth * 14}px">
							{#if node.isObject}
								<button onclick={() => toggle(node.path)} class="shrink-0 mt-0.5 text-gray-400 hover:text-gray-600">
									<Icon src={collapsed.has(node.path) ? ChevronRight : ChevronDown} class="h-3 w-3" />
								</button>
							{:else}
								<span class="w-3 shrink-0"></span>
							{/if}
							<span class="text-blue-700">{node.key}</span>
							<span class="text-gray-300 mx-0.5">:</span>
							{#if !node.isObject}
								<span class="text-gray-800 break-all">
									{Array.isArray(node.value) ? JSON.stringify(node.value) : String(node.value)}
								</span>
							{/if}
						</div>
					{/if}
				{/each}
			{/if}
		</div>
	</div>

	<!-- Assessment results -->
	{#if resourceResults.length > 0}
		<div>
			<p class="mb-1.5 text-xs font-semibold tracking-wide text-gray-400 uppercase">Assessment results</p>
			<ul class="space-y-1">
				{#each resourceResults as r}
					<li class="flex items-center gap-1.5 text-xs">
						{#if r.compliant}
							<Icon src={CheckCircle} class="h-3.5 w-3.5 shrink-0 text-green-600" />
						{:else}
							<Icon src={XCircle} class="h-3.5 w-3.5 shrink-0 text-red-600" />
						{/if}
						<span class="text-gray-700">{r.metricId}</span>
					</li>
				{/each}
			</ul>
		</div>
	{/if}
</div>
