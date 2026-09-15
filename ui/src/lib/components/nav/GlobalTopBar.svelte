<script lang="ts">
	import { page } from '$app/state';
	import type { SchemaUser } from '$lib/api/openapi/orchestrator';
	import { logout } from '$lib/auth';
	import { roleLabel } from '$lib/roles';

	let { currentUser = null }: { currentUser?: SchemaUser | null } = $props();

	const links = [
		{ name: 'Dashboard', href: '/dashboard/' },
		{ name: 'Targets', href: '/toe/' },
		{ name: 'Metrics', href: '/metrics/' },
		{ name: 'Users', href: '/users/' }
	];

	const displayName = $derived(() => {
		if (!currentUser) return null;
		const full = [currentUser.firstName, currentUser.lastName].filter(Boolean).join(' ');
		return full || currentUser.username || currentUser.id;
	});

	const currentRoleLabel = $derived(() => {
		const first = currentUser?.roles?.find((r) => r !== 'ROLE_UNSPECIFIED');
		return first ? roleLabel(first) : null;
	});
</script>

<div class="border-b border-gray-200 bg-white">
	<div class="relative flex h-14 items-center px-6">
		<a href="/dashboard/" class="flex flex-col leading-tight">
			<div class="text-lg font-bold text-confirmate">confirmate</div>
			<div class="text-xs text-gray-400">Compliance &amp; Certification</div>
		</a>
		<nav class="absolute left-1/2 flex -translate-x-1/2 gap-1">
			{#each links as link}
				{@const active = page.url.pathname.startsWith(link.href)}
				<a
					href={link.href}
					class="{active
						? 'bg-blue-50 text-confirmate'
						: 'text-gray-600 hover:bg-gray-50 hover:text-gray-900'} rounded-md px-3 py-1.5 text-sm font-medium"
				>
					{link.name}
				</a>
			{/each}
		</nav>
		<div class="ml-auto flex items-center gap-3">
			{#if currentUser}
				<div class="flex items-center gap-2">
					<div class="flex h-7 w-7 items-center justify-center rounded-full bg-confirmate text-xs font-semibold text-white">
						{(currentUser.firstName?.[0] ?? currentUser.username?.[0] ?? '?').toUpperCase()}
					</div>
					<div class="flex flex-col leading-tight">
						<span class="text-sm font-medium text-gray-800">{displayName()}</span>
						{#if currentRoleLabel()}
							<span class="text-xs text-gray-400">{currentRoleLabel()}</span>
						{/if}
					</div>
				</div>
				<div class="h-4 w-px bg-gray-200"></div>
			{/if}
			<button onclick={() => logout()} class="text-sm text-gray-500 hover:text-gray-900">
				Sign out
			</button>
		</div>
	</div>
</div>
