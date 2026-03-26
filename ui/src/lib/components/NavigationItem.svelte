<script module lang="ts">
	import { Icon, type IconSource } from '@steeze-ui/svelte-icon';
	import { page } from '$app/stores';
	import NavigationItem from '$lib/components/NavigationItem.svelte';

	/**
	 * An interface that contains necessary data to display a navigation item.
	 */
	export interface NavigationItemData {
		name: string;
		href: string;
		icon?: IconSource;
		isSub?: boolean;
		isNew?: boolean;
		children?: NavigationItemData[];
		disabled?: boolean;
	}
</script>

<script lang="ts">
	interface Props {
		/**
		 * The data for this navigation item.
		 */
		item: NavigationItemData;
	}

	let { item }: Props = $props();

	// React to changes in the route and set the current status (true/false)
	let current = $derived($page.url.pathname.startsWith(item.href));
</script>

<li>
	{#if item.disabled}
		<span
			class="cursor-default text-gray-400
    {item.isSub
				? 'block rounded-md py-2 pr-2 pl-9 text-sm leading-6'
				: 'group flex gap-x-3 rounded-md p-2 text-sm leading-6 font-semibold'}
    {item.isNew ? 'border-t' : ''}
    "
		>
			{#if item.icon}
				<Icon src={item.icon} class="h-6  w-6 shrink-0 text-gray-400" aria-hidden="true" />
			{/if}
			{item.name}
		</span>
	{:else}
		<a
			href={item.href}
			class="
      {current
				? 'text-confirmate bg-gray-50'
				: 'hover:text-confirmate text-gray-700 hover:bg-gray-50'}
      {item.isSub
				? 'block rounded-md py-2 pr-2 pl-9 text-sm leading-6'
				: 'group flex gap-x-3 rounded-md p-2 text-sm leading-6 font-semibold'}
      {item.isNew ? 'border-t' : ''}
      "
		>
			{#if item.icon}
				<Icon
					src={item.icon}
					class="{current ? 'text-confirmate' : 'group-hover:text-confirmate text-gray-400'}
        h-6 w-6 shrink-0"
					aria-hidden="true"
				/>
			{/if}
			{item.name}
		</a>
	{/if}
	<ul class="mt-1 px-2">
		{#each item.children ?? [] as subItem}
			<NavigationItem item={subItem} />
		{/each}
	</ul>
</li>
