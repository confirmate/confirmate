<script lang="ts">
	import type { SchemaAuditTrailEvent, SchemaUser } from '$lib/api/openapi/orchestrator';

	let {
		events,
		users = [],
		controlById = {},
		auditScopeId,
		targetId
	}: {
		events: SchemaAuditTrailEvent[];
		users?: SchemaUser[];
		controlById?: Record<string, { shortName?: string }>;
		auditScopeId?: string;
		targetId?: string;
	} = $props();

	function userLabel(u: SchemaUser): string {
		const full = [u.firstName, u.lastName].filter(Boolean).join(' ');
		return full || u.username || u.email || u.id;
	}

	const userById = $derived(Object.fromEntries(users.map((u) => [u.id, userLabel(u)])));

	function formatDate(iso?: string) {
		if (!iso) return '';
		return new Date(iso).toLocaleString();
	}

	function actorLabel(actorId?: string) {
		if (!actorId) return 'System';
		return userById[actorId] ?? actorId;
	}

	// Derive the control short name from event data.
	function controlLabel(event: SchemaAuditTrailEvent): string | null {
		const data = event.eventData as Record<string, unknown> | undefined;
		if (!data) return null;
		// ControlScopingEvent carries controlId directly
		const controlId = data['controlId'] as string | undefined;
		if (controlId && controlById[controlId]?.shortName) {
			return controlById[controlId].shortName ?? null;
		}
		return null;
	}

	function eventSummary(event: SchemaAuditTrailEvent): string {
		const data = event.eventData as Record<string, unknown> | undefined;
		if (!data) return 'Event recorded';
		const type = (data['@type'] as string | undefined) ?? '';
		if (type.includes('ControlScopingEvent')) {
			return data['inScope'] ? 'Control brought into scope' : 'Control removed from scope';
		}
		if (type.includes('ControlInScopeTransitionEvent')) {
			const toState = (data['toState'] as string | undefined) ?? '';
			return `State → ${STATE_LABELS[toState] ?? toState}`;
		}
		if (type.includes('ControlInScopeAssigneeChangedEvent')) {
			const newId = data['newAssigneeId'] as string | undefined;
			const newName = newId ? (userById[newId] ?? newId) : null;
			return newName ? `Assigned to ${newName}` : 'Assignee removed';
		}
		if (type.includes('ControlInScopeDetailsChangedEvent')) {
			return 'Implementation details updated';
		}
		return type.split('.').pop() ?? 'Event recorded';
	}

	const STATE_LABELS: Record<string, string> = {
		CONTROL_IN_SCOPE_STATE_OPEN: 'Open',
		CONTROL_IN_SCOPE_STATE_IN_PROGRESS: 'In Progress',
		CONTROL_IN_SCOPE_STATE_IMPLEMENTED: 'Implemented',
		CONTROL_IN_SCOPE_STATE_READY_FOR_REVIEW: 'Ready for Review',
		CONTROL_IN_SCOPE_STATE_ACCEPTED: 'Accepted'
	};
</script>

<div class="mt-8">
	<h2 class="mb-3 text-base font-semibold text-gray-900">Audit Trail</h2>
	{#if events.length === 0}
		<p class="text-sm text-gray-400">No audit trail events yet.</p>
	{:else}
		<div class="overflow-hidden rounded-lg border border-gray-200 bg-white">
			<table class="min-w-full divide-y divide-gray-100 text-sm">
				<thead class="bg-gray-50 text-xs font-medium uppercase tracking-wide text-gray-500">
					<tr>
						<th class="px-4 py-2 text-left">Time</th>
						<th class="px-4 py-2 text-left">Actor</th>
						<th class="px-4 py-2 text-left">Event</th>
						{#if Object.keys(controlById).length > 0}
							<th class="px-4 py-2 text-left">Control</th>
						{/if}
						<th class="px-4 py-2 text-left">Comment</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100">
					{#each events as event}
						{@const ctrl = controlLabel(event)}
						<tr class="hover:bg-gray-50">
							<td class="whitespace-nowrap px-4 py-2 text-gray-500">{formatDate(event.createdAt)}</td>
							<td class="px-4 py-2 font-medium text-gray-700">{actorLabel(event.actorId)}</td>
							<td class="px-4 py-2 text-gray-700">{eventSummary(event)}</td>
						{#if Object.keys(controlById).length > 0}
							<td class="px-4 py-2">
								{#if ctrl}
									{#if auditScopeId && targetId}
										{@const controlId = (event.eventData as Record<string, unknown> | undefined)?.['controlId'] as string | undefined}
										{#if controlId}
											<a
												href="/toe/{targetId}/audit-scopes/{auditScopeId}/controls/{controlId}/"
												class="font-mono text-xs text-confirmate hover:underline"
											>{ctrl}</a>
										{:else}
											<span class="font-mono text-xs text-gray-500">{ctrl}</span>
										{/if}
									{:else}
										<span class="font-mono text-xs text-gray-500">{ctrl}</span>
									{/if}
								{/if}
							</td>
						{/if}
							<td class="px-4 py-2 text-gray-500">{event.comment ?? ''}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>
