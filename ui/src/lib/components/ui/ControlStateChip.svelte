<script lang="ts">
	let { state }: { state?: string } = $props();

	const CONFIG: Record<string, { label: string; class: string }> = {
		CONTROL_IN_SCOPE_STATE_OPEN: { label: 'Open', class: 'bg-gray-100 text-gray-600' },
		CONTROL_IN_SCOPE_STATE_IN_PROGRESS: { label: 'In Progress', class: 'bg-blue-50 text-blue-700' },
		CONTROL_IN_SCOPE_STATE_IMPLEMENTED: { label: 'Implemented', class: 'bg-green-50 text-green-700' },
		CONTROL_IN_SCOPE_STATE_READY_FOR_REVIEW: { label: 'Ready for Review', class: 'bg-yellow-50 text-yellow-700' },
		CONTROL_IN_SCOPE_STATE_ACCEPTED: { label: 'Accepted', class: 'bg-purple-50 text-purple-700' }
	};

	const cfg = $derived(state ? (CONFIG[state] ?? { label: state, class: 'bg-gray-100 text-gray-600' }) : null);
</script>

{#if cfg}
	<span class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium {cfg.class}">
		{cfg.label}
	</span>
{/if}
