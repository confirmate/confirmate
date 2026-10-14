<script lang="ts">
	import type { FrameworkProfile } from '../../profile';
	import { wizard } from '../../wizard.svelte';

	let { profile }: { profile: FrameworkProfile } = $props();

	const result = $derived(profile.choices.find((c) => c.id === wizard.choiceId)?.result);
</script>

{#if result}
	<div class="flex flex-col gap-3.5 rounded-card border border-line bg-surface p-6">
		<span class="text-[0.8125rem] uppercase tracking-wider text-ink-muted">Result</span>
		<span class="font-display text-heading font-(--cf-display-weight) leading-tight">{result.title}</span>
		<p class="text-ink-soft">{result.text}</p>
		{#if result.note}<p class="text-[0.8125rem] text-ink-muted">{result.note}</p>{/if}
	</div>
{/if}
