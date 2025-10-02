<!--
@component Tabs

A responsive tab navigation component that adapts between desktop and mobile layouts.
Displays as a select dropdown on mobile and tab bar on desktop.
-->
<script lang="ts">
	import TabItem, { type TabItemData } from './TabItem.svelte';

	/**
	 * Props for the Tabs component
	 */
	interface Props {
		/** Array of tab items to render */
		items: TabItemData[];
	}

	let { items }: Props = $props();
</script>

<div class="mt-4">
	<div class="sm:hidden">
		<label for="tabs" class="sr-only">Select a tab</label>
		<select
			id="tabs"
			name="tabs"
			class="block w-full rounded-md border-gray-300 focus:border-blue-500 focus:ring-blue-500"
		>
			{#each items as item (item.href)}
				<TabItem {item} mobile={true} />
			{/each}
		</select>
	</div>
	<div class="hidden sm:block">
		<div class="border-b border-gray-200">
			<nav class="-mb-px flex space-x-8" aria-label="Tabs">
				{#each items as item (item.href)}
					<TabItem {item} />
				{/each}
			</nav>
		</div>
	</div>
</div>
