<script lang="ts">
	import type { PageProps } from './$types';
	import {
		Button,
		Card,
		Header,
		BelowHeader,
		Tabs,
		Input,
		Textarea,
		Select,
		StarterHint,
		ActivityItem,
		type TabItemData,
		type ActivityItemData
	} from '$lib/components';

	let { data }: PageProps = $props();

	// Demo data
	let inputValue = $state('');
	let textareaValue = $state('');
	let selectValue = $state('');

	const tabs: TabItemData[] = [
		{ name: 'Overview', href: '/', current: true },
		{ name: 'Assessment', href: '/assessment' },
		{ name: 'Compliance', href: '/compliance' }
	];

	const selectOptions = [
		{ value: 'option1', label: 'Option 1' },
		{ value: 'option2', label: 'Option 2' },
		{ value: 'option3', label: 'Option 3' }
	];

	const activities: ActivityItemData[] = [
		{
			content: 'Assessment completed for',
			target: 'Production Environment',
			href: '/assessment/1',
			date: '2h ago',
			datetime: '2025-10-02T14:00:00Z',
			icon: null,
			iconBackground: 'bg-green-500'
		},
		{
			content: 'New compliance rule added to',
			target: 'Security Framework',
			href: '/compliance/1',
			date: '4h ago',
			datetime: '2025-10-02T12:00:00Z',
			icon: null,
			iconBackground: 'bg-blue-500'
		}
	];
</script>

<div class="min-h-screen bg-gray-50">
	<div class="mx-auto max-w-7xl px-4 py-8">
		<Header
			title="Confirmate Dashboard"
			description="Compliance assessment and monitoring platform"
			showActions={false}
		/>

		<BelowHeader>
			Welcome to the migrated Confirmate UI with modern Svelte 5 components. This page showcases the
			reorganized component library.
		</BelowHeader>

		<Tabs items={tabs} />

		<div class="mt-8 grid grid-cols-1 gap-6 lg:grid-cols-2">
			<!-- Component Showcase -->
			<Card title="Form Components" description="Modern form elements with built-in validation">
				<div class="space-y-4">
					<Input
						id="demo-input"
						label="Sample Input"
						placeholder="Enter some text..."
						bind:value={inputValue}
						help="This is a help message"
					/>

					<Textarea
						id="demo-textarea"
						label="Description"
						placeholder="Enter a description..."
						bind:value={textareaValue}
					/>

					<Select
						id="demo-select"
						label="Choose Option"
						options={selectOptions}
						bind:value={selectValue}
					/>
				</div>

				{#snippet actions()}
					<Button variant="secondary">Cancel</Button>
					<Button>Save</Button>
				{/snippet}
			</Card>

			<Card title="Button Variants" description="Various button styles and states">
				<div class="space-y-4">
					<div class="flex flex-wrap gap-2">
						<Button>Primary</Button>
						<Button variant="secondary">Secondary</Button>
						<Button variant="danger">Danger</Button>
						<Button variant="ghost">Ghost</Button>
					</div>
					<div class="flex flex-wrap gap-2">
						<Button size="sm">Small</Button>
						<Button size="md">Medium</Button>
						<Button size="lg">Large</Button>
					</div>
					<div class="flex flex-wrap gap-2">
						<Button disabled>Disabled</Button>
					</div>
				</div>
			</Card>
		</div>

		<!-- Targets of Evaluation -->
		{#if data.toes.length > 0}
			<Card
				title="Targets of Evaluation"
				description="{data.toes.length} target(s) configured"
				class="mt-8"
			>
				<div class="divide-y divide-gray-200">
					{#each data.toes as toe}
						<div class="py-4">
							<h3 class="text-lg font-medium text-gray-900">{toe.name}</h3>
							{#if toe.description}
								<p class="mt-1 text-sm text-gray-500">{toe.description}</p>
							{/if}
						</div>
					{/each}
				</div>
			</Card>
		{:else}
			<div class="mt-8">
				<StarterHint type="targets of evaluation" icon={null}>
					{#snippet component()}
						<Button>Create Target</Button>
					{/snippet}
				</StarterHint>
			</div>
		{/if}

		<!-- Activity Feed -->
		{#if activities.length > 0}
			<Card title="Recent Activity" class="mt-8">
				<ul class="space-y-6">
					{#each activities as activity, index (activity.datetime)}
						<ActivityItem event={activity} last={index === activities.length - 1} />
					{/each}
				</ul>
			</Card>
		{/if}
	</div>
</div>
