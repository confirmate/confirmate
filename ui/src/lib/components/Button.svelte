<script lang="ts">
	import type { HTMLButtonAttributes } from 'svelte/elements';
	import type { Snippet } from 'svelte';

	interface Props extends HTMLButtonAttributes {
		class?: string | undefined | null;
		children?: Snippet;
	}

	let { class: clazz, children, disabled, onclick, ...rest }: Props = $props();
</script>

<button
	type="button"
	class="{disabled ? 'bg-gray-500' : 'bg-confirmate hover:bg-confirmate-light'}
    focus-visible:outline-confirmate rounded-md px-3 py-2 text-sm font-semibold text-white shadow-xs focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-solid
    {clazz ?? ''}
    "
	{disabled}
	{onclick}
	{...rest}
>
	{@render children?.()}
</button>
