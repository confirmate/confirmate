<script lang="ts">
	import { run } from 'svelte/legacy';

	import { goto } from '$app/navigation';
	import type { ComplianceStatus } from '$lib/api/evaluation';
	import type { AuditScope } from '$lib/api/orchestrator';
	import { Chart, type ChartConfiguration, type ChartData } from 'chart.js/auto';
	import { onMount } from 'svelte';

	let canvas: HTMLCanvasElement = $state();
	let chart: Chart<'doughnut', { status: string[]; num: number }[]>;
	interface Props {
		compliance: Map<string, ComplianceStatus>;
		auditScope: AuditScope;
	}

	let { compliance, auditScope }: Props = $props();

	let merge = $state(true);

	let config: ChartConfiguration<'doughnut', { status: string[]; num: number }[]> = {
		type: 'doughnut',
		data: buildData(merge),
		options: {
			parsing: {
				key: 'num'
			},
			animation: false,
			plugins: {
				tooltip: {
					titleFont: {
						family: 'InterVariable'
					},
					bodyFont: {
						family: 'InterVariable'
					},
					footerFont: {
						family: 'InterVariable'
					}
				},
				legend: {
					display: false,
					labels: {
						font: {
							family: 'InterVariable'
						}
					}
				}
			}
		}
	};

	onMount(() => {
		chart = new Chart(canvas, config);

		canvas.onclick = (evt) => {
			const res = chart.getElementsAtEventForMode(evt, 'nearest', { intersect: true }, true);

			if (res.length === 0) {
				return;
			} else {
				const data = chart.data.datasets[0].data[res[0].index];
				const params = new URLSearchParams();

				for (const s of data.status) {
					params.append('status', s);
				}

				goto(
					`/cloud/${auditScope.targetOfEvaluationId}/compliance/${auditScope.catalogId}?${params.toString()}`
				);
			}
		};
	});

	function updateChart(data: ChartData<'doughnut', { status: string[]; num: number }[]>) {
		if (chart) {
			chart.data = data;
			chart.update();
		}
	}

	function buildData(merge: boolean): ChartData<'doughnut', { status: string[]; num: number }[]> {
		if (merge) {
			return {
				labels: ['Non Compliant', 'Compliant', 'Waiting for Data'],
				datasets: [
					{
						label: auditScope.catalogId,
						data: [
							filter([
								'EVALUATION_STATUS_NOT_COMPLIANT',
								'EVALUATION_STATUS_NOT_COMPLIANT_MANUALLY'
							]),
							filter(['EVALUATION_STATUS_COMPLIANT', 'EVALUATION_STATUS_COMPLIANT_MANUALLY']),
							filter(['EVALUATION_STATUS_PENDING'])
						],
						backgroundColor: ['#991b1b', '#166534', '#d4d4d4'],
						hoverOffset: 4
					}
				]
			};
		} else {
			return {
				labels: [
					'Non Compliant',
					'Manually set to Non Compliant',
					'Compliant',
					'Manually set to Compliant',
					'Waiting for Data'
				],
				datasets: [
					{
						label: auditScope.catalogId,
						data: [
							filter(['EVALUATION_STATUS_NOT_COMPLIANT']),
							filter(['EVALUATION_STATUS_NOT_COMPLIANT_MANUALLY']),
							filter(['EVALUATION_STATUS_COMPLIANT']),
							filter(['EVALUATION_STATUS_COMPLIANT_MANUALLY']),
							filter(['EVALUATION_STATUS_PENDING'])
						],
						backgroundColor: ['#991b1b', 'rgb(185 28 28)', '#166534', 'rgb(21 128 61)', '#d4d4d4'],
						hoverOffset: 4
					}
				]
			};
		}
	}

	function filter(status: string[]) {
		return {
			status: status,
			num: Array.from(compliance.values()).filter((value) => status.includes(value)).length
		};
	}
	let data = $derived(buildData(merge));
	run(() => {
		updateChart(data);
	});
</script>

<div class="py-3">
	<div class="relative mb-2 flex items-start">
		<div class="flex h-6 items-center">
			<input
				id="merge-{auditScope.catalogId}"
				aria-describedby="merge-description-{auditScope.catalogId}"
				name="merge-{auditScope.catalogId}"
				type="checkbox"
				class="text-confirmate focus:ring-confirmate h-4 w-4 rounded-sm border-gray-300"
				bind:checked={merge}
			/>
		</div>
		<div class="ml-3 text-sm leading-6">
			<label for="merge-{auditScope.catalogId}" class="font-medium text-gray-900"
				>Merge manual results</label
			>
			<span id="merge-description-{auditScope.catalogId}" class="text-gray-500">
				<span class="sr-only">Merge manual results </span>
				with automatic results.
			</span>
		</div>
	</div>
	<canvas id="chart" bind:this={canvas} class="mr-auto ml-auto h-72 w-72"></canvas>
</div>
