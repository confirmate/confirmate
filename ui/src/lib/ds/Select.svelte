<script lang="ts">
	let {
		id,
		value = $bindable(''),
		options
	}: { id: string; value?: string; options: { value: string; label: string }[] } = $props();
</script>

<select
	{id}
	bind:value
	class="min-h-11 w-full rounded-control border border-line-strong bg-surface px-3.5 font-body text-ink focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-brand"
>
	{#each options as o (o.value)}
		<option value={o.value}>{o.label}</option>
	{/each}
</select>
