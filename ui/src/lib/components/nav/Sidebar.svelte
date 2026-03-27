<script lang="ts">
	import { AdjustmentsHorizontal, Folder, Home } from '@steeze-ui/heroicons';
	import { Icon } from '@steeze-ui/svelte-icon';
	import NavigationItem from './NavigationItem.svelte';
	import type { SchemaTargetOfEvaluation } from '$lib/api/openapi/orchestrator';

	interface Props {
		services: SchemaTargetOfEvaluation[];
		mobile?: boolean;
	}

	let { services, mobile = false }: Props = $props();

	let navigation = $derived([
		{ name: 'Dashboard', href: '/dashboard', icon: Home },
		{
			name: 'Targets of Evaluation',
			href: '/toe',
			icon: Folder,
			children: [
				...services.map((s) => ({ name: s.name ?? s.id ?? '', href: `/toe/${s.id}/` , isSub: true })),
				{ name: 'New...', href: '/toe/new', isSub: true, isNew: true }
			]
		},
		{ name: 'Metrics', href: '/metrics', icon: AdjustmentsHorizontal }
	]);
</script>

<div
	class="{mobile ? '' : 'border-r border-gray-200'} flex grow flex-col gap-y-5 overflow-y-auto bg-slate-50 px-6 pb-4"
>
	<div class="flex h-16 shrink-0 items-center border-b border-gray-200">
		<div class="flex flex-col">
			<div class="text-xl font-bold text-confirmate">confirmate</div>
			<div class="text-sm text-gray-500">Compliance &amp; Certification</div>
		</div>
	</div>
	<nav class="flex flex-1 flex-col">
		<ul class="flex flex-1 flex-col gap-y-7">
			<li>
				<ul class="-mx-2 space-y-1">
					{#each navigation as item}
						<NavigationItem {item} />
					{/each}
				</ul>
			</li>
		</ul>
	</nav>
</div>
