<script lang="ts">
	import { Chip, type Tone } from '$lib/ds';
	import type { RequirementStatus, Task } from '../source';

	let { status }: { status: RequirementStatus | Task['urgency'] | 'not-met' } = $props();

	const view: Record<string, { tone: Tone; label: string }> = {
		met: { tone: 'ok', label: 'Met' },
		attention: { tone: 'bad', label: 'Needs attention' },
		'not-met': { tone: 'bad', label: 'Not met' },
		urgent: { tone: 'bad', label: 'Due soon' },
		open: { tone: 'neutral', label: 'Open' }
	};
</script>

<Chip size="sm" tone={view[status].tone}>{view[status].label}</Chip>
