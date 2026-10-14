<script lang="ts">
	let {
		label,
		hint,
		selected = false,
		onselect
	}: { label: string; hint?: string; selected?: boolean; onselect: () => void } = $props();
</script>

<button
	type="button"
	role="radio"
	aria-checked={selected}
	onclick={onselect}
	class="flex min-h-14 w-full cursor-pointer items-center gap-3.5 rounded-card px-4.5 py-3 text-left font-body text-ink {selected
		? 'border-2 border-brand bg-brand-soft'
		: 'border border-line-strong bg-surface hover:bg-sunken'}"
>
	<span
		class="box-border size-5 shrink-0 rounded-full bg-surface {selected
			? 'border-[6px] border-brand'
			: 'border-2 border-ink-muted'}"
	></span>
	<span class="flex flex-col">
		<strong class="font-medium">{label}</strong>
		{#if hint}<span class="text-[0.8125rem] text-ink-muted">{hint}</span>{/if}
	</span>
</button>
