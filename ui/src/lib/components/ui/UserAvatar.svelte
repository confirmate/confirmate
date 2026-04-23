<script lang="ts">
	import type { SchemaUser } from '$lib/api/openapi/orchestrator';

	let { user, size = 'md' }: { user: SchemaUser; size?: 'sm' | 'md' } = $props();

	const initials = $derived(
		user.firstName && user.lastName
			? `${user.firstName[0]}${user.lastName[0]}`
			: (user.username ?? user.email ?? '?')[0].toUpperCase()
	);

	const displayName = $derived(
		user.firstName && user.lastName
			? `${user.firstName} ${user.lastName}`
			: (user.username ?? user.email ?? user.id)
	);
</script>

<div class="flex items-center gap-2">
	<div
		class="{size === 'sm' ? 'h-6 w-6 text-xs' : 'h-8 w-8 text-sm'} flex shrink-0 items-center justify-center rounded-full bg-confirmate font-medium text-white"
	>
		{initials}
	</div>
	<div class="min-w-0">
		<div class="{size === 'sm' ? 'text-xs' : 'text-sm'} truncate font-medium text-gray-900">
			{displayName}
		</div>
		{#if size === 'md' && user.email}
			<div class="truncate text-xs text-gray-500">{user.email}</div>
		{/if}
	</div>
</div>
