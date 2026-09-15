<script lang="ts">
	import type { SchemaResourceSnapshot as Resource } from '$lib/api/openapi/evidence';
	import { XMark, ShieldExclamation, ArrowTopRightOnSquare } from '@steeze-ui/heroicons';
	import { Icon } from '@steeze-ui/svelte-icon';

	interface Props {
		resource: Resource | null;
		onclose: () => void;
	}

	let { resource, onclose }: Props = $props();

	function shortName(id: string | undefined) {
		if (!id) return '—';
		const parts = id.split('/');
		return parts[parts.length - 1];
	}

	let concreteResource = $derived.by(() => {
		if (!resource?.resource) return {};
		for (const val of Object.values(resource.resource)) {
			if (val && typeof val === 'object') return val as Record<string, unknown>;
		}
		return {};
	});

	let resourceType = $derived(resource?.resourceType.split(',')[0].trim() ?? '');
	let resourceName = $derived(
		(concreteResource.name as string) ?? shortName(resource?.id)
	);

	interface Vulnerability {
		cve?: string;
		criticality?: string;
		description?: string;
		cwe?: string[];
		url?: string;
		exploitable?: boolean;
	}

	let vulnerabilities = $derived.by(() => {
		const v = concreteResource.vulnerabilities;
		if (!Array.isArray(v)) return [];
		return (v as Vulnerability[]).filter((vuln) => vuln.cve);
	});

	// Build a structured tree of properties (excluding vulnerabilities, which get a custom renderer)
	interface PropertyNode {
		key: string;
		value: string | PropertyNode[];
		isArray: boolean;
	}

	function buildTree(obj: Record<string, unknown>, prefix = ''): PropertyNode[] {
		const nodes: PropertyNode[] = [];
		for (const [k, v] of Object.entries(obj)) {
			if (k === 'raw' || k === 'labels' || k === 'vulnerabilities') continue;
			if (v === null || v === undefined) continue;
			const key = prefix ? `${prefix}.${k}` : k;

			if (Array.isArray(v)) {
				if (v.length === 0) continue;
				const children: PropertyNode[] = [];
				for (let i = 0; i < v.length; i++) {
					if (v[i] && typeof v[i] === 'object') {
						children.push(...buildTree(v[i] as Record<string, unknown>, `${key}[${i}]`));
					} else {
						children.push({ key: `${key}[${i}]`, value: String(v[i]), isArray: false });
					}
				}
				nodes.push({ key, value: children, isArray: true });
			} else if (typeof v === 'object') {
				const children = buildTree(v as Record<string, unknown>, key);
				if (children.length > 0) {
					nodes.push({ key, value: children, isArray: false });
				}
			} else {
				nodes.push({ key, value: String(v), isArray: false });
			}
		}
		return nodes;
	}

	let tree = $derived(buildTree(concreteResource));

	let expanded = $state<Set<string>>(new Set());
	function toggle(key: string) {
		const next = new Set(expanded);
		if (next.has(key)) next.delete(key);
		else next.add(key);
		expanded = next;
	}
	function isExpanded(key: string) {
		return expanded.has(key);
	}

	// Auto-expand top-level object nodes when the resource changes.
	// Uses a sentinel to track the previous resource id so we only
	// re-expand when a new resource is opened, not on every state change.
	let lastResourceId = $state<string | null>(null);
	$effect(() => {
		const id = resource?.id ?? null;
		if (id !== lastResourceId && tree.length > 0) {
			lastResourceId = id;
			expanded = new Set(tree.filter((n) => Array.isArray(n.value)).map((n) => n.key));
		}
	});

	const CRITICALITY_COLORS: Record<string, string> = {
		critical: 'text-red-700 bg-red-100 border-red-200',
		high: 'text-red-600 bg-red-50 border-red-200',
		medium: 'text-amber-700 bg-amber-50 border-amber-200',
		low: 'text-blue-700 bg-blue-50 border-blue-200',
	};
</script>

{#if resource}
	<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" onclick={onclose}>
		<div class="relative z-10 flex max-h-[80vh] w-full max-w-2xl flex-col overflow-hidden rounded-lg bg-white shadow-xl" onclick={(e) => e.stopPropagation()}>
			<!-- Header -->
			<div class="flex items-center justify-between border-b border-gray-100 px-6 py-4">
				<div class="min-w-0">
					<h3 class="text-base font-semibold text-gray-900">{resourceName}</h3>
					<p class="mt-0.5 text-xs text-gray-400">{resourceType} &middot; {resource.id}</p>
				</div>
				<button onclick={onclose} class="ml-4 text-gray-300 hover:text-gray-500">
					<Icon src={XMark} class="h-5 w-5" />
				</button>
			</div>

			<!-- Content -->
			<div class="flex-1 overflow-y-auto px-6 py-4">
				<!-- Properties tree -->
				{#if tree.length === 0 && vulnerabilities.length === 0}
					<p class="text-sm text-gray-400">No properties available.</p>
				{:else}
					{#if tree.length > 0}
						{#if vulnerabilities.length > 0}
							<h4 class="mb-3 text-sm font-semibold text-gray-900">Properties</h4>
						{/if}
						{#snippet treeNode(nodes: PropertyNode[], depth: number)}
							{#each nodes as node}
								{#if Array.isArray(node.value)}
									<div style="padding-left: {depth * 1.5}rem">
										<button
											type="button"
											onclick={() => toggle(node.key)}
											class="flex w-full items-center gap-1.5 py-1 text-left text-xs font-medium text-gray-600 hover:text-gray-900"
										>
											<span class="text-gray-400">{isExpanded(node.key) ? '▼' : '▶'}</span>
											<span>{node.key}</span>
											<span class="text-gray-300">({(node.value as PropertyNode[]).length} items)</span>
										</button>
										{#if isExpanded(node.key)}
											{@render treeNode(node.value as PropertyNode[], depth + 1)}
										{/if}
									</div>
								{:else}
									<div style="padding-left: {depth * 1.5}rem" class="flex items-baseline gap-2 py-0.5">
										<span class="text-xs font-medium text-gray-400">{node.key}</span>
										<span class="text-xs text-gray-700 break-all">{node.value}</span>
									</div>
								{/if}
							{/each}
						{/snippet}
						{@render treeNode(tree, 0)}
					{/if}
				{/if}

				<!-- Vulnerabilities (custom renderer) -->
				{#if vulnerabilities.length > 0}
					<div class="mt-6">
						<h4 class="mb-3 flex items-center gap-1.5 text-sm font-semibold text-gray-900">
							<Icon src={ShieldExclamation} class="h-4 w-4 text-red-500" />
							Vulnerabilities
							<span class="rounded-full bg-red-50 px-2 py-0.5 text-xs font-medium text-red-600">{vulnerabilities.length}</span>
						</h4>
						<div class="space-y-3">
							{#each vulnerabilities as vuln}
								<div class="rounded-lg border border-gray-200 p-4">
									<div class="flex items-center gap-2">
										<a href={vuln.url} target="_blank" rel="noopener" class="font-mono text-sm font-semibold text-blue-600 hover:underline">
											{vuln.cve}
										</a>
										{#if vuln.criticality}
											<span class="ml-auto rounded-full border px-2.5 py-0.5 text-xs font-medium {CRITICALITY_COLORS[vuln.criticality.toLowerCase()] ?? 'text-gray-600 bg-gray-50 border-gray-200'}">
												{vuln.criticality}
											</span>
										{/if}
										{#if vuln.exploitable === true}
											<span class="rounded-full border border-red-300 bg-red-100 px-2.5 py-0.5 text-xs font-medium text-red-700">
												Exploitable
											</span>
										{/if}
									</div>
									{#if vuln.description}
										<p class="mt-2 text-sm text-gray-600">{vuln.description}</p>
									{/if}
									{#if vuln.cwe && vuln.cwe.length > 0}
										<div class="mt-2 flex flex-wrap gap-1.5">
											{#each vuln.cwe as cwe}
												<span class="rounded bg-gray-100 px-2 py-0.5 font-mono text-xs text-gray-500">{cwe}</span>
											{/each}
										</div>
									{/if}
									{#if vuln.url}
										<a href={vuln.url} target="_blank" rel="noopener" class="mt-2 inline-flex items-center gap-1 text-xs text-blue-600 hover:underline">
											<Icon src={ArrowTopRightOnSquare} class="h-3 w-3" />
											Advisory
										</a>
									{/if}
								</div>
							{/each}
						</div>
					</div>
				{/if}
			</div>
		</div>
	</div>
{/if}
