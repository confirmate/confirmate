<script lang="ts">
	let {
		value = 0,
		max = 100,
		label,
		showPercent = true
	}: {
		value?: number;
		max?: number;
		label?: string;
		showPercent?: boolean;
	} = $props();

	const percent = $derived(max > 0 ? Math.round((value / max) * 100) : 0);
	const isComplete = $derived(percent === 100);
</script>

<div class="w-full">
	{#if label || showPercent}
		<div class="mb-2 flex items-center justify-between">
			{#if label}
				<span class="text-sm font-medium text-gray-700">{label}</span>
			{/if}
			{#if showPercent}
				<span class="text-sm tabular-nums text-gray-500">{value} / {max} ({percent}%)</span>
			{/if}
		</div>
	{/if}
	<div class="h-2 w-full rounded-full bg-gray-100">
		<div
			class="h-2 rounded-full transition-all duration-500 ease-out {isComplete ? 'bg-emerald-500' : 'bg-confirmate'}"
			style="width: {percent}%"
		></div>
	</div>
</div>