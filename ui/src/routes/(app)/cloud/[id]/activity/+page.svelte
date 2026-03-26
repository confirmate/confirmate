<script lang="ts">
	import type { AssessmentResult } from '$lib/api/assessment';
	import ActivityItem, { type ActivityItemData } from '$lib/components/ActivityItem.svelte';
	import { Cloud, QueueList } from '@steeze-ui/heroicons';
	import type { PageData } from './$types';

	interface Props {
		data: PageData;
	}

	let { data }: Props = $props();

	function shortResourceId(id: string): string {
		let parts = id.split('/');

		return parts[parts.length - 1];
	}

	function buildTimeline(results: AssessmentResult[]): ActivityItemData[] {
		const timeline: ActivityItemData[] = [];
		let groupedResult = new Map<string, AssessmentResult[]>();
		// First, group resources together
		// TODO: group according to interval
		for (const result of results) {
			let group = groupedResult.get(result.resourceId);
			if (group == undefined) {
				group = [];
				groupedResult.set(result.resourceId, group);
			}

			group.push(result);
		}

		for (let group of groupedResult.values()) {
			const date = new Date(group[0].timestamp);

			timeline.push({
				content: `Assessed ${group.length} metrics for ${group[0].resourceTypes[0]}`,
				target: shortResourceId(group[0].resourceId),
				href: '/cloud/' + data.service.id + '/graph/?id=' + group[0].resourceId,
				date: formatDate(date),
				datetime: group[0].timestamp,
				icon: QueueList,
				iconBackground: 'bg-blue-500'
			});
		}

		const date = new Date(data.service.createdAt);

		timeline.push({
			content: 'Created Target of Evaluation',
			target: data.service.name,
			href: '/cloud/' + data.service.id,
			date: formatDate(date),
			datetime: data.service.createdAt,
			icon: Cloud,
			iconBackground: 'bg-gray-400'
		});

		return timeline;
	}

	function formatDate(date: Date): string {
		let formatter: Intl.DateTimeFormat;

		// Display time if its today
		if (isToday(date)) {
			formatter = new Intl.DateTimeFormat(undefined, { hour: '2-digit', minute: '2-digit' });
		} else {
			// Otherwise, display date
			formatter = new Intl.DateTimeFormat(undefined, { month: 'short', day: 'numeric' });
		}
		return formatter.format(date);
	}

	function isToday(date: Date): boolean {
		const today = new Date();

		return (
			date.getDate() == today.getDate() &&
			date.getMonth() == today.getMonth() &&
			date.getFullYear() == today.getFullYear()
		);
	}

	let timeline = $derived(buildTimeline(data.results));
</script>

<div class="flow-root">
	<ul class="-mb-8">
		{#each timeline as event, eventIdx}
			<ActivityItem last={eventIdx == timeline.length - 1} {event} />
		{/each}
	</ul>
</div>
