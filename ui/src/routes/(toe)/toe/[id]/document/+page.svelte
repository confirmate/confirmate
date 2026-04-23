<script lang="ts">
  import type { SchemaEvidence as Evidence } from '$lib/api/openapi/evidence';
  import type { SchemaTargetOfEvaluation } from '$lib/api/openapi/orchestrator';
  import PageHeader from '$lib/components/navigation/PageHeader.svelte';
  import DocumentBody from '$lib/components/document/DocumentBody.svelte';
  import EvidencePopover from '$lib/components/document/EvidencePopover.svelte';
  import type { ManufacturerInfo, Placeholders } from '$lib/types/evidence';

  let { data }: { 
    data: { 
      toe: SchemaTargetOfEvaluation; 
      evidences: Evidence[] 
    } 
  } = $props();

  let manufacturer = $state<ManufacturerInfo>({
    name: '',
    tradeName: '',
    postalAddress: '',
    generalEmail: '',
    securityEmail: '',
    website: '',
    securityPortalUrl: ''
  });
  let placeholders = $state<Placeholders>({
    userInstructionsReference: '',
    automaticUpdateMethod: '',
    dataRemovalMethod: '',
    disableUpdatesPath: '',
    integrationDocUrl: '',
    architectureDocumentReference: '',
    hardwareLayoutReference: '',
    updateInstructionsUrl: ''
  });
  let showManufacturerForm = $state(false);

  const evidences = $state(data.evidences);

  $effect(() => {
    manufacturer = {
      name: '',
      tradeName: '',
      postalAddress: '',
      generalEmail: '',
      securityEmail: '',
      website: '',
      securityPortalUrl: ''
    };
    placeholders = {
      userInstructionsReference: '',
      automaticUpdateMethod: '',
      dataRemovalMethod: '',
      disableUpdatesPath: '',
      integrationDocUrl: '',
      architectureDocumentReference: '',
      hardwareLayoutReference: '',
      updateInstructionsUrl: ''
    };
  });

  let docHtmlElement: HTMLDivElement | undefined = $state();
  let fieldCount = $state({ total: 0, filled: 0 });

  $effect(() => {
    if (!docHtmlElement) return;
    const recount = () => {
      const filled = docHtmlElement!.querySelectorAll('[data-field="filled"]').length;
      const missing = docHtmlElement!.querySelectorAll('[data-field="missing"]').length;
      fieldCount = { total: filled + missing, filled };
    };
    recount();
    const observer = new MutationObserver(recount);
    observer.observe(docHtmlElement, { childList: true, subtree: true, attributes: true, attributeFilter: ['data-field'] });
    return () => observer.disconnect();
  });

  const progressPercent = $derived(
    fieldCount.total > 0 ? Math.round((fieldCount.filled / fieldCount.total) * 100) : 0
  );
</script>

<PageHeader
  title="Document Generator"
  description="CRA-compliant technical documentation for {data.toe.name}"
/>

<div class="mb-6">
  <div class="border-b border-gray-200">
    <nav class="-mb-px flex space-x-8" aria-label="Tabs">
      <a
        href="/toe/{data.toe.id}/document/"
        class="border-confirmate text-confirmate flex items-center gap-2 border-b-2 px-1 py-4 text-sm font-medium"
      >
        Document Generator
      </a>
    </nav>
  </div>
</div>

<div class="mb-6 rounded-lg border border-gray-200 bg-white px-5 py-4 shadow-sm">
  <div class="mb-2 flex items-center justify-between">
    <span class="text-sm font-medium text-gray-700">Field completion</span>
    <span class="text-sm tabular-nums text-gray-500">{fieldCount.filled} / {fieldCount.total} ({progressPercent}%)</span>
  </div>
  <div class="h-2 w-full rounded-full bg-gray-100">
    <div
      class="h-2 rounded-full transition-all duration-500 ease-out {progressPercent === 100 ? 'bg-emerald-500' : 'bg-confirmate'}"
      style="width: {progressPercent}%"
    ></div>
  </div>
</div>

<div class="mb-6 overflow-hidden rounded-lg border border-gray-200 bg-white shadow-sm">
  <button
    class="flex w-full items-center justify-between px-6 py-4 text-left hover:bg-gray-50"
    onclick={() => (showManufacturerForm = !showManufacturerForm)}
  >
    <div class="flex items-center gap-2">
      <svg
        class="h-5 w-5 text-gray-400 transition-transform duration-200 {showManufacturerForm ? 'rotate-90' : ''}"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor"
      >
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
      </svg>
      <span class="text-sm font-semibold text-gray-900">Manufacturer Information</span>
    </div>
    <span class="text-xs text-gray-400">Manual input</span>
  </button>

  {#if showManufacturerForm}
    <div class="border-t border-gray-100 px-6 pb-6">
      <div class="grid grid-cols-1 gap-4 pt-4 sm:grid-cols-2">
        <label class="flex flex-col gap-1.5">
          <span class="text-sm font-medium text-gray-700">Manufacturer name</span>
          <input
            type="text"
            class="rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-confirmate focus:outline-none focus:ring-1 focus:ring-confirmate"
            placeholder="Your Company"
            bind:value={manufacturer.name}
          />
        </label>
        <label class="flex flex-col gap-1.5">
          <span class="text-sm font-medium text-gray-700">Trade name / trademark</span>
          <input
            type="text"
            class="rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-confirmate focus:outline-none focus:ring-1 focus:ring-confirmate"
            placeholder="Trade Name"
            bind:value={manufacturer.tradeName}
          />
        </label>
        <label class="flex flex-col gap-1.5">
          <span class="text-sm font-medium text-gray-700">Postal address</span>
          <input
            type="text"
            class="rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-confirmate focus:outline-none focus:ring-1 focus:ring-confirmate"
            placeholder="123 Main Street, City"
            bind:value={manufacturer.postalAddress}
          />
        </label>
        <label class="flex flex-col gap-1.5">
          <span class="text-sm font-medium text-gray-700">General contact email</span>
          <input
            type="email"
            class="rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-confirmate focus:outline-none focus:ring-1 focus:ring-confirmate"
            placeholder="info@example.com"
            bind:value={manufacturer.generalEmail}
          />
        </label>
        <label class="flex flex-col gap-1.5">
          <span class="text-sm font-medium text-gray-700">Security contact email</span>
          <input
            type="email"
            class="rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-confirmate focus:outline-none focus:ring-1 focus:ring-confirmate"
            placeholder="security@example.com"
            bind:value={manufacturer.securityEmail}
          />
        </label>
        <label class="flex flex-col gap-1.5">
          <span class="text-sm font-medium text-gray-700">Website</span>
          <input
            type="url"
            class="rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-confirmate focus:outline-none focus:ring-1 focus:ring-confirmate"
            placeholder="https://example.com"
            bind:value={manufacturer.website}
          />
        </label>
        <label class="flex flex-col gap-1.5 sm:col-span-2">
          <span class="text-sm font-medium text-gray-700">Alternative reporting channel</span>
          <input
            type="url"
            class="rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-confirmate focus:outline-none focus:ring-1 focus:ring-confirmate"
            placeholder="https://security.example.com"
            bind:value={manufacturer.securityPortalUrl}
          />
        </label>
      </div>
    </div>
  {/if}
</div>

<div bind:this={docHtmlElement} class="rounded-lg border border-gray-200 bg-white px-10 py-8 shadow-sm">
  <DocumentBody {evidences} bind:manufacturer bind:placeholders />
</div>

<EvidencePopover />