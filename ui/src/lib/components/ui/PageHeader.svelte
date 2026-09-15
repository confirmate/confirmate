<script lang="ts">
	let {
		label,
		title,
		description,
		onremove
	}: {
		label?: string;
		title: string;
		description?: string;
		onremove?: () => void;
	} = $props();
</script>

<div class="mb-8 flex items-start justify-between">
	<div>
		{#if label}
			<div class="text-sm text-gray-500">{label}</div>
		{/if}
		<h1 class="mt-1 text-2xl font-bold text-gray-900">{title}</h1>
		{#if description}
			<p class="mt-1 text-sm text-gray-500">{description}</p>
		{/if}
	</div>
	{#if onremove}
		<button
			onclick={onremove}
			class="rounded-md bg-red-50 px-3 py-2 text-sm font-medium text-red-700 hover:bg-red-100"
		>
			Remove
		</button>
	{/if}
</div>
