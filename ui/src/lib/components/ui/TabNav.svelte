<script lang="ts">
	import { page } from '$app/state';
	import { Icon } from '@steeze-ui/svelte-icon';
	import type { ComponentType } from 'svelte';

	let {
		tabs
	}: {
		tabs: { name: string; href: string; icon?: ComponentType }[];
	} = $props();
</script>

<div class="mb-6 border-b border-gray-200">
	<nav class="-mb-px flex space-x-8">
		{#each tabs as tab}
			{@const active = page.url.pathname.startsWith(tab.href)}
			<a
				href={tab.href}
				class="{active
					? 'border-confirmate text-confirmate'
					: 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700'} flex items-center gap-2 border-b-2 px-1 py-4 text-sm font-medium"
			>
				{#if tab.icon}
					<Icon src={tab.icon} class="h-4 w-4" />
				{/if}
				{tab.name}
			</a>
		{/each}
	</nav>
</div>
