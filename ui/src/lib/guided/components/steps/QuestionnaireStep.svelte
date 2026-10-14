<script lang="ts">
	import { Banner, Segmented } from '$lib/ds';
	import type { Answer, FrameworkProfile } from '../../profile';
	import { activeNotes } from '../../scope';
	import { wizard } from '../../wizard.svelte';

	let { profile }: { profile: FrameworkProfile } = $props();

	const q = $derived(profile.questionnaire);
	const notes = $derived(activeNotes(q, wizard.answers));
	const options = [
		{ value: 'yes', label: 'Yes' },
		{ value: 'no', label: 'No' },
		{ value: 'unknown', label: "Don't know" }
	];
</script>

<div class="flex flex-col gap-7">
	<div class="grid grid-cols-1 gap-3 md:grid-cols-3">
		<div class="rounded-card border-2 border-brand bg-brand-soft p-4.5">
			<strong>Quick check</strong>
			<p class="text-sm text-ink-soft">Five questions, about five minutes. For products without an assessment yet.</p>
		</div>
		<div class="rounded-card border border-line-strong bg-surface p-4.5">
			<strong>Import from a tool</strong>
			<p class="text-sm text-ink-soft">Use the result of QuBA-libre or another tool. You do not answer anything twice.</p>
		</div>
		<div class="rounded-card border border-line-strong bg-surface p-4.5">
			<strong>Upload an assessment</strong>
			<p class="text-sm text-ink-soft">Attach the document you already have and tick the requirements yourself.</p>
		</div>
	</div>

	<div class="flex flex-col rounded-card border border-line bg-surface px-7 py-2">
		{#each q.questions as question (question.id)}
			<div class="flex flex-wrap items-center justify-between gap-5 border-b border-line py-5 last:border-b-0">
				<div class="flex min-w-[20rem] flex-1 flex-col gap-0.5">
					<strong class="font-medium">{question.text}</strong>
					{#if question.help}<span class="text-[0.8125rem] text-ink-muted">{question.help}</span>{/if}
				</div>
				<Segmented
					label={question.text}
					{options}
					value={wizard.answers[question.id]}
					onchange={(v) => (wizard.answers[question.id] = v as Answer)}
				/>
			</div>
		{/each}
	</div>

	{#each notes as note (note.text)}
		<Banner tone={note.tone === 'warn' ? 'warn' : 'brand'}>{note.text}</Banner>
	{/each}
</div>
