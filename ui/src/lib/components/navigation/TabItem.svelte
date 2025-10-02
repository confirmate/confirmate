<!--
@component TabItem

An individual tab item component for use within the Tabs component.
Automatically detects current page and applies appropriate styling.
-->
<script lang="ts" module>
	/**
	 * Data structure for a tab item
	 */
	export interface TabItemData {
		/** Display name of the tab */
		name: string;
		/** Optional icon component */
		icon?: any;
		/** URL to navigate to */
		href: string;
		/** Whether the tab is disabled @default false */
		disabled?: boolean;
		/** Whether the tab is currently active (auto-detected) */
		current?: boolean;
	}
</script>

<script lang="ts">
	import { page } from '$app/stores';

	/**
	 * Props for the TabItem component
	 */
	interface Props {
		/** Tab item data */
		item: TabItemData;
		/** Whether to render in mobile mode @default false */
		mobile?: boolean;
	}

	let { item, mobile = false }: Props = $props();

	// Reactive current status based on page URL
	let current = $derived($page.url.pathname.startsWith(item.href));
</script>

{#if mobile}
	<option selected={current} disabled={item.disabled}>
		{item.name}
	</option>
{:else if item.disabled}
	<span
		class="group inline-flex cursor-default items-center border-b-2 border-transparent px-1 py-4 text-sm font-medium text-gray-400"
	>
		{#if item.icon}
			{@const Icon = item.icon}
			<Icon class="mr-2 h-4 w-4" />
		{/if}
		{item.name}
	</span>
{:else}
	<a
		href={item.href}
		class="{current
			? 'border-blue-500 text-blue-600'
			: 'border-transparent text-gray-500 hover:border-gray-200 hover:text-gray-700'} group inline-flex items-center border-b-2 px-1 py-4 text-sm font-medium"
		aria-current={current ? 'page' : undefined}
	>
		{#if item.icon}
			{@const Icon = item.icon}
			<Icon
				class="{current ? 'text-blue-500' : 'text-gray-400 group-hover:text-gray-500'} mr-2 h-4 w-4"
			/>
		{/if}
		{item.name}
	</a>
{/if}
