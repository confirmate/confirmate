<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { updateTargetOfEvaluation } from '$lib/api/orchestrator';
	import type { PageData } from './$types';

	interface Props {
		data: PageData;
	}

	let { data }: Props = $props();

	let name = $state(data.service.name);
	let description = $state(data.service.description ?? '');
	let saving = $state(false);
	let saved = $state(false);

	async function save() {
		saving = true;
		saved = false;
		await updateTargetOfEvaluation({ ...data.service, name, description });
		await invalidateAll();
		saving = false;
		saved = true;
	}
</script>

<div class="max-w-2xl space-y-8">
	<div>
		<h2 class="text-base font-semibold text-gray-900">General</h2>
		<p class="mt-1 text-sm text-gray-500">Update the name and description of this target.</p>

		<div class="mt-6 space-y-4">
			<div>
				<label for="name" class="block text-sm font-medium text-gray-700">Name</label>
				<input
					id="name"
					type="text"
					bind:value={name}
					class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-confirmate focus:outline-none focus:ring-1 focus:ring-confirmate"
				/>
			</div>
			<div>
				<label for="description" class="block text-sm font-medium text-gray-700">Description</label>
				<textarea
					id="description"
					rows="3"
					bind:value={description}
					class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-confirmate focus:outline-none focus:ring-1 focus:ring-confirmate"
				></textarea>
			</div>
		</div>
	</div>

	<div class="flex items-center gap-x-3">
		<button
			onclick={save}
			disabled={saving}
			class="rounded-md bg-confirmate px-4 py-2 text-sm font-semibold text-white shadow-sm hover:bg-confirmate-light focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-confirmate disabled:opacity-50"
		>
			{saving ? 'Saving…' : 'Save'}
		</button>
		{#if saved}
			<span class="text-sm text-green-600">Saved.</span>
		{/if}
	</div>
</div>
