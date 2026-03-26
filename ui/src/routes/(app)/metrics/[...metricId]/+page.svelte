<script lang="ts">
	import { page } from '$app/stores';
	import BelowHeader from '$lib/components/BelowHeader.svelte';
	import Header from '$lib/components/Header.svelte';
	import { Icon } from '@steeze-ui/svelte-icon';
	import type { PageData } from './$types';
	import { Calculator, ListBullet, Trophy } from '@steeze-ui/heroicons';

	interface Props {
		data: PageData;
	}

	let { data }: Props = $props();
</script>

<Header
	name={data.metrics.get(data.metric.id)?.name ?? $page.params.metricId}
	description=""
	icon={false}
	buttons={false}
/>

<div>
	<h3 class="text-base leading-6 font-semibold text-gray-900">{data.metric.description}</h3>
	<dl class="mt-5 grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3">
		<div class="relative overflow-hidden rounded-lg bg-white px-4 pt-5 pb-4 shadow-sm">
			<dt>
				<div class="absolute rounded-md bg-blue-500 p-3">
					<Icon src={Calculator} class="h-6 w-6 text-white" />
				</div>
				<p class="ml-16 truncate text-sm font-medium text-gray-500">Operator</p>
			</dt>
			<dd class="ml-16 flex items-baseline">
				<p class="text-2xl font-semibold text-gray-900">{data.metricConfiguration.operator}</p>
			</dd>
		</div>
		<div class="relative overflow-hidden rounded-lg bg-white px-4 pt-5 pb-4 shadow-sm">
			<dt>
				<div class="absolute rounded-md bg-blue-500 p-3">
					<Icon src={Trophy} class="h-6 w-6 text-white" />
				</div>
				<p class="ml-16 truncate text-sm font-medium text-gray-500">Target Value</p>
			</dt>
			<dd class="ml-16 flex items-baseline">
				<p class="text-2xl font-semibold text-gray-900">
					{JSON.stringify(data.metricConfiguration.targetValue)}
				</p>
			</dd>
		</div>
		<div class="relative overflow-hidden rounded-lg bg-white px-4 pt-5 pb-4 shadow-sm">
			<dt>
				<div class="absolute rounded-md bg-blue-500 p-3">
					<Icon src={ListBullet} class="h-6 w-6 text-white" />
				</div>
				<p class="ml-16 truncate text-sm font-medium text-gray-500">Is Default</p>
			</dt>
			<dd class="ml-16 flex items-baseline">
				<p class="text-2xl font-semibold text-gray-900">{data.metricConfiguration.isDefault}</p>
			</dd>
		</div>
	</dl>
</div>
