<script lang="ts">
  import type { Snippet } from 'svelte';

  interface Props {
    title: string;
    description?: string;
    actions?: Snippet;
  }

  let { title, description, actions }: Props = $props();
</script>

<div class="mb-6 flex items-start justify-between border-b border-gray-200 pb-6">
  <div>
    <h1 class="text-2xl font-bold text-gray-900">{title}</h1>
    {#if description}
      <p class="mt-1 text-sm text-gray-500">{description}</p>
    {/if}
  </div>
  {#if actions}
    <div class="ml-4 shrink-0">
      {@render actions()}
    </div>
  {/if}
</div>
