<script lang="ts">
	import Button from '$lib/components/Button.svelte';

	interface Props {
		name: string;
		description: string;
		buttons?: boolean;
		icon?: string | false;
		onremove?: () => void;
	}

	let { name, description, buttons = true, icon = '', onremove }: Props = $props();
</script>

<div class="md:flex md:items-center md:justify-between md:space-x-5">
	<div class="flex items-start space-x-5">
		{#if icon !== false}
			<div class="shrink-0">
				<div class="relative">
					<img
						class="h-16 w-16 rounded-full"
						src="https://images.unsplash.com/photo-1499346030926-9a72daac6c63?ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D&auto=format&fit=crop&w=2070&q=80"
						alt=""
					/>

					<span class="absolute inset-0 rounded-full shadow-inner" aria-hidden="true"></span>
				</div>
			</div>
		{/if}
		<!--
      Use vertical padding to simulate center alignment when both lines of text are one line,
      but preserve the same layout if the text wraps without making the image jump around.
    -->
		<div class="pt-1.5">
			{#if name != ''}
				<h1 class="text-2xl font-bold text-gray-900">{name}</h1>
			{:else}
				<h1 class="text-2xl font-bold text-gray-400">Enter a service name</h1>
			{/if}
			<p class="text-sm font-medium text-gray-500">
				{description}
			</p>
		</div>
	</div>
	<div
		class="mt-6 flex flex-col-reverse justify-stretch space-y-4 space-y-reverse sm:flex-row-reverse sm:justify-end sm:space-y-0 sm:space-x-3 sm:space-x-reverse md:mt-0 md:flex-row md:space-x-3"
	>
		{#if buttons}
			<Button onclick={onremove} class="bg-red-800 hover:bg-red-700">Delete</Button>
			<Button>Save</Button>
		{/if}
	</div>
</div>
