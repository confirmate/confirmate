<script lang="ts">
	/**
	 * A styled single choice (the combobox pattern with a listbox): a button shows the current choice and opens a list
	 * of options. It works with the keyboard (arrow keys, Home, End, Enter, Escape, type-ahead) and screen readers. Use
	 * it instead of the native select element, whose list cannot be styled.
	 */
	let {
		id,
		value = $bindable(''),
		options,
		size = 'md',
		onchange
	}: {
		id: string;
		value?: string;
		options: { value: string; label: string }[];
		size?: 'sm' | 'md';
		onchange?: (value: string) => void;
	} = $props();

	let open = $state(false);
	let active = $state(0);
	let root: HTMLDivElement | undefined = $state();
	let typed = '';
	let typedTimer: ReturnType<typeof setTimeout> | undefined;

	const selected = $derived(options.find((o) => o.value === value));
	const listId = $derived(`${id}-list`);

	function show() {
		active = Math.max(
			0,
			options.findIndex((o) => o.value === value)
		);
		open = true;
	}

	function choose(index: number) {
		const option = options[index];
		if (!option) return;
		value = option.value;
		open = false;
		onchange?.(option.value);
	}

	function move(delta: number) {
		active = (active + delta + options.length) % options.length;
	}

	function onkeydown(e: KeyboardEvent) {
		if (!open) {
			if (['ArrowDown', 'ArrowUp', 'Enter', ' '].includes(e.key)) {
				e.preventDefault();
				show();
			}
			return;
		}

		switch (e.key) {
			case 'ArrowDown':
				e.preventDefault();
				move(1);
				break;
			case 'ArrowUp':
				e.preventDefault();
				move(-1);
				break;
			case 'Home':
				e.preventDefault();
				active = 0;
				break;
			case 'End':
				e.preventDefault();
				active = options.length - 1;
				break;
			case 'Enter':
			case ' ':
				e.preventDefault();
				choose(active);
				break;
			case 'Escape':
				e.preventDefault();
				open = false;
				break;
			case 'Tab':
				open = false;
				break;
			default:
				// Type-ahead: jump to the next option that starts with the typed letters.
				if (e.key.length === 1 && !e.ctrlKey && !e.metaKey) {
					typed += e.key.toLowerCase();
					clearTimeout(typedTimer);
					typedTimer = setTimeout(() => (typed = ''), 600);
					const hit = options.findIndex((o) => o.label.toLowerCase().startsWith(typed));
					if (hit >= 0) active = hit;
				}
		}
	}

	function onfocusout(e: FocusEvent) {
		if (!root?.contains(e.relatedTarget as Node | null)) open = false;
	}
</script>

<div bind:this={root} class="relative {size === 'sm' ? 'inline-block' : 'block w-full'}" {onfocusout}>
	<button
		{id}
		type="button"
		role="combobox"
		aria-haspopup="listbox"
		aria-expanded={open}
		aria-controls={listId}
		aria-activedescendant={open ? `${id}-option-${active}` : undefined}
		onclick={() => (open ? (open = false) : show())}
		{onkeydown}
		class="flex w-full cursor-pointer items-center justify-between gap-3 border border-line-strong bg-surface text-left font-body text-ink rounded-control focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-brand {size ===
		'sm'
			? 'min-h-8 px-2.5 text-sm'
			: 'min-h-11 px-3.5'}"
	>
		<span class="truncate">{selected?.label ?? ''}</span>
		<svg class="size-4 shrink-0 text-ink-muted" viewBox="0 0 20 20" fill="none" aria-hidden="true">
			<path d="M5 7.5 10 12.5 15 7.5" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round" />
		</svg>
	</button>

	{#if open}
		<ul
			id={listId}
			role="listbox"
			tabindex="-1"
			class="absolute z-50 mt-1 max-h-64 min-w-full overflow-auto rounded-control border border-line bg-surface py-1 font-body shadow-lg"
		>
			{#each options as option, i (option.value)}
				<!-- The keyboard is handled by the combobox button (aria-activedescendant), as in the ARIA pattern. -->
				<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_noninteractive_element_interactions -->
				<li
					id="{id}-option-{i}"
					role="option"
					aria-selected={option.value === value}
					onmousedown={(e) => {
						// Keep the focus on the button, so that the list does not close before the click.
						e.preventDefault();
					}}
					onclick={() => choose(i)}
					onmouseenter={() => (active = i)}
					class="flex cursor-pointer items-center justify-between gap-3 px-3.5 py-2 whitespace-nowrap {i === active
						? 'bg-brand-soft text-brand-ink'
						: 'text-ink'} {option.value === value ? 'font-medium' : ''}"
				>
					{option.label}
					{#if option.value === value}
						<svg class="size-4 shrink-0 text-brand" viewBox="0 0 20 20" fill="none" aria-hidden="true">
							<path d="M4.5 10.5 8.5 14.5 15.5 6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
						</svg>
					{/if}
				</li>
			{/each}
		</ul>
	{/if}
</div>
