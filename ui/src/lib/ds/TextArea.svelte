<script lang="ts">
	let {
		id,
		value = $bindable(''),
		rows = 4,
		placeholder
	}: { id: string; value?: string; rows?: number; placeholder?: string } = $props();
</script>

<textarea
	{id}
	{rows}
	{placeholder}
	bind:value
	class="w-full rounded-control border border-line-strong bg-surface px-3.5 py-2.5 font-body text-ink focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-brand"
></textarea>
