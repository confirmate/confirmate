<script lang="ts">
	import AuditScopeCard from '$lib/components/toe/AuditScopeCard.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import SectionHeader from '$lib/components/ui/SectionHeader.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const catalogMap = $derived(new Map(data.catalogs.map((c) => [c.id, c])));
</script>

<div>
	<SectionHeader
		title="Audit Scopes"
		description="Each audit scope evaluates this target against a compliance catalog."
	>
		{#snippet actions()}
			<Button href="/toe/{data.toe.id}/audit-scopes/new/">New Audit Scope</Button>
		{/snippet}
	</SectionHeader>

	<div class="mt-6">
		{#if data.auditScopes.length > 0}
			<ul class="space-y-3">
				{#each data.auditScopes as auditScope}
					<li>
						<AuditScopeCard {auditScope} catalog={catalogMap.get(auditScope.catalogId)} />
					</li>
				{/each}
			</ul>
		{:else}
			<EmptyState
				title="No audit scopes yet"
				description="Create your first audit scope to start tracking compliance against a catalog."
				actionHref="/toe/{data.toe.id}/audit-scopes/new/"
				actionLabel="New Audit Scope"
			/>
		{/if}
	</div>
</div>
