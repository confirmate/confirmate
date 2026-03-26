<script lang="ts">
	import TabItem, { type TabItemData } from './TabItem.svelte';

	interface Props {
		items: TabItemData[];
	}

	let { items }: Props = $props();
</script>

<div class="mt-4">
	<div class="sm:hidden">
		<label for="tabs" class="sr-only">Select a tab</label>
		<!-- Use an "onChange" listener to redirect the user to the selected tab URL. -->
		<select
			id="tabs"
			name="tabs"
			class="focus:border-confirmate focus:ring-confirmate block w-full rounded-md border-gray-300"
		>
			{#each items as item}
				<TabItem {item} mobile={true} />
			{/each}
		</select>
	</div>
	<div class="hidden sm:block">
		<div class="border-b border-gray-200">
			<nav class="-mb-px flex space-x-8" aria-label="Tabs">
				{#each items as item}
					<TabItem {item} />
				{/each}
			</nav>
		</div>
	</div>
</div>
