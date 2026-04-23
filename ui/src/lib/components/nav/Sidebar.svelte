<script lang="ts">
	import { page } from '$app/state';
	import { ClipboardDocumentCheck, Cog6Tooth, ChartBar, Users, CommandLine, Beaker, ServerStack, ClipboardDocumentList, CheckCircle } from '@steeze-ui/heroicons';
	import { Icon } from '@steeze-ui/svelte-icon';
	import type { SchemaTargetOfEvaluation } from '$lib/api/openapi/orchestrator';

	let { services }: { services: SchemaTargetOfEvaluation[] } = $props();

	const toeId = $derived(page.params.id);
	const toe = $derived(services.find((s) => s.id === toeId));

	const navItems = $derived([
		{ name: 'Audit Scopes', href: `/toe/${toeId}/audit-scopes/`, icon: ClipboardDocumentCheck },
		{ name: 'Resources', href: `/toe/${toeId}/resources/`, icon: ServerStack },
		{ name: 'Access', href: `/toe/${toeId}/access/`, icon: Users },
		{ name: 'Documentation Generation', href: `/toe/${toeId}/documentation/`, icon: ClipboardDocumentList },
		{ name: 'Settings', href: `/toe/${toeId}/settings/`, icon: Cog6Tooth }
	]);

	const expertItems = $derived([
		{ name: 'Evidences', href: `/toe/${toeId}/evidences/`, icon: ChartBar },
		{ name: 'Assessment Results', href: `/toe/${toeId}/assessment-results/`, icon: Beaker },
		{ name: 'Evaluation Results', href: `/toe/${toeId}/evaluation-results/`, icon: CheckCircle }
	]);

	const integrationItems = $derived([
		{ name: 'Code Analysis', href: `/toe/${toeId}/code-analysis/`, icon: CommandLine }
	]);
</script>

<div class="flex w-64 shrink-0 flex-col border-r border-gray-200 bg-slate-50">
	<!-- ToE identity -->
	<div class="border-b border-gray-200 px-5 py-4">
		<div class="text-xs font-medium uppercase tracking-wide text-gray-400">
			Target of Evaluation
		</div>
		<div class="mt-1 truncate font-semibold text-gray-900">{toe?.name ?? toeId}</div>
		{#if toe?.description}
			<div class="mt-0.5 truncate text-xs text-gray-500">{toe.description}</div>
		{/if}
	</div>

	<!-- Primary nav -->
	<nav class="flex-1 px-3 py-4">
		<ul class="space-y-1">
			{#each navItems as item}
				{@const active = page.url.pathname.startsWith(item.href)}
				<li>
					<a
						href={item.href}
						class="{active
							? 'bg-gray-100 text-confirmate'
							: 'text-gray-700 hover:bg-gray-100 hover:text-confirmate'} flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium"
					>
						<Icon src={item.icon} class="h-5 w-5 shrink-0" />
						{item.name}
					</a>
				</li>
			{/each}
		</ul>

		<!-- Expert section -->
		<div class="mt-6">
			<div class="px-3 text-xs font-medium uppercase tracking-wide text-gray-400">Expert</div>
			<ul class="mt-2 space-y-1">
				{#each expertItems as item}
					{@const active = page.url.pathname.startsWith(item.href)}
					<li>
						<a
							href={item.href}
							class="{active
								? 'bg-gray-100 text-confirmate'
								: 'text-gray-500 hover:bg-gray-100 hover:text-gray-700'} flex items-center gap-3 rounded-md px-3 py-2 text-sm"
						>
							<Icon src={item.icon} class="h-4 w-4 shrink-0" />
							{item.name}
						</a>
					</li>
				{/each}
			</ul>
		</div>

		<!-- Integrations section -->
		<div class="mt-6">
			<div class="px-3 text-xs font-medium uppercase tracking-wide text-gray-400">Integrations</div>
			<ul class="mt-2 space-y-1">
				{#each integrationItems as item}
					{@const active = page.url.pathname.startsWith(item.href)}
					<li>
						<a
							href={item.href}
							class="{active
								? 'bg-gray-100 text-confirmate'
								: 'text-gray-500 hover:bg-gray-100 hover:text-gray-700'} flex items-center gap-3 rounded-md px-3 py-2 text-sm"
						>
							<Icon src={item.icon} class="h-4 w-4 shrink-0" />
							{item.name}
						</a>
					</li>
				{/each}
			</ul>
		</div>
	</nav>
</div>
