<script lang="ts">
	let {
		label,
		options,
		value = $bindable()
	}: { label: string; options: string[]; value?: string } = $props();
</script>

<div role="group" aria-label={label} class="flex flex-wrap gap-2">
	{#each options as option (option)}
		<button
			type="button"
			aria-pressed={value === option}
			onclick={() => (value = option)}
			class="h-9 cursor-pointer rounded-pill px-4 font-body font-medium {value === option
				? 'bg-brand text-on-brand'
				: 'border border-line-strong bg-surface text-ink hover:bg-sunken'}"
		>
			{option}
		</button>
	{/each}
</div>
