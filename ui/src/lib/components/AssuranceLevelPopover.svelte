<script lang="ts" module>
	export interface AssuranceLevelEvent {
		catalog: Catalog;
		assuranceLevel: string;
	}
</script>

<script lang="ts">
	import type { Catalog } from '$lib/api/orchestrator';
	import { createPopover } from 'svelte-headlessui';
	import { createEventDispatcher, type Snippet } from 'svelte';
	import Transition from 'svelte-transition';

	const dispatch = createEventDispatcher<{
		select: AssuranceLevelEvent;
	}>();

	interface $$Events {
		select: CustomEvent<AssuranceLevelEvent>;
	}

	function select(level: string) {
		dispatch('select', {
			catalog: catalog,
			assuranceLevel: level
		});
	}

	interface Props {
		catalog: Catalog;
		children: Snippet;
	}

	let { catalog, children }: Props = $props();
	const popover = createPopover({});
</script>

<div class="z-20">
	<div class="h-full w-full" use:popover.button>
		{@render children()}
	</div>

	<Transition
		show={$popover.expanded}
		enter="transition ease-out duration-200"
		enterFrom="opacity-0 translate-y-1"
		enterTo="opacity-100 translate-y-0"
		leave="transition ease-in duration-150"
		leaveFrom="opacity-100 translate-y-0"
		leaveTo="opacity-0 translate-y-1"
	>
		<div use:popover.panel>
			<div
				class="w-screen max-w-sm flex-auto rounded-3xl bg-white p-4 text-sm leading-6 shadow-lg ring-1 ring-gray-900/5"
			>
				<div class="mt-4 mr-4 ml-4">
					<p class="font-medium text-gray-900">Please choose an assurance level</p>
					<p class="text-gray-500">
						The selected catalog requires the selection of an assurance level.
					</p>
				</div>

				{#each catalog.assuranceLevels as level}
					<div class="relative rounded-lg p-4 hover:bg-gray-50">
						<button onclick={() => select(level)} class="font-semibold text-gray-900">
							{level}
							<span class="absolute inset-0"></span>
						</button>
					</div>
				{/each}
			</div>
		</div>
	</Transition>
</div>
