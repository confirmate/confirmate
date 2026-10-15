<script lang="ts">
	let {
		label,
		options,
		value = $bindable(),
		onchange
	}: {
		label: string;
		options: { value: string; label: string }[];
		value?: string;
		onchange?: (value: string) => void;
	} = $props();
</script>

<div role="group" aria-label={label} class="flex overflow-hidden rounded-control border border-line-strong">
	{#each options as o (o.value)}
		<button
			type="button"
			aria-pressed={value === o.value}
			onclick={() => {
				value = o.value;
				onchange?.(o.value);
			}}
			class="min-h-11 cursor-pointer px-4.5 font-body {value === o.value
				? 'bg-brand font-medium text-on-brand'
				: 'bg-surface text-ink hover:bg-sunken'}"
		>
			{o.label}
		</button>
	{/each}
</div>
