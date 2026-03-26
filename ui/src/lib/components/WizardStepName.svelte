<script lang="ts">
	import type { WizardData } from '$lib/components/Wizard.svelte';

	interface Props {
		data: WizardData;
	}

	let { data = $bindable() }: Props = $props();
</script>

<div
	class="focus-within:ring-confirmate rounded-md px-3 pt-2.5 pb-1.5 shadow-xs ring-1 ring-gray-300 ring-inset focus-within:ring-2"
>
	<label for="name" class="block text-xs font-medium text-gray-900">Name</label>
	<input
		type="text"
		name="name"
		id="name"
		class="block w-full border-0 p-0 text-gray-900 placeholder:text-gray-400 focus:ring-0 sm:text-sm sm:leading-6"
		placeholder="Enter a service name"
		bind:value={data.service.name}
	/>
</div>
