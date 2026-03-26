<script lang="ts">
	import type { ComplianceStatus } from '$lib/api/evaluation';
	import { createListbox } from 'svelte-headlessui';
	import { Transition } from 'svelte-transition';
	import { Check, ChevronDown } from '@steeze-ui/heroicons';
	import { Icon } from '@steeze-ui/svelte-icon';

	const publishingOptions = [
		{
			name: 'Compliant',
			description:
				'This evaluation result will be considered as compliant. It will be marked as manually evaluated.',
			bgColor: 'bg-green-800',
			hoverColor: 'hover:bg-green-900 focus:ring-green-800',
			divideColor: 'divide-green-900',
			status: 'EVALUATION_STATUS_COMPLIANT_MANUALLY'
		},
		{
			name: 'Non-Compliant',
			description:
				'This evaluation result will be considered as compliant. It will be marked as manually evaluated.',
			bgColor: 'bg-red-800',
			hoverColor: 'hover:bg-red-900 focus:ring-red-900',
			divideColor: 'divide-red-900',
			status: 'EVALUATION_STATUS_NOT_COMPLIANT_MANUALLY'
		}
	];

	interface Props {
		status?: ComplianceStatus;
	}

	let { status = $bindable('EVALUATION_STATUS_COMPLIANT_MANUALLY') }: Props = $props();

	const listbox = createListbox({
		label: 'Compliance Status',
		selected:
			status == 'EVALUATION_STATUS_COMPLIANT_MANUALLY' ? publishingOptions[0] : publishingOptions[1]
	});
</script>

<div class="mt-2">
	<div class="sr-only">Change published status</div>
	<div class="relative">
		<div class="inline-flex divide-x {$listbox.selected.divideColor} rounded-md shadow-xs">
			<div
				class="inline-flex items-center gap-x-1.5 rounded-l-md px-3 py-2 text-white shadow-xs {$listbox
					.selected.bgColor}"
			>
				<Icon src={Check} class="-ml-0.5 h-5 w-5" aria-hidden="true" />
				<p class="text-sm font-semibold">{$listbox.selected.name}</p>
			</div>
			<div use:listbox.button>
				class="inline-flex items-center rounded-l-none rounded-r-md {$listbox.selected.bgColor}
				{$listbox.selected.hoverColor} p-2 focus:outline-hidden focus:ring-2 focus:ring-offset-2 focus:ring-offset-gray-50"
				>
				<span class="sr-only">Change published status</span>
				<Icon src={ChevronDown} class="h-5 w-5 text-white" aria-hidden="true" />
			</div>
		</div>

		<Transition
			show={$listbox.expanded}
			leave="transition ease-in duration-100"
			leaveFrom="opacity-100"
			leaveTo="opacity-0"
		>
			<ul
				use:listbox.items
				class="ring-opacity-5 absolute right-0 z-10 mt-2 w-72 origin-top-right divide-y divide-gray-200 overflow-hidden rounded-md bg-white shadow-lg ring-1 ring-black focus:outline-hidden"
			>
				{#each publishingOptions as option (option.name)}
					{@const active = $listbox.active === option.status}
					{@const selected = $listbox.selected === option.status}
					<li
						use:listbox.item={{ value: option.status }}
						class="{active
							? 'bg-confirmate text-white'
							: 'text-gray-900'} cursor-default p-4 text-sm select-none"
					>
						<div class="flex flex-col">
							<div class="flex justify-between">
								<p class={selected ? 'font-semibold' : 'font-normal'}>{option.name}</p>
								{#if selected}
									<span class="active ? 'text-white' : 'text-confirmate'">
										<Icon src={Check} class="h-5 w-5" aria-hidden="true" />
									</span>
								{/if}
							</div>
							<p class="{active ? 'text-indigo-200' : 'text-gray-500'} mt-2">
								{option.description}
							</p>
						</div>
					</li>
				{/each}
			</ul>
		</Transition>
	</div>
</div>
