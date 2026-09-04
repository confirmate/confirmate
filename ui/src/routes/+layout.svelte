<script lang="ts">
	import '../app.css';
	import { navigating } from '$app/state';

	let { children } = $props();
</script>

{#if navigating}
	<div class="fixed inset-x-0 top-0 z-[100] h-0.5 bg-confirmate transition-opacity">
		<div class="h-full w-full animate-pulse bg-confirmate"></div>
	</div>
{/if}

{@render children()}
