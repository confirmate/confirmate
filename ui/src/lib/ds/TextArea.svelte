<script lang="ts">
	let {
		id,
		value = $bindable(''),
		rows = 4,
		placeholder,
		oninput
	}: {
		id: string;
		value?: string;
		rows?: number;
		placeholder?: string;
		/** Called with the new text, for fields that are not bound. */
		oninput?: (value: string) => void;
	} = $props();
</script>

<textarea
	{id}
	{rows}
	{placeholder}
	bind:value
	oninput={() => oninput?.(value)}
	class="w-full rounded-control border border-line-strong bg-surface px-3.5 py-2.5 font-body text-ink focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-brand"
></textarea>
