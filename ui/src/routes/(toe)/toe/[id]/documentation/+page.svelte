<script lang="ts">
  import type { SchemaEvidence as Evidence } from '$lib/api/openapi/evidence';
  import type { SchemaTargetOfEvaluation } from '$lib/api/openapi/orchestrator';
  import PageHeader from '$lib/components/navigation/PageHeader.svelte';
  import Card from '$lib/components/ui/Card.svelte';
  import Collapsible from '$lib/components/ui/Collapsible.svelte';
  import ProgressBar from '$lib/components/ui/ProgressBar.svelte';
  import TextField from '$lib/components/ui/TextField.svelte';
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
    const el = docHtmlElement;
    if (!el) return;
    const recount = () => {
      const filled = el.querySelectorAll('[data-field="filled"]').length;
      const missing = el.querySelectorAll('[data-field="missing"]').length;
      fieldCount = { total: filled + missing, filled };
    };
    recount();
    const observer = new MutationObserver(recount);
    observer.observe(el, { childList: true, subtree: true, attributes: true, attributeFilter: ['data-field'] });
    return () => observer.disconnect();
  });
</script>

<PageHeader
  title="Document Generator"
  description="CRA-compliant technical documentation for {data.toe.name}"
/>

<div class="mb-6">
  <div class="border-b border-gray-200">
    <nav class="-mb-px flex space-x-8" aria-label="Tabs">
      <a
        href="/toe/{data.toe.id}/documentation/"
        class="border-confirmate text-confirmate flex items-center gap-2 border-b-2 px-1 py-4 text-sm font-medium"
      >
        Document Generator
      </a>
    </nav>
  </div>
</div>

<Card>
  <ProgressBar value={fieldCount.filled} max={fieldCount.total} label="Field completion" />
</Card>

<div class="mb-6"></div>

<Collapsible title="Manufacturer Information" subtitle="Manual input" bind:open={showManufacturerForm}>
  <div class="grid grid-cols-1 gap-4 pt-4 sm:grid-cols-2">
    <TextField label="Manufacturer name" placeholder="Your Company" bind:value={manufacturer.name} />
    <TextField label="Trade name / trademark" placeholder="Trade Name" bind:value={manufacturer.tradeName} />
    <TextField label="Postal address" placeholder="123 Main Street, City" bind:value={manufacturer.postalAddress} />
    <TextField label="General contact email" type="email" placeholder="info@example.com" bind:value={manufacturer.generalEmail} />
    <TextField label="Security contact email" type="email" placeholder="security@example.com" bind:value={manufacturer.securityEmail} />
    <TextField label="Website" type="url" placeholder="https://example.com" bind:value={manufacturer.website} />
    <div class="sm:col-span-2">
      <TextField label="Alternative reporting channel" type="url" placeholder="https://security.example.com" bind:value={manufacturer.securityPortalUrl} />
    </div>
  </div>
</Collapsible>

<div class="mb-6"></div>

<Card padding="lg" bind:element={docHtmlElement}>
  <DocumentBody {evidences} bind:manufacturer bind:placeholders />
</Card>

<EvidencePopover />