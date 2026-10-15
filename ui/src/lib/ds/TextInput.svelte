<script lang="ts">
	let {
		id,
		value = $bindable(''),
		type = 'text',
		placeholder
	}: { id: string; value?: string; type?: string; placeholder?: string } = $props();
</script>

<input
	{id}
	{type}
	{placeholder}
	bind:value
	class="min-h-11 w-full rounded-control border border-line-strong bg-surface px-3.5 font-body text-ink focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-brand"
/>
