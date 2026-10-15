<script lang="ts">
	import TodoView from '$lib/guided/components/TodoView.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<TodoView target={data.target!} />
