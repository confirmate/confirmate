<script lang="ts">
	import { resolve } from '$app/paths';
	import { Button } from '$lib/ds';
	import TargetShell from '$lib/guided/components/TargetShell.svelte';
	import type { LayoutProps } from './$types';

	let { children, data }: LayoutProps = $props();
</script>

{#if data.target}
	<TargetShell target={data.target}>{@render children()}</TargetShell>
{:else}
	<main class="mx-auto flex max-w-2xl flex-col items-start gap-4 px-10 py-20">
		<h1 class="font-display text-4xl">No details in the preview yet</h1>
		<p class="text-ink-muted">Only the sample product Pythonpass has an overview in the preview.</p>
		<Button href={resolve('/guided/')}>Back to the Targets of Evaluation</Button>
	</main>
{/if}
