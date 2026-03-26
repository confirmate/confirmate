<script lang="ts">
	import { AdjustmentsHorizontal, ArchiveBox, Folder } from '@steeze-ui/heroicons';
	import { Icon } from '@steeze-ui/svelte-icon';
	import type { PageData } from './$types';

	interface Props {
		data: PageData;
	}

	let { data }: Props = $props();

	let stats = $derived([
		{
			name: 'Targets of Evaluation',
			value: data.services.length,
			href: '/cloud',
			icon: Folder
		},
		{
			name: 'Available Catalogs',
			value: data.catalogs.length,
			href: '#',
			icon: ArchiveBox
		},
		{
			name: 'Available Metrics',
			value: data.metrics.size,
			href: '/metrics',
			icon: AdjustmentsHorizontal
		}
	]);
</script>

<div class="space-y-8">
	<div>
		<h1 class="text-2xl font-bold text-gray-900">Dashboard</h1>
		<p class="mt-1 text-sm text-gray-500">Overview of your cloud audit environment.</p>
	</div>

	<dl class="grid grid-cols-1 gap-5 sm:grid-cols-3">
		{#each stats as stat}
			<div class="overflow-hidden rounded-lg bg-white px-4 py-5 shadow sm:p-6">
				<dt class="flex items-center gap-x-2 truncate text-sm font-medium text-gray-500">
					<Icon src={stat.icon} class="h-5 w-5 shrink-0 text-gray-400" aria-hidden="true" />
					{stat.name}
				</dt>
				<dd class="mt-1 text-3xl font-semibold tracking-tight text-gray-900">
					{stat.value}
				</dd>
			</div>
		{/each}
	</dl>

	<div>
		<h2 class="text-base font-semibold text-gray-900">Targets of Evaluation</h2>
		{#if data.services.length === 0}
			<p class="mt-2 text-sm text-gray-500">
				No targets yet. <a href="/cloud/new" class="text-confirmate hover:underline">Create one</a> to get started.
			</p>
		{:else}
			<ul class="mt-4 divide-y divide-gray-100 rounded-lg border border-gray-200 bg-white">
				{#each data.services as service}
					<li class="flex items-center justify-between px-6 py-4">
						<div>
							<a
								href="/cloud/{service.id}/activity"
								class="text-sm font-medium text-gray-900 hover:text-confirmate"
							>
								{service.name}
							</a>
							{#if service.description}
								<p class="mt-0.5 text-xs text-gray-500">{service.description}</p>
							{/if}
						</div>
						<a
							href="/cloud/{service.id}/compliance"
							class="text-xs text-confirmate hover:underline"
						>
							View compliance
						</a>
					</li>
				{/each}
			</ul>
		{/if}
	</div>
</div>
