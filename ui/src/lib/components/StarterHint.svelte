<script lang="ts">
	import { Icon, type IconSource } from '@steeze-ui/svelte-icon';
	import type { Snippet } from 'svelte';

	interface Props {
		icon: IconSource;
		type: string;
		component: Snippet;
	}

	let { icon, type, component }: Props = $props();
</script>

<div class="mt-20 text-center">
	<Icon src={icon} theme="outline" class="h-12" />
	<h3 class="mt-2 text-sm font-semibold text-gray-900">No {type}</h3>
	<span class="mt-1 text-sm text-gray-500">
		Get started by launching a {@render component()}
	</span>
</div>
