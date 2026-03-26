<script module lang="ts">
	import { Icon, type IconSource } from '@steeze-ui/svelte-icon';

	export interface ActivityItemData {
		content: string;
		target: string;
		href: string;
		date: string;
		datetime: string;
		icon: IconSource;
		iconBackground: string;
	}
</script>

<script lang="ts">
	interface Props {
		event: ActivityItemData;
		last?: boolean;
	}

	let { event, last = false }: Props = $props();
</script>

<li>
	<div class="relative pb-8">
		{#if !last}
			<span class="absolute top-4 left-4 -ml-px h-full w-0.5 bg-gray-200" aria-hidden="true"></span>
		{/if}
		<div class="relative flex space-x-3">
			<div>
				<span
					class="{event.iconBackground} flex h-8 w-8 items-center justify-center rounded-full ring-8 ring-white"
				>
					<Icon src={event.icon} class="h-5 w-5 text-white" aria-hidden="true" />
				</span>
			</div>
			<div class="flex min-w-0 flex-1 justify-between space-x-4 pt-1.5">
				<div>
					<p class="text-sm text-gray-500">
						{event.content}
						<a href={event.href} class="font-medium text-gray-900">{event.target}</a>
					</p>
				</div>
				<div class="text-right text-sm whitespace-nowrap text-gray-500">
					<time datetime={event.datetime}>{event.date}</time>
				</div>
			</div>
		</div>
	</div>
</li>
