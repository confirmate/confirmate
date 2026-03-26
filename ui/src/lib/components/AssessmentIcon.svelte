<script lang="ts">
	import type { AssessmentResult } from '$lib/api/assessment';
	import { CheckCircle, XCircle } from '@steeze-ui/heroicons';
	import { Icon } from '@steeze-ui/svelte-icon';

	interface Props {
		result: AssessmentResult;
	}

	let { result }: Props = $props();
</script>

{#if result.compliant}
	<Icon src={CheckCircle} theme="solid" class="mr-2 h-5 w-5 text-green-800" />
{:else}
	<Icon src={XCircle} theme="solid" class="mr-2 h-5 w-5 text-red-800" />
{/if}
