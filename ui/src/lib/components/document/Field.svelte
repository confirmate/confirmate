<script lang="ts">
  import type { SchemaEvidence as Evidence } from '$lib/api/openapi/evidence';
  import { openPopover, popover } from './popoverStore.svelte';
  import { tick } from 'svelte';

  interface Props {
    value?: string | number | null;
    placeholder: string;
    readonly?: boolean;
    evidence?: Evidence | null;
    resourceType?: string;
    field?: string;
  }

  let {
    value = $bindable(),
    placeholder,
    readonly = false,
    evidence = null,
    resourceType = '',
    field = ''
  }: Props = $props();

  const display = $derived(
    value !== undefined && value !== null && value !== '' ? String(value) : null
  );

  const isActive = $derived(
    !!evidence &&
      popover.evidence === evidence &&
      popover.resourceType === resourceType &&
      popover.field === field
  );

  let editing = $state(false);
  let inputEl: HTMLInputElement | undefined = $state();
  let draft = $state('');

  function onEvidenceClick(e: MouseEvent) {
    if (!evidence) return;
    openPopover(evidence, resourceType, field, e.currentTarget as HTMLElement);
  }

  async function startEdit() {
    draft = display ?? '';
    editing = true;
    await tick();
    inputEl?.focus();
    inputEl?.select();
  }

  function commit() {
    value = draft;
    editing = false;
  }

  function cancel() {
    editing = false;
  }

  function onKey(e: KeyboardEvent) {
    if (e.key === 'Enter') {
      e.preventDefault();
      commit();
    } else if (e.key === 'Escape') {
      e.preventDefault();
      cancel();
    }
  }
</script>

{#if readonly && evidence && display !== null}
  <button
    type="button"
    data-field="filled"
    class="inline rounded px-1 py-0.5 text-sm font-medium transition-all duration-150 cursor-pointer
      {isActive
        ? 'bg-confirmate/10 text-confirmate ring-1 ring-confirmate'
        : 'bg-gray-50 text-gray-800 ring-1 ring-gray-200 hover:bg-confirmate/5 hover:text-confirmate hover:ring-confirmate/50'}"
    onclick={onEvidenceClick}
  >
    {display}
  </button>
{:else if readonly}
  {#if display !== null}
    <span data-field="filled" class="text-sm font-medium text-gray-700">{display}</span>
  {:else}
    <span
      data-field="missing"
      class="inline rounded px-1.5 py-0.5 bg-amber-50 text-amber-600 ring-1 ring-amber-200 text-xs font-medium"
    >
      {placeholder}
    </span>
  {/if}
{:else if editing}
  <input
    bind:this={inputEl}
    type="text"
    bind:value={draft}
    onblur={commit}
    onkeydown={onKey}
    class="inline-block min-w-[8ch] rounded border border-gray-300 px-2 py-0.5 text-sm font-medium text-gray-900 focus:border-confirmate focus:outline-none focus:ring-1 focus:ring-confirmate"
  />
{:else if display !== null}
  <button
    type="button"
    data-field="filled"
    class="group inline-flex items-center gap-1 rounded border border-dashed border-gray-300 px-2 py-0.5 text-sm font-medium text-gray-700 bg-white hover:border-confirmate hover:text-confirmate cursor-text transition-all duration-150"
    onclick={startEdit}
  >
    <span>{display}</span>
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 20 20"
      fill="currentColor"
      class="h-3 w-3 opacity-0 group-hover:opacity-70 transition-opacity duration-150 text-gray-400"
      aria-hidden="true"
    >
      <path d="M13.586 3.586a2 2 0 112.828 2.828l-.793.793-2.828-2.828.793-.793zM11.379 5.793L3 14.172V17h2.828l8.379-8.379-2.828-2.828z" />
    </svg>
  </button>
{:else}
  <button
    type="button"
    data-field="missing"
    class="group inline-flex items-center gap-1 rounded border border-dashed border-amber-300 bg-amber-50 px-2 py-0.5 text-xs font-medium text-amber-600 hover:bg-amber-100 hover:text-amber-700 cursor-text transition-all duration-150"
    onclick={startEdit}
  >
    <span>{placeholder}</span>
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 20 20"
      fill="currentColor"
      class="h-3 w-3 opacity-0 group-hover:opacity-70 transition-opacity duration-150"
      aria-hidden="true"
    >
      <path d="M13.586 3.586a2 2 0 112.828 2.828l-.793.793-2.828-2.828.793-.793zM11.379 5.793L3 14.172V17h2.828l8.379-8.379-2.828-2.828z" />
    </svg>
  </button>
{/if}