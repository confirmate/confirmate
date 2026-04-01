<script lang="ts">
	let ready = $state(false);

	async function poll() {
		while (true) {
			try {
				await fetch('http://localhost:8000', { mode: 'no-cors' });
				ready = true;
				return;
			} catch {
				await new Promise((r) => setTimeout(r, 3000));
			}
		}
	}

	$effect(() => {
		poll();
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
		</div>
	</div>
{/if}
