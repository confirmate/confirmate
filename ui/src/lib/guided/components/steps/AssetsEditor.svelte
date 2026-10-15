<script lang="ts">
	import { Button, Select, TextInput } from '$lib/ds';
	import type { FrameworkProfile } from '../../profile';
	import { impactLevels } from '../../risk';
	import { wizard } from '../../wizard.svelte';

	let { profile }: { profile: FrameworkProfile } = $props();

	const properties = [
		{ key: 'confidentiality', label: 'Disclosure', hint: 'Confidentiality is lost' },
		{ key: 'integrity', label: 'Manipulation', hint: 'Integrity is lost' },
		{ key: 'availability', label: 'Outage', hint: 'Availability is lost' }
	] as const;

	function add() {
		const used = new Set(wizard.risk.assets.map((a) => a.name));
		const name = profile.risk.assetSuggestions.find((s) => !used.has(s)) ?? 'New asset';
		wizard.risk.assets.push({ name, confidentiality: 3, integrity: 3, availability: 2 });
		wizard.approved = false;
	}

	function remove(index: number) {
		wizard.risk.assets.splice(index, 1);
		wizard.approved = false;
	}
</script>

<section class="flex flex-col gap-4 rounded-card border border-line bg-surface p-(--cf-card-pad)" aria-labelledby="assets-heading">
	<div class="flex flex-col gap-1">
		<h2 id="assets-heading" class="text-[1.0625rem] font-semibold">What does the product protect?</h2>
		<p class="text-sm text-ink-muted">
			List the data and functions that an attacker would be after. For each one, say how bad it is if it is disclosed,
			manipulated or not available: from 1, very low, to 5, very high. Requirements that protect more than that do not
			apply to your product.
		</p>
	</div>

	{#each wizard.risk.assets as asset, i (i)}
		<div class="flex flex-col gap-3 rounded-control border border-line p-4">
			<div class="flex items-end gap-3">
				<div class="flex min-w-0 flex-1 flex-col gap-1.5">
					<label for="asset-name-{i}" class="text-sm font-medium">Asset</label>
					<TextInput id="asset-name-{i}" bind:value={asset.name} />
				</div>
				<Button variant="quiet" size="sm" onclick={() => remove(i)}>Remove</Button>
			</div>
			<div class="grid grid-cols-1 gap-3 sm:grid-cols-3">
				{#each properties as p (p.key)}
					<div class="flex flex-col gap-1.5">
						<label for="asset-{p.key}-{i}" class="text-sm font-medium">{p.label}</label>
						<Select
							id="asset-{p.key}-{i}"
							options={impactLevels}
							value={String(asset[p.key])}
							onchange={(v) => {
								asset[p.key] = Number(v);
								wizard.approved = false;
							}}
						/>
						<span class="text-xs text-ink-muted">{p.hint}</span>
					</div>
				{/each}
			</div>
		</div>
	{:else}
		<p class="rounded-control bg-sunken px-4 py-3 text-sm text-ink-muted">
			No asset yet. Without one, every requirement that depends on the risk stays in scope.
		</p>
	{/each}

	<div><Button variant="quiet" size="sm" onclick={add}>Add an asset</Button></div>
</section>
