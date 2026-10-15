<script lang="ts">
	/**
	 * A row of tabs that switch between panels of one page. The panels belong to the caller: give each panel
	 * role="tabpanel", id="panel-<id>" and aria-labelledby="tab-<id>". The tabs work with the keyboard (arrow keys, Home and
	 * End move and select) and show an optional short summary, for example a count.
	 */
	let {
		label,
		tabs,
		value = $bindable(),
		onchange
	}: {
		label: string;
		tabs: { id: string; label: string; badge?: string }[];
		value?: string;
		onchange?: (id: string) => void;
	} = $props();

	function select(id: string) {
		value = id;
		onchange?.(id);
		document.getElementById(`tab-${id}`)?.focus();
	}

	function onkeydown(e: KeyboardEvent, index: number) {
		const last = tabs.length - 1;
		const target =
			e.key === 'ArrowRight' ? (index === last ? 0 : index + 1)
			: e.key === 'ArrowLeft' ? (index === 0 ? last : index - 1)
			: e.key === 'Home' ? 0
			: e.key === 'End' ? last
			: -1;
		if (target < 0) return;
		e.preventDefault();
		select(tabs[target].id);
	}
</script>

<div role="tablist" aria-label={label} class="flex gap-1 overflow-x-auto border-b border-line-strong">
	{#each tabs as tab, i (tab.id)}
		<button
			type="button"
			role="tab"
			id="tab-{tab.id}"
			aria-selected={value === tab.id}
			aria-controls="panel-{tab.id}"
			tabindex={value === tab.id ? 0 : -1}
			onclick={() => select(tab.id)}
			onkeydown={(e) => onkeydown(e, i)}
			class="-mb-px flex min-h-11 cursor-pointer items-center gap-2 whitespace-nowrap border-b-2 px-4.5 font-body focus-visible:outline-2 focus-visible:outline-offset-[-2px] focus-visible:outline-brand {value ===
			tab.id
				? 'border-brand font-semibold text-brand-ink'
				: 'border-transparent text-ink-soft hover:border-line-strong hover:text-ink'}"
		>
			{tab.label}
			{#if tab.badge}
				<span class="rounded-pill bg-sunken px-2 py-0.5 text-xs font-semibold text-ink-soft">{tab.badge}</span>
			{/if}
		</button>
	{/each}
</div>
