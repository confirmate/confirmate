<script lang="ts">
	interface Props {
		children?: import('svelte').Snippet;
	}

	let { children }: Props = $props();
</script>

<div class="mt-4 text-sm text-gray-700">
	{@render children?.()}
</div>
