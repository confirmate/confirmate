<script lang="ts">
	import { Field, Select, TextInput } from '$lib/ds';
	import type { FrameworkProfile } from '../../profile';
	import { wizard } from '../../wizard.svelte';

	let { profile }: { profile: FrameworkProfile } = $props();
</script>

<div class="flex flex-col gap-6">
	{#each profile.details as field (field.id)}
		<Field label={field.label} hint={field.hint} id="detail-{field.id}">
			{#if field.type === 'select'}
				<Select
					id="detail-{field.id}"
					bind:value={wizard.details[field.id]}
					options={(field.options ?? []).map((o) => ({ value: o, label: o }))}
					onchange={() => wizard.applyChoices(profile)}
				/>
			{:else}
				<TextInput id="detail-{field.id}" type={field.type} bind:value={wizard.details[field.id]} />
			{/if}
		</Field>
	{/each}
</div>
