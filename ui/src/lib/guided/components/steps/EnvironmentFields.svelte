<script lang="ts">
	import { Segmented } from '$lib/ds';
	import type { Connectivity, OperatorSkill, PhysicalAccess } from '../../profile';
	import { connectivityOptions, operatorSkillOptions, physicalAccessOptions } from '../../risk';
	import { wizard } from '../../wizard.svelte';

	const fields = [
		{ id: 'physical', label: 'Who can reach the device?', options: physicalAccessOptions, key: 'physicalAccess' },
		{ id: 'connectivity', label: 'How can the product be reached?', options: connectivityOptions, key: 'connectivity' },
		{ id: 'operator', label: 'Who operates it?', options: operatorSkillOptions, key: 'operatorSkill' }
	] as const;

	function set(key: (typeof fields)[number]['key'], value: string) {
		if (key === 'physicalAccess') wizard.risk.physicalAccess = value as PhysicalAccess;
		else if (key === 'connectivity') wizard.risk.connectivity = value as Connectivity;
		else wizard.risk.operatorSkill = value as OperatorSkill;
		wizard.approved = false;
	}
</script>

<section class="flex flex-col gap-1 rounded-card border border-line bg-surface px-7 py-2" aria-labelledby="environment-heading">
	<h2 id="environment-heading" class="pt-4 text-[1.0625rem] font-semibold">Where does it run?</h2>
	{#each fields as f (f.id)}
		{@const current = wizard.risk[f.key]}
		<div class="flex flex-wrap items-center justify-between gap-5 border-b border-line py-5 last:border-b-0">
			<div class="flex min-w-[16rem] flex-1 flex-col gap-0.5">
				<strong class="font-medium">{f.label}</strong>
				<span class="text-[0.8125rem] text-ink-muted">
					{f.options.find((o) => o.value === current)?.hint ?? 'Not said. Requirements that depend on this stay in scope.'}
				</span>
			</div>
			<Segmented
				label={f.label}
				options={f.options.map((o) => ({ value: o.value, label: o.label, title: o.hint }))}
				value={current}
				onchange={(v) => set(f.key, v)}
			/>
		</div>
	{/each}
</section>
