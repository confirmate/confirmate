<script lang="ts">
	import SectionHeader from '$lib/components/ui/SectionHeader.svelte';
</script>

<div>
	<SectionHeader
		title="Evidence"
		description="Raw evidence pipeline: resources → evidence → assessment results."
	>
		{#snippet actions()}
			<span class="inline-flex items-center rounded-full bg-amber-50 px-2.5 py-1 text-xs font-medium text-amber-700">
				Expert Mode
			</span>
		{/snippet}
	</SectionHeader>

	<div class="mt-8 rounded-lg border border-dashed border-gray-300 p-16 text-center">
		<p class="font-medium text-gray-400">Evidence explorer coming soon</p>
		<p class="mt-1 text-sm text-gray-400">
			This view will show a graph and table of resources, evidence, and assessment results for this target.
		</p>
	</div>
</div>
