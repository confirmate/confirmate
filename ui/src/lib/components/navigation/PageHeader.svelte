<script lang="ts">
  interface Props {
    title: string;
    description?: string;
  }

  let { title, description }: Props = $props();
</script>

<div class="mb-6 border-b border-gray-200 pb-6">
  <h1 class="text-2xl font-bold text-gray-900">{title}</h1>
  {#if description}
    <p class="mt-1 text-sm text-gray-500">{description}</p>
  {/if}
</div>