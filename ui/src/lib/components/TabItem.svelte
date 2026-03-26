<script module lang="ts">
	import { Icon, type IconSource } from '@steeze-ui/svelte-icon';
	import { page } from '$app/stores';

	export interface TabItemData {
		name: string;
		icon: IconSource;
		href: string;
		disabled?: boolean;
	}
</script>

<script lang="ts">
	interface Props {
		/**
		 * The data for this tab item.
		 */
		item: TabItemData;
		mobile?: boolean;
	}

	let { item, mobile = false }: Props = $props();

	// React to changes in the route and set the current status (true/false)
	let current = $derived($page.url.pathname.startsWith(item.href));
</script>

{#if mobile}
	<option selected={current} disabled={item.disabled}>{item.name}</option>
{:else if item.disabled}
	<span
		class="group inline-flex cursor-default items-center border-b-2 border-transparent px-1 py-4 text-sm font-medium text-gray-400"
	>
		<Icon src={item.icon} class="mr-2 -ml-0.5 h-5 w-5 text-gray-400" aria-hidden="true" />
		<span>{item.name}</span>
	</span>
{:else}
	<a
		href={item.href}
		class="{current
			? 'border-confirmate text-confirmate'
			: 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700'}
              group inline-flex items-center border-b-2 px-1 py-4 text-sm font-medium"
		aria-current={current ? 'page' : undefined}
	>
		<Icon
			src={item.icon}
			class="{current ? 'text-confirmate' : 'text-gray-400 group-hover:text-gray-500'}
                mr-2 -ml-0.5 h-5 w-5"
			aria-hidden="true"
		/>
		<span>{item.name}</span>
	</a>
{/if}
