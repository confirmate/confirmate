<script lang="ts">
	import { browser } from '$app/environment';

	let ready = $state(false);
	let elapsed = $state(0);

	async function poll() {
		const startTime = Date.now();
		while (true) {
			try {
				await fetch('http://localhost:8000', { mode: 'no-cors' });
				ready = true;
				return;
			} catch {
				elapsed = Math.floor((Date.now() - startTime) / 1000);
				await new Promise((r) => setTimeout(r, 3000));
			}
		}
	}

	$effect(() => {
		if (browser) poll();
	});
</script>

{#if ready}
	<iframe src="http://localhost:8000" class="-m-8 block h-[calc(100vh-4rem)] w-full border-0" title="Code Analysis"></iframe>
{:else}
	<div class="-m-8 flex h-[calc(100vh-4rem)] items-center justify-center bg-slate-50">
		<div class="text-center">
			<div class="mb-3 text-2xl animate-spin inline-block">⟳</div>
			<p class="font-medium text-gray-600">Waiting for code analysis to complete...</p>
			<p class="mt-1 text-sm text-gray-400">The dashboard starts on port 8000 once the analysis finishes.</p>
			{#if elapsed > 120}
				<p class="mt-3 text-sm text-amber-600">This is taking longer than expected ({Math.floor(elapsed / 60)} min). Check the code-analysis logs if needed.</p>
			{/if}
		</div>
	</div>
{/if}
