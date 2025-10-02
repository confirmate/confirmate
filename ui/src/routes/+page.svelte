<script lang="ts">
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<h1>Welcome to SvelteKit</h1>
<p>Visit <a href="https://svelte.dev/docs/kit">svelte.dev/docs/kit</a> to read the documentation</p>

TOES:

{#each data.toes as toe}
	<div>
		<h2>{toe.name}</h2>
		<p>{toe.description}</p>
	</div>
{/each}
