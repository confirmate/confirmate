<script lang="ts">
	import { invalidate } from '$app/navigation';
	import { createEvaluationResult, type EvaluationResult } from '$lib/api/evaluation';
	import type { Control } from '$lib/api/orchestrator';
	import type { AddEvaluationResultEvent } from '$lib/components/AddEvaluationResultDialog.svelte';
	import ControlComplianceItem from '$lib/components/ControlComplianceItem.svelte';
	import { Minus, Plus } from '@steeze-ui/heroicons';
	import { Icon } from '@steeze-ui/svelte-icon';
	import type { PageData } from './$types';
	import { createDisclosure } from 'svelte-headlessui';

	interface Props {
		data: PageData;
	}

	let { data }: Props = $props();

	interface TreeItemData {
		result: EvaluationResult;
		children: EvaluationResult[];
	}

	/**
	 * This function builds a tree-like structure out of the evaluation results,
	 * with the first level comprising of the top level controls. The second level
	 * consists of their sub controls.
	 *
	 * @param results
	 */
	function buildTree(results: EvaluationResult[], status: string[]): Map<string, TreeItemData> {
		const tree = new Map<string, TreeItemData>();

		for (const result of results) {
			if (status !== null && !status.includes(result.status) && status.length != 0) {
				continue;
			}

			// Top level control is at the first level of the tree
			if (result.parentControlId === undefined) {
				tree.set(result.controlId, {
					result: result,
					children: []
				});
			} else {
				// Look for parent tree item
				let parent = tree.get(result.parentControlId);
				if (parent === undefined) {
					continue;
				}

				if (status !== null && (status.includes(result.status) || status.length == 0)) {
					parent.children.push(result);
				}
			}
		}

		return tree;
	}

	async function addResult(e: AddEvaluationResultEvent, control?: Control) {
		if (control == undefined) {
			return;
		}

		let result: EvaluationResult = {
			id: '',
			controlId: control.id,
			targetOfEvaluationId: data.service.id,
			controlCategoryName: control.categoryName,
			controlCatalogId: data.catalog.id,
			parentControlId: control.parentControlId,
			status: e.status,
			timestamp: new Date().toISOString(),
			failingAssessmentResultIds: [],
			comment: e.comment,
			validUntil: e.validUntil.toISOString()
		};

		result = await createEvaluationResult(result);
		invalidate((url) => url.pathname == '/v1/evaluation/results');
	}
	let tree = $derived(buildTree(data.evaluations, data.filterStatus));

	const disc = createDisclosure({ label: 'Details', expanded: true });
</script>

<div class="border-b border-gray-200 pb-5">
	<h3 class="text-base leading-6 font-semibold text-gray-900">{data.catalog.name}</h3>
	<p class="mt-2 max-w-4xl text-sm text-gray-500">
		{data.catalog.description}
	</p>
</div>

<dl class="space-y-6 divide-y divide-gray-900/10">
	{#each [...tree] as [_, item] (item.result.id)}
		<div class="pt-6">
			<dt>
				<div class="flex w-full items-start justify-between text-left text-gray-900">
					<ControlComplianceItem
						result={item.result}
						control={data.controls.get(item.result.controlId)}
						addResult={(e) => addResult(e, data.controls.get(item.result.controlId))}
					/>
					<button use:disc.button>
						<span class="ml-6 flex h-7 items-center">
							{#if !open}
								<Icon src={Plus} class="h-6 w-6" aria-hidden="true" />
							{:else}
								<Icon src={Minus} class="h-6 w-6" aria-hidden="true" />
							{/if}
						</span>
					</button>
				</div>
				{#if $disc.expanded}
					<div class="mt-2 pr-12">
						{#each item.children as result (result.controlId)}
							<div class="mt-6 ml-12">
								<ControlComplianceItem
									{result}
									{addResult}
									control={data.controls.get(result.controlId)}
								/>
							</div>
						{/each}
					</div>
				{/if}
			</dt>
		</div>
	{/each}
</dl>
