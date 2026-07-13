<script lang="ts">
	import { page } from '$app/state';
	import { exchangeCode, getReturnTo, clearReturnTo } from '$lib/auth';
	import { onMount } from 'svelte';

	let error = $state<string | null>(null);

	onMount(async () => {
		const code = page.url.searchParams.get('code');
		const errorParam = page.url.searchParams.get('error');

		if (errorParam) {
			error = `Authorization denied: ${errorParam}`;
			return;
		}

		if (!code) {
			error = 'No authorization code received';
			return;
		}

		try {
			await exchangeCode(code);
			const returnTo = getReturnTo();
			clearReturnTo();
			window.location.replace(returnTo);
		} catch (e) {
			error = e instanceof Error ? e.message : String(e);
		}
	});
</script>

<div class="flex h-screen items-center justify-center">
	{#if error}
		<div class="text-center">
			<p class="text-red-600">{error}</p>
			<button
				onclick={() => window.location.href = '/dashboard/'}
				class="mt-4 rounded bg-blue-600 px-4 py-2 text-white hover:bg-blue-700"
			>
				Back to dashboard
			</button>
		</div>
	{:else}
		<p class="text-gray-500">Completing sign-in…</p>
	{/if}
</div>
