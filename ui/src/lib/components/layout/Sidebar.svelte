<!--
@component Sidebar

A navigation sidebar with hierarchical menu support.
Supports icons, disabled states, sub-navigation, and custom bottom content.
-->
<script lang="ts">
	import type { Snippet } from 'svelte';

	/**
	 * Navigation item structure for the sidebar
	 */
	interface NavigationItem {
		/** Display name of the navigation item */
		name: string;
		/** URL to navigate to */
		href: string;
		/** Optional icon component */
		icon?: any;
		/** Whether the item is disabled @default false */
		disabled?: boolean;
		/** Child navigation items for hierarchical menus */
		children?: NavigationItem[];
		/** Whether this is a sub-navigation item @default false */
		isSub?: boolean;
		/** Whether to highlight as a new item @default false */
		isNew?: boolean;
	}

	/**
	 * Props for the Sidebar component
	 */
	interface Props {
		/** Array of navigation items to display */
		navigation: NavigationItem[];
		/** Whether to use mobile layout @default false */
		mobile?: boolean;
		/** Optional content to display at the bottom of the sidebar */
		children?: Snippet;
	}

	let { navigation, mobile = false, children }: Props = $props();
</script>

<div
	class="{mobile
		? ''
		: 'border-r border-gray-200'} flex grow flex-col gap-y-5 overflow-y-auto bg-slate-50 px-6 pb-4"
>
	<div class="flex h-16 shrink-0 items-center border-b border-gray-200">
		<div class="flex items-center">
			<div class="ml-4 flex flex-col">
				<div class="text-xl font-semibold">Confirmate</div>
				<div class="text-sm text-gray-500">Compliance Assessment</div>
			</div>
		</div>
	</div>

	<nav class="flex flex-1 flex-col">
		<ul class="flex flex-1 flex-col gap-y-7">
			<li>
				<ul class="-mx-2 space-y-1">
					{#each navigation as item (item.href)}
						<li>
							{#if item.disabled}
								<span
									class="-mx-2 flex cursor-not-allowed gap-x-3 rounded-md p-2 text-sm leading-6 font-semibold text-gray-400"
								>
									{#if item.icon}
										{@const Icon = item.icon}
										<Icon class="h-6 w-6 shrink-0" aria-hidden="true" />
									{/if}
									{item.name}
								</span>
							{:else}
								<a
									href={item.href}
									class="group -mx-2 flex gap-x-3 rounded-md p-2 text-sm leading-6 font-semibold text-gray-700 hover:bg-gray-50 hover:text-blue-600"
								>
									{#if item.icon}
										{@const Icon = item.icon}
										<Icon
											class="h-6 w-6 shrink-0 text-gray-400 group-hover:text-blue-600"
											aria-hidden="true"
										/>
									{/if}
									{item.name}
								</a>
							{/if}

							{#if item.children && item.children.length > 0}
								<ul class="mt-2 ml-6 space-y-1">
									{#each item.children as child (child.href)}
										<li>
											<a
												href={child.href}
												class="{child.isNew
													? 'text-blue-600'
													: 'text-gray-600'} block rounded-md px-2 py-1 text-sm hover:bg-gray-50 hover:text-blue-600"
											>
												{child.name}
											</a>
										</li>
									{/each}
								</ul>
							{/if}
						</li>
					{/each}
				</ul>
			</li>

			{#if children}
				<li class="mt-auto">
					{@render children()}
				</li>
			{/if}
		</ul>
	</nav>
</div>
