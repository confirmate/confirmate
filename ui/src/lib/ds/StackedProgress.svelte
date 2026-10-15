<script lang="ts">
	import { solidTone, type Tone } from './tones';

	let {
		segments,
		total
	}: { segments: { value: number; tone: Tone; label: string }[]; total: number } = $props();
</script>

<div
	class="flex h-2.5 overflow-hidden rounded-pill bg-neutral-soft"
	role="img"
	aria-label={segments.map((s) => `${s.value} ${s.label}`).join(', ')}
>
	{#each segments as s (s.label)}
		<div class={solidTone[s.tone]} style:width="{total > 0 ? (s.value / total) * 100 : 0}%"></div>
	{/each}
</div>
