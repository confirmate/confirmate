<!--
@component ActivityItem

A timeline activity item component for displaying activity feeds.
Shows an icon, content, target, and timestamp in a timeline format.
-->
<script lang="ts" module>
	/**
	 * Data structure for an activity item in the timeline
	 */
	export interface ActivityItemData {
		/** Main descriptive text for the activity */
		content: string;
		/** Target/subject of the activity */
		target: string;
		/** Link URL for the activity */
		href: string;
		/** Human-readable date string */
		date: string;
		/** ISO datetime string */
		datetime: string;
		/** Icon component to display */
		icon: any;
		/** Background color class for the icon (e.g., 'bg-green-500') */
		iconBackground: string;
	}
</script>

<script lang="ts">
	/**
	 * Props for the ActivityItem component
	 */
	interface Props {
		/** Activity event data */
		event: ActivityItemData;
		/** Whether this is the last item (removes connecting line) @default false */
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
					{#if event.icon}
						{@const Icon = event.icon}
						<Icon class="h-5 w-5 text-white" aria-hidden="true" />
					{/if}
				</span>
			</div>
			<div class="flex min-w-0 flex-1 justify-between space-x-4 pt-1.5">
				<div>
					<p class="text-sm text-gray-500">
						{event.content}
						<a href={event.href} class="font-medium text-gray-900">
							{event.target}
						</a>
					</p>
				</div>
				<div class="text-right text-sm whitespace-nowrap text-gray-500">
					<time datetime={event.datetime}>{event.date}</time>
				</div>
			</div>
		</div>
	</div>
</li>
