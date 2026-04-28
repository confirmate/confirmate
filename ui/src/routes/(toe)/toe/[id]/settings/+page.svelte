<script lang="ts">
	import SectionHeader from '$lib/components/ui/SectionHeader.svelte';
import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<div>
	<SectionHeader title="Settings" description="Manage this target of evaluation." />

	<div class="mt-6 max-w-xl space-y-4">
		<div>
			<label for="toe-name" class="block text-sm font-medium text-gray-700">Name</label>
			<input
				id="toe-name"
				type="text"
				value={data.toe.name}
				disabled
				class="mt-2 block w-full rounded-md border border-gray-300 bg-gray-50 px-3 py-2 text-sm text-gray-500"
			/>
		</div>
		<div>
			<label for="toe-description" class="block text-sm font-medium text-gray-700">Description</label>
			<textarea
				id="toe-description"
				rows="3"
				disabled
				class="mt-2 block w-full rounded-md border border-gray-300 bg-gray-50 px-3 py-2 text-sm text-gray-500"
			>{data.toe.description ?? ''}</textarea>
		</div>
		<div class="rounded-md bg-amber-50 px-4 py-3 text-sm text-amber-700">
			Editing is not yet implemented.
		</div>
	</div>

	<div class="mt-10 border-t border-gray-200 pt-6">
		<h3 class="text-sm font-medium text-gray-900">Danger Zone</h3>
		<p class="mt-1 text-sm text-gray-500">Permanently delete this target and all its data.</p>
		<button
			disabled
			class="mt-3 cursor-not-allowed rounded-md bg-red-50 px-3 py-2 text-sm font-medium text-red-400"
		>
			Delete Target
		</button>
	</div>
</div>
