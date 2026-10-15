<script lang="ts">
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import type { SchemaUser } from '$lib/api/openapi/orchestrator';
	import { logout } from '$lib/auth';
	import { Select } from '$lib/ds';
	import { invalidateAll } from '$app/navigation';
	import { themes, theme } from '$lib/theme/theme.svelte';
	import { sourceMode } from '../mode.svelte';

	let { currentUser = null }: { currentUser?: SchemaUser | null } = $props();

	const links = [
		{ name: 'Targets of Evaluation', href: resolve('/guided/'), prefix: resolve('/guided/') },
		{ name: 'People', href: resolve('/users/'), prefix: resolve('/users/') },
		{ name: 'Advanced', href: resolve('/dashboard/'), prefix: resolve('/dashboard/') }
	];

	const displayName = $derived(
		currentUser
			? [currentUser.firstName, currentUser.lastName].filter(Boolean).join(' ') ||
					currentUser.username ||
					currentUser.id
			: null
	);
</script>

<header class="flex min-h-16 flex-wrap items-center gap-x-8 gap-y-2 border-b border-line bg-surface px-10 font-body">
	<a href={resolve('/guided/')} class="flex items-baseline gap-2.5 no-underline">
		<span class="font-display text-2xl font-(--cf-display-weight) text-brand">confirmate</span>
		<span class="text-[0.8125rem] text-ink-muted">Compliance</span>
	</a>
	<nav class="flex gap-2">
		{#each links as link (link.name)}
			{@const active = page.url.pathname.startsWith(link.prefix)}
			<a
				href={link.href}
				aria-current={active ? 'page' : undefined}
				class="rounded-control px-3.5 py-2 no-underline {active
					? 'bg-brand-soft font-medium text-brand-ink'
					: 'text-ink-soft hover:bg-sunken'}"
			>
				{link.name}
			</a>
		{/each}
	</nav>
	<div class="ml-auto flex items-center gap-4">
		<span class="flex items-center gap-2 text-sm text-ink-muted">
			<label for="topbar-data">Data</label>
			<Select
				id="topbar-data"
				size="sm"
				value={sourceMode.current}
				options={[
					{ value: 'real', label: 'Real' },
					{ value: 'sample', label: 'Sample' }
				]}
				onchange={async (v) => {
					sourceMode.set(v as 'real' | 'sample');
					await invalidateAll();
				}}
			/>
		</span>
		<span class="flex items-center gap-2 text-sm text-ink-muted">
			<label for="topbar-theme">Theme</label>
			<Select
				id="topbar-theme"
				size="sm"
				value={theme.current}
				options={themes.map((t) => ({ value: t.id, label: t.label }))}
				onchange={(v) => theme.set(v as (typeof themes)[number]['id'])}
			/>
		</span>
		{#if displayName}<span class="text-sm text-ink-soft">{displayName}</span>{/if}
		<button class="cursor-pointer text-sm text-ink-muted hover:text-ink" onclick={() => logout()}>Sign out</button>
	</div>
</header>
