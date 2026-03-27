<script lang="ts">
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<div>
	<h1 class="text-2xl font-bold text-gray-900">Dashboard</h1>
	<div class="mt-6 grid grid-cols-1 gap-6 sm:grid-cols-3">
		<div class="rounded-lg border border-gray-200 bg-white p-6">
			<div class="text-sm font-medium text-gray-500">Targets of Evaluation</div>
			<div class="mt-2 text-3xl font-bold text-confirmate">{data.services.length}</div>
		</div>
		<div class="rounded-lg border border-gray-200 bg-white p-6">
			<div class="text-sm font-medium text-gray-500">Catalogs</div>
			<div class="mt-2 text-3xl font-bold text-confirmate">{data.catalogs.length}</div>
		</div>
		<div class="rounded-lg border border-gray-200 bg-white p-6">
			<div class="text-sm font-medium text-gray-500">Metrics</div>
			<div class="mt-2 text-3xl font-bold text-confirmate">{data.metrics.size}</div>
		</div>
	</div>

	{#if data.services.length > 0}
		<div class="mt-8">
			<h2 class="text-lg font-semibold text-gray-900">Targets of Evaluation</h2>
			<ul class="mt-4 space-y-2">
				{#each data.services as service}
					<li>
						<a
							href="/toe/{service.id}/audit-scopes/"
							class="block rounded-lg border border-gray-200 bg-white p-4 hover:border-confirmate hover:bg-blue-50"
						>
							<div class="font-medium text-gray-900">{service.name}</div>
							{#if service.description}
								<div class="mt-1 text-sm text-gray-500">{service.description}</div>
							{/if}
						</a>
					</li>
				{/each}
			</ul>
		</div>
	{:else}
		<div class="mt-8 rounded-lg border border-dashed border-gray-300 p-12 text-center">
			<p class="text-gray-500">No targets of evaluation yet.</p>
			<a
				href="/toe/new/"
				class="mt-4 inline-block rounded-md bg-confirmate px-4 py-2 text-sm font-medium text-white hover:bg-confirmate-light"
			>
				Add your first target
			</a>
		</div>
	{/if}
</div>
