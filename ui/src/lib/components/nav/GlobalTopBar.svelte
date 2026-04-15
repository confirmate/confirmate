<script lang="ts">
	import { page } from '$app/state';

	const links = [
		{ name: 'Dashboard', href: '/dashboard/' },
		{ name: 'Targets', href: '/toe/' },
		{ name: 'Metrics', href: '/metrics/' },
		{ name: 'Users', href: '/users/' }
	];
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
	</div>
</div>
