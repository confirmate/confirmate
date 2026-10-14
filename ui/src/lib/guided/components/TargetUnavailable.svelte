<script lang="ts">
	import { resolve } from '$app/paths';
	import { Button, PageHeading } from '$lib/ds';
</script>

<main class="mx-auto flex max-w-2xl flex-col items-start gap-6 px-10 py-20">
	<PageHeading
		title="No details in the preview yet"
		intro="Only the sample product Pythonpass has an overview in the preview."
	/>
	<Button href={resolve('/guided/')}>Back to the Targets of Evaluation</Button>
</main>
