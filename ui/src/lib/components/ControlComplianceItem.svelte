<script lang="ts">
	import type { EvaluationResult } from '$lib/api/evaluation';
	import type { Control } from '$lib/api/orchestrator';
	import AddEvaluationResultDialog, {
		type AddEvaluationResultEvent
	} from '$lib/components/AddEvaluationResultDialog.svelte';
	import { CheckCircle, EllipsisHorizontalCircle, XCircle } from '@steeze-ui/heroicons';
	import { Icon } from '@steeze-ui/svelte-icon';

	interface Props {
		result: EvaluationResult;
		control: Control | undefined;
		addResult: (event: AddEvaluationResultEvent) => void;
	}

	let { result, control, addResult }: Props = $props();

	let open = $state(false);
</script>

<div class="flex">
	<div class="mr-4 shrink-0">
		{#if result.status == 'EVALUATION_STATUS_COMPLIANT'}
			<Icon src={CheckCircle} theme="solid" class="h-8 w-8 text-green-800" />
		{:else if result.status == 'EVALUATION_STATUS_COMPLIANT_MANUALLY'}
			<Icon src={CheckCircle} theme="solid" class="h-8 w-8 text-green-800" />
		{:else if result.status == 'EVALUATION_STATUS_NOT_COMPLIANT_MANUALLY'}
			<Icon src={XCircle} theme="solid" class="h-8 w-8 text-red-800" />
		{:else if result.status == 'EVALUATION_STATUS_PENDING'}
			<button onclick={() => addResult}>
				<Icon src={EllipsisHorizontalCircle} theme="solid" class="h-8 w-8 text-gray-400" />
			</button>
		{:else}
			<a href="../../assessments/?filterIds={result.failingAssessmentResultIds.join(',')}">
				<Icon src={XCircle} theme="solid" class="h-8 w-8 text-red-800" />
			</a>
		{/if}
	</div>
	<div>
		<h4 class="text-base font-semibold">
			{#if control?.parentControlId === undefined}
				<div>
					{result.controlId}{#if control?.name !== undefined && control.name != ''}: {control.name}{/if}
				</div>
			{:else}
				{result.controlId}{#if control?.name !== undefined && control.name != ''}: {control.name}{/if}
			{/if}
		</h4>
		<p class="mt-1 text-sm text-gray-500">
			{control?.description}
		</p>
	</div>
</div>

<AddEvaluationResultDialog bind:open {addResult} />
