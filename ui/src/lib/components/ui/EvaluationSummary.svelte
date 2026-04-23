<script lang="ts">
	interface Props {
		results: { status: string; controlId: string }[];
	}

	let { results }: Props = $props();

	const counts = $derived(() => {
		const c = { compliant: 0, notCompliant: 0, pending: 0, manual: 0 };
		for (const r of results) {
			if (r.status === 'EVALUATION_STATUS_COMPLIANT') c.compliant++;
			else if (r.status === 'EVALUATION_STATUS_NOT_COMPLIANT') c.notCompliant++;
			else if (r.status === 'EVALUATION_STATUS_PENDING') c.pending++;
			else if (r.status.includes('MANUALLY')) c.manual++;
		}
		return c;
	});

	const total = $derived(counts().compliant + counts().notCompliant + counts().pending + counts().manual);

	const colors = $derived({
		compliant: '#10b981',
		notCompliant: '#ef4444',
		pending: '#f59e0b',
		manual: '#6b7280'
	});

	const getPercent = (n: number) => total > 0 ? (n / total) * 100 : 0;
</script>

{#if total > 0}
	<div class="mb-6 flex items-center gap-6 rounded-lg border border-gray-200 bg-white px-5 py-4">
		<div class="relative h-20 w-20 shrink-0">
			<svg viewBox="0 0 36 36" class="h-20 w-20 -rotate-90">
				<circle cx="18" cy="18" r="15.9" fill="none" stroke="#e5e7eb" stroke-width="3" />
				{#if counts().compliant > 0}
					<circle
						cx="18" cy="18" r="15.9"
						fill="none"
						stroke={colors.compliant}
						stroke-width="3"
						stroke-dasharray="{getPercent(counts().compliant)}, 100"
					/>
				{/if}
				{#if counts().notCompliant > 0}
					<circle
						cx="18" cy="18" r="15.9"
						fill="none"
						stroke={colors.notCompliant}
						stroke-width="3"
						stroke-dasharray="{getPercent(counts().notCompliant)}, 100"
						stroke-dashoffset="-{getPercent(counts().compliant)}"
					/>
				{/if}
				{#if counts().pending > 0}
					<circle
						cx="18" cy="18" r="15.9"
						fill="none"
						stroke={colors.pending}
						stroke-width="3"
						stroke-dasharray="{getPercent(counts().pending)}, 100"
						stroke-dashoffset="-{getPercent(counts().compliant + counts().notCompliant)}"
					/>
				{/if}
				{#if counts().manual > 0}
					<circle
						cx="18" cy="18" r="15.9"
						fill="none"
						stroke={colors.manual}
						stroke-width="3"
						stroke-dasharray="{getPercent(counts().manual)}, 100"
						stroke-dashoffset="-{getPercent(counts().compliant + counts().notCompliant + counts().pending)}"
					/>
				{/if}
			</svg>
		</div>

		<div class="flex flex-wrap gap-4">
			<div class="flex items-center gap-2">
				<span class="h-3 w-3 rounded-full bg-emerald-500"></span>
				<span class="text-sm text-gray-600">Compliant ({counts().compliant})</span>
			</div>
			<div class="flex items-center gap-2">
				<span class="h-3 w-3 rounded-full bg-red-500"></span>
				<span class="text-sm text-gray-600">Not Compliant ({counts().notCompliant})</span>
			</div>
			<div class="flex items-center gap-2">
				<span class="h-3 w-3 rounded-full bg-amber-500"></span>
				<span class="text-sm text-gray-600">Pending ({counts().pending})</span>
			</div>
			<div class="flex items-center gap-2">
				<span class="h-3 w-3 rounded-full bg-gray-500"></span>
				<span class="text-sm text-gray-600">Manual ({counts().manual})</span>
			</div>
		</div>
	</div>
{/if}