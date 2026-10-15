<script module lang="ts">
	export type StepState = 'done' | 'current' | 'todo';
</script>

<script lang="ts">
	let {
		steps,
		heading
	}: { steps: { id: string; label: string; state: StepState; href?: string }[]; heading?: string } =
		$props();
</script>

<div class="font-body">
	{#if heading}
		<p class="mb-3 text-[0.8125rem] uppercase tracking-wider text-ink-muted">{heading}</p>
	{/if}
	<ol class="m-0 list-none p-0">
		{#each steps as step, i (step.id)}
			<li class="flex items-center gap-3 py-2.5" aria-current={step.state === 'current' ? 'step' : undefined}>
				<span
					class="flex size-7 shrink-0 items-center justify-center rounded-full text-[0.8125rem] font-semibold {step.state ===
					'done'
						? 'bg-ok text-on-brand'
						: step.state === 'current'
							? 'bg-brand text-on-brand'
							: 'bg-neutral-soft text-ink-soft'}"
				>
					{i + 1}
				</span>
				{#if step.href && step.state === 'done'}
					<a href={step.href} class="text-ink no-underline hover:underline">{step.label}</a>
				{:else}
					<span class={step.state === 'current' ? 'font-semibold text-ink' : 'text-ink-soft'}>{step.label}</span>
				{/if}
			</li>
		{/each}
	</ol>
</div>
