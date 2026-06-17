<script lang="ts">
  import type { SchemaResource } from '$lib/api/openapi/evidence';
  import { popover, closePopover } from './popoverStore.svelte';
  import { tick } from 'svelte';

  let popoverEl: HTMLDivElement | undefined = $state();
  let position = $state({ top: 0, left: 0 });

  const resourceData = $derived.by(() => {
    const resource = popover.evidence?.resource;
    if (!resource) return null;
    const key = Object.keys(resource)[0] as keyof SchemaResource | undefined;
    return key ? resource[key] ?? null : null;
  });

  $effect(() => {
    if (popover.evidence && popover.anchor) {
      tick().then(updatePosition);
    }
  });

  $effect(() => {
    if (!popover.evidence) return;

    function handleClick(e: MouseEvent) {
      const target = e.target as Node;
      if (popoverEl && !popoverEl.contains(target) && popover.anchor && !popover.anchor.contains(target)) {
        closePopover();
      }
    }

    document.addEventListener('mousedown', handleClick);
    document.addEventListener('scroll', updatePosition, true);

    return () => {
      document.removeEventListener('mousedown', handleClick);
      document.removeEventListener('scroll', updatePosition, true);
    };
  });

  function updatePosition() {
    if (!popover.anchor || !popoverEl) return;

    const anchorRect = popover.anchor.getBoundingClientRect();
    const popoverRect = popoverEl.getBoundingClientRect();
    const viewportW = window.innerWidth;
    const viewportH = window.innerHeight;
    const gap = 12;

    let left = anchorRect.right + gap;
    if (left + popoverRect.width > viewportW - 16) {
      left = anchorRect.left - popoverRect.width - gap;
    }

    const anchorMidY = anchorRect.top + anchorRect.height / 2;
    let top = anchorMidY - popoverRect.height / 2;

    if (top + popoverRect.height > viewportH - 16) {
      top = viewportH - popoverRect.height - 16;
    }
    if (top < 16) top = 16;
    if (left < 16) left = 16;

    position = { top, left };
  }

  function formatValue(value: unknown): string {
    if (value === null || value === undefined) return '—';
    if (typeof value === 'object') return JSON.stringify(value, null, 2);
    return String(value);
  }
</script>

{#if popover.evidence}
  <div
    bind:this={popoverEl}
    class="fixed z-50 w-[320px] max-h-[55vh] bg-white rounded-lg shadow-lg ring-1 ring-gray-200 flex flex-col overflow-hidden animate-in"
    style="top: {position.top}px; left: {position.left}px;"
  >
    <div class="flex items-center justify-between px-4 py-2.5 border-b border-gray-100 bg-gray-50/50">
      <div class="flex items-center gap-2">
        <span class="inline-flex items-center px-1.5 py-0.5 rounded bg-confirmate text-[10px] font-medium text-white tracking-wide">
          {popover.evidence.toolId}
        </span>
        <span class="text-[11px] text-gray-400 font-medium">{popover.resourceType}</span>
      </div>
      <button
        type="button"
        class="p-0.5 text-gray-400 hover:text-gray-600 transition-colors rounded"
        onclick={closePopover}
        aria-label="Close"
      >
        <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    </div>

    <div class="flex gap-4 px-4 py-2 border-b border-gray-50 text-[11px]">
      <div class="flex flex-col">
        <span class="text-gray-400 font-medium">Timestamp</span>
        <span class="text-gray-600 tabular-nums">{popover.evidence.timestamp}</span>
      </div>
      <div class="flex flex-col">
        <span class="text-gray-400 font-medium">Evidence ID</span>
        <span class="text-gray-600 font-mono" title={popover.evidence.id}>{popover.evidence.id}</span>
      </div>
    </div>

    <div class="flex-1 overflow-y-auto px-3 py-2.5">
      {#if resourceData}
        <div class="space-y-px">
          {#each Object.entries(resourceData) as [key, value]}
            <div class="rounded-md px-2.5 py-1.5 {key === popover.field ? 'bg-confirmate/10' : 'hover:bg-gray-50'} transition-colors">
              <div class="text-[10px] font-medium uppercase tracking-wider {key === popover.field ? 'text-confirmate' : 'text-gray-300'}">
                {key}
              </div>
              {#if typeof value === 'object' && value !== null}
                <pre class="text-[12px] text-gray-700 whitespace-pre-wrap break-all mt-0.5 leading-relaxed">{JSON.stringify(value, null, 2)}</pre>
              {:else}
                <div class="text-[13px] {key === popover.field ? 'text-confirmate font-medium' : 'text-gray-700'} mt-0.5">
                  {formatValue(value)}
                </div>
              {/if}
            </div>
          {/each}
        </div>
      {/if}
    </div>
  </div>
{/if}

<style>
  .animate-in {
    animation: popover-in 120ms ease-out;
  }
  @keyframes popover-in {
    from {
      opacity: 0;
      transform: scale(0.97) translateY(2px);
    }
    to {
      opacity: 1;
      transform: scale(1) translateY(0);
    }
  }
</style>