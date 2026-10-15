<script lang="ts">
	import type { Snippet } from 'svelte';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import type { TargetDetail } from '../source';

	let { target, children }: { target: TargetDetail; children: Snippet } = $props();

	const base = $derived(`/guided/${target.id}/`);
	const links = $derived([
		{ name: 'Overview', href: resolve(`/guided/${target.id}/`), exact: true },
		{ name: 'To do', href: resolve(`/guided/${target.id}/todo/`), exact: false, count: target.tasks.length },
		{ name: 'Requirements', href: resolve(`/guided/${target.id}/requirements/`), exact: false },
		{ name: 'Accepted risks', href: resolve(`/guided/${target.id}/risks/`), exact: false }
	]);
</script>

<div class="flex min-h-[calc(100vh-4rem)] flex-wrap">
	<aside class="box-border flex w-62 shrink-0 flex-col gap-6 border-r border-line bg-surface px-5 py-8">
		<div class="flex flex-col gap-0.5 px-3.5">
			<span class="text-xs uppercase tracking-wider text-ink-muted">ToE, {target.typeLabel.toLowerCase()}, {target.frameworkLabel.split(',')[0]}</span>
			<strong class="font-display text-xl font-(--cf-display-weight)">{target.name}</strong>
		</div>
		<nav class="flex flex-col gap-0.5" aria-label="{target.name} sections">
			{#each links as link (link.name)}
				{@const active = link.exact ? page.url.pathname === link.href : page.url.pathname.startsWith(link.href)}
				<a
					href={link.href}
					aria-current={active ? 'page' : undefined}
					class="flex items-center justify-between rounded-control px-3.5 py-2.5 no-underline {active
						? 'bg-brand-soft font-medium text-brand-ink'
						: 'text-ink-soft hover:bg-sunken'}"
				>
					{link.name}
					{#if link.count}<span class="rounded-full bg-bad px-2 text-xs font-semibold text-on-brand">{link.count}</span>{/if}
				</a>
			{/each}
		</nav>
		<div class="border-t border-line pt-4">
			<p class="px-3.5 pb-1.5 text-xs uppercase tracking-wider text-ink-muted">For experts</p>
			<a
				href={resolve(`/toe/${target.id}/audit-scopes/`)}
				class="block rounded-control px-3.5 py-2.5 text-ink-soft no-underline hover:bg-sunken"
			>
				Evidence and results
			</a>
		</div>
	</aside>

	<div class="box-border min-w-0 flex-[999_1_40rem] px-12 py-10" data-base={base}>
		{@render children()}
	</div>
</div>
