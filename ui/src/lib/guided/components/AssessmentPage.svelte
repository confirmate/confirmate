<script lang="ts">
	import { resolve } from '$app/paths';
	import { Button, Card, Chip, PageHeading, Tabs } from '$lib/ds';
	import { connectivityOptions, impactName, operatorSkillOptions, physicalAccessOptions } from '../risk';
	import { notApplicableReasons } from '../scope';
	import type { AssessmentDecision, RiskAssessmentView, TargetDetail } from '../source';

	let { target, assessment }: { target: TargetDetail; assessment?: RiskAssessmentView } = $props();

	let tab = $state('profile');

	const sourceLabels = { rule: 'From the risk profile', import: 'Imported', person: 'Decided by a person' };

	// The decisions in groups: by the reason for not applicable, then accepted risks, then requirements that a person kept
	const groups = $derived.by(() => {
		const decisions = assessment?.decisions ?? [];
		const out = notApplicableReasons
			.map((r) => ({
				key: r.value as string,
				title: `Not applicable: ${r.label.toLowerCase()}`,
				items: decisions.filter((d) => d.outcome === 'not-applicable' && (d.reason ?? 'TARGET_ABSENT') === r.value)
			}))
			.filter((g) => g.items.length > 0);
		const risks = decisions.filter((d) => d.outcome === 'accepted-risk');
		const kept = decisions.filter((d) => d.outcome === 'kept');
		if (risks.length) out.push({ key: 'risk', title: 'Applies, and the residual risk is accepted', items: risks });
		if (kept.length) out.push({ key: 'kept', title: 'Kept although the risk profile would take it out', items: kept });
		return out;
	});

	const tabs = $derived([
		{ id: 'profile', label: 'Risk profile' },
		{ id: 'changes', label: 'What it changed', badge: String(assessment?.decisions.length ?? 0) },
		{ id: 'versions', label: 'Versions', badge: String(assessment?.versions.length ?? 0) }
	]);

	const environment = $derived(
		assessment?.profile
			? [
					{ label: 'Who can reach the device', options: physicalAccessOptions, value: assessment.profile.physicalAccess },
					{ label: 'How the product can be reached', options: connectivityOptions, value: assessment.profile.connectivity },
					{ label: 'Who operates it', options: operatorSkillOptions, value: assessment.profile.operatorSkill }
				]
			: []
	);
</script>

{#snippet decision(d: AssessmentDecision)}
	<li class="flex flex-wrap gap-x-6 gap-y-2 border-t border-line px-3 py-4 first:border-t-0">
		<div class="flex min-w-[18rem] flex-1 flex-col gap-1">
			<strong class="font-medium">{d.title}</strong>
			<span class="text-[0.8125rem] text-ink-muted">
				<span class="font-code">{d.requirementId}</span>{d.reference ? `, ${d.reference}` : ''}
			</span>
			{#if d.text}<p class="mt-1.5 text-ink-soft">{d.text}</p>{/if}
			{#if d.compensatedBy}<p class="text-sm text-ink-muted">Fulfilled instead: {d.compensatedBy}</p>{/if}
		</div>
		<div class="w-52 text-sm text-ink-muted">{sourceLabels[d.source]}</div>
	</li>
{/snippet}

<div class="flex max-w-4xl flex-col gap-6">
	{#if assessment}
		<PageHeading
			title="Risk assessment, version {assessment.version}"
			intro="{assessment.method}. Approved by {assessment.approvedBy} on {assessment.approvedAt}."
		/>
		<div class="-mt-3 flex flex-wrap gap-2">
			<Chip tone="ok">Approved</Chip>
			<Chip>Review by {assessment.reviewDueAt}</Chip>
		</div>
		{#if assessment.summary}<p class="text-ink-soft">{assessment.summary}</p>{/if}

		<div class="flex flex-col gap-5">
			<Tabs label="Parts of the risk assessment" {tabs} bind:value={tab} />

			<div role="tabpanel" id="panel-profile" aria-labelledby="tab-profile" hidden={tab !== 'profile'}>
				{#if tab === 'profile'}
					{#if assessment.profile}
						<div class="flex flex-col gap-5">
							<Card>
								<h2 class="text-lg font-semibold">What the product protects</h2>
								{#if assessment.profile.assets.length === 0}
									<p class="text-ink-soft">No asset is listed, so every requirement that depends on the risk applies.</p>
								{:else}
									<ul class="m-0 flex list-none flex-col p-0">
										{#each assessment.profile.assets as a (a.name)}
											<li class="flex flex-wrap items-center justify-between gap-x-6 gap-y-2 border-t border-line py-3 first:border-t-0">
												<strong class="font-medium">{a.name}</strong>
												<span class="flex flex-wrap gap-2 text-sm text-ink-soft">
													<Chip size="sm">Disclosure: {impactName(a.confidentiality)}</Chip>
													<Chip size="sm">Manipulation: {impactName(a.integrity)}</Chip>
													<Chip size="sm">Outage: {impactName(a.availability)}</Chip>
												</span>
											</li>
										{/each}
									</ul>
								{/if}
							</Card>
							<Card>
								<h2 class="text-lg font-semibold">Where it runs</h2>
								<dl class="m-0 flex flex-col">
									{#each environment as e (e.label)}
										{@const option = e.options.find((o) => o.value === e.value)}
										<div class="flex flex-wrap justify-between gap-x-6 gap-y-1 border-t border-line py-3 first:border-t-0">
											<dt class="text-ink-soft">{e.label}</dt>
											<dd class="m-0 font-medium">{option?.label ?? 'Not said'}</dd>
										</div>
									{/each}
								</dl>
							</Card>
						</div>
					{:else}
						<Card>
							<p class="text-ink-soft">This assessment was not made from a risk profile, so there are no assets and environment to show.</p>
						</Card>
					{/if}
				{/if}
			</div>

			<div role="tabpanel" id="panel-changes" aria-labelledby="tab-changes" hidden={tab !== 'changes'}>
				{#if tab === 'changes'}
					<div class="flex flex-col gap-5">
						{#each groups as g (g.key)}
							<Card padding="sm">
								<div class="flex items-baseline justify-between gap-3 px-3 pt-3">
									<h2 class="text-base font-semibold">{g.title}</h2>
									<span class="text-sm text-ink-muted">{g.items.length}</span>
								</div>
								<ul class="m-0 mt-2 flex list-none flex-col p-0">
									{#each g.items as d (d.requirementId)}{@render decision(d)}{/each}
								</ul>
							</Card>
						{:else}
							<Card>
								<p class="text-ink-soft">This assessment did not take out any requirement. Everything in the catalog applies.</p>
							</Card>
						{/each}
					</div>
				{/if}
			</div>

			<div role="tabpanel" id="panel-versions" aria-labelledby="tab-versions" hidden={tab !== 'versions'}>
				{#if tab === 'versions'}
					<Card padding="sm">
						<ul class="m-0 flex list-none flex-col p-0">
							{#each assessment.versions as v (v.version)}
								<li class="flex flex-wrap items-center justify-between gap-x-6 gap-y-1 border-t border-line px-3 py-3.5 first:border-t-0">
									<span class="flex items-center gap-3">
										<strong class="font-medium">Version {v.version}</strong>
										<Chip size="sm" tone={v.status === 'approved' ? 'ok' : 'neutral'}>{v.status}</Chip>
									</span>
									<span class="text-sm text-ink-muted">{v.createdAt}</span>
									{#if v.summary}<p class="w-full text-sm text-ink-soft">{v.summary}</p>{/if}
								</li>
							{/each}
						</ul>
					</Card>
					<p class="mt-3 text-sm text-ink-muted">
						A change of the treatment of a requirement creates a new version with all earlier decisions and the new one. The
						older versions stay as they were approved.
					</p>
				{/if}
			</div>
		</div>
	{:else}
		<PageHeading title="Risk assessment" intro="The risk assessment decides which requirements apply to {target.name}." />
		<Card>
			<p class="text-ink-soft">No risk assessment is approved for this Target of Evaluation yet. All requirements of the catalog apply.</p>
		</Card>
	{/if}

	<div>
		<Button variant="quiet" href={resolve(`/guided/${target.id}/`)}>Back to the overview</Button>
	</div>
</div>
