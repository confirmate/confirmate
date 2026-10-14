<script lang="ts">
	import { resolve } from '$app/paths';
	import { Chip, StackedProgress } from '$lib/ds';
	import type { TargetSummary } from '../source';

	let { targets }: { targets: TargetSummary[] } = $props();

	const th = 'px-4 py-2.5 text-left text-xs font-semibold uppercase tracking-wider text-ink-muted';
</script>

<div class="overflow-x-auto rounded-card border border-line bg-surface">
	<table class="w-full border-collapse text-left">
		<thead class="border-b border-line bg-sunken">
			<tr>
				<th scope="col" class={th}>Target of Evaluation</th>
				<th scope="col" class="{th} hidden md:table-cell">Framework</th>
				<th scope="col" class="{th} min-w-44">Requirements met</th>
				<th scope="col" class="{th} text-right">Attention</th>
				<th scope="col" class="{th} hidden lg:table-cell">Next deadline</th>
			</tr>
		</thead>
		<tbody>
			{#each targets as t (t.id)}
				<tr class="border-t border-line first:border-t-0 hover:bg-sunken/60">
					<td class="px-4 py-3.5">
						<a href={resolve(`/guided/${t.id}/`)} class="font-medium text-ink no-underline hover:text-brand">{t.name}</a>
						<div class="text-[0.8125rem] text-ink-muted">{t.typeLabel}, {t.subtitle}</div>
					</td>
					<td class="hidden px-4 py-3.5 md:table-cell"><Chip tone="brand" size="sm">{t.frameworkLabel}</Chip></td>
					<td class="px-4 py-3.5">
						<div class="flex flex-col gap-1.5">
							<span class="text-sm tabular-nums"><strong class="font-semibold">{t.progress.met}</strong> <span class="text-ink-muted">of {t.progress.total}</span></span>
							<StackedProgress
								total={t.progress.total}
								segments={[
									{ value: t.progress.met, tone: 'ok', label: 'met' },
									{ value: t.progress.attention, tone: 'bad', label: 'need attention' }
								]}
							/>
						</div>
					</td>
					<td class="px-4 py-3.5 text-right text-sm tabular-nums">
						{#if t.progress.attention > 0}
							<span class="font-semibold text-bad">{t.progress.attention}</span>
						{:else}
							<span class="text-ink-muted">0</span>
						{/if}
					</td>
					<td class="hidden max-w-xs px-4 py-3.5 text-sm text-ink-soft lg:table-cell">{t.nextDeadline ?? ''}</td>
				</tr>
			{/each}
		</tbody>
	</table>
</div>
