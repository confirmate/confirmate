<script lang="ts">
	let {
		title,
		description,
		actionHref,
		actionLabel
	}: {
		title: string;
		description?: string;
		actionHref?: string;
		actionLabel?: string;
	} = $props();
</script>

<div class="rounded-lg border border-dashed border-gray-300 p-12 text-center">
	<p class="font-medium text-gray-900">{title}</p>
	{#if description}
		<p class="mt-1 text-sm text-gray-500">{description}</p>
	{/if}
	{#if actionHref && actionLabel}
		<a
			href={actionHref}
			class="mt-4 inline-block rounded-md bg-confirmate px-4 py-2 text-sm font-medium text-white hover:bg-confirmate-light"
		>
			{actionLabel}
		</a>
	{/if}
</div>
