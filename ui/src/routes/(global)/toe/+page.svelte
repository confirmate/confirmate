<script lang="ts">
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<div>
	<div class="flex items-center justify-between">
		<h1 class="text-2xl font-bold text-gray-900">Targets of Evaluation</h1>
		<a
			href="/toe/new/"
			class="rounded-md bg-confirmate px-4 py-2 text-sm font-medium text-white hover:bg-confirmate-light"
		>
			New Target
		</a>
	</div>

	{#if data.services.length > 0}
		<ul class="mt-6 space-y-2">
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
