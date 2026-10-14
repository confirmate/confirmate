<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		open = $bindable(false),
		title,
		children,
		footer
	}: { open: boolean; title: string; children: Snippet; footer?: Snippet } = $props();

	let dialog: HTMLDialogElement | undefined = $state();

	// The native dialog element keeps the focus inside, closes on Escape and marks the rest of the page inert.
	$effect(() => {
		if (!dialog) return;
		if (open && !dialog.open) dialog.showModal();
		if (!open && dialog.open) dialog.close();
	});
</script>

<dialog
	bind:this={dialog}
	onclose={() => (open = false)}
	aria-labelledby="dialog-title"
	class="m-auto w-full max-w-lg rounded-card border border-line bg-surface p-0 font-body text-ink backdrop:bg-ink/50"
>
	<div class="flex flex-col gap-5 p-6">
		<h2 id="dialog-title" class="text-lg font-semibold">{title}</h2>
		{@render children()}
		{#if footer}<div class="flex flex-wrap justify-end gap-2.5">{@render footer()}</div>{/if}
	</div>
</dialog>
