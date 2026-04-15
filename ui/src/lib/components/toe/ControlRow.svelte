<script lang="ts">
	import type { SchemaControl } from '$lib/api/openapi/orchestrator';
	import { ChevronDown } from '@steeze-ui/heroicons';
	import { Icon } from '@steeze-ui/svelte-icon';

	let { control, depth = 0 }: { control: SchemaControl; depth?: number } = $props();

	const hasChildren = $derived((control.controls?.length ?? 0) > 0);
	let open = $state(false);
</script>

<div class="{depth > 0 ? 'ml-5 border-l border-gray-100 pl-4' : ''}">
	<div class="flex items-start gap-3 py-2.5">
		<span class="mt-0.5 shrink-0 rounded bg-gray-100 px-1.5 py-0.5 font-mono text-xs text-gray-500">
			{control.id}
		</span>
		<div class="min-w-0 flex-1">
			<div class="text-sm font-medium text-gray-900">{control.name}</div>
			{#if control.description}
				<div class="mt-0.5 text-sm text-gray-500">{control.description}</div>
			{/if}
		</div>
		{#if hasChildren}
			<button
				type="button"
				onclick={() => (open = !open)}
				class="mt-0.5 flex shrink-0 items-center gap-1 rounded-md px-1.5 py-0.5 text-xs text-gray-400 hover:bg-gray-100 hover:text-gray-600"
			>
				<Icon
					src={ChevronDown}
					class="h-3.5 w-3.5 transition-transform {open ? '' : '-rotate-90'}"
				/>
				{control.controls!.length}
			</button>
		{/if}
	</div>

	{#if hasChildren && open}
		<div class="mb-1">
			{#each control.controls! as sub}
				<svelte:self control={sub} depth={depth + 1} />
			{/each}
		</div>
	{/if}
</div>
