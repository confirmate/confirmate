<script lang="ts" module>
	export interface AddEvaluationResultEvent {
		name: string;
		comment: string;
		validUntil: Date;
		status: ComplianceStatus;
	}
</script>

<script lang="ts">
	import type { ComplianceStatus } from '$lib/api/evaluation';
	import ComplianceStatusSelect from '$lib/components/ComplianceStatusSelect.svelte';
	import { createDialog } from 'svelte-headlessui';
	import { Transition } from 'svelte-transition';

	interface Props {
		open?: boolean;
		addResult: (event: AddEvaluationResultEvent) => void;
	}

	let { open = $bindable(false), addResult }: Props = $props();
	let name = $state('');
	let comment = $state('');
	function submit() {
		let date = new Date();
		date.setDate(date.getDate() + 30);
		addResult({ comment: comment, name: name, validUntil: date, status: status });
		open = false;
	}

	let status: ComplianceStatus = $state('EVALUATION_STATUS_COMPLIANT_MANUALLY');
	const dialog = createDialog({ label: 'Add Evaluation Result', expanded: open });
</script>

<Transition show={$dialog.expanded}>
	<div class="relative z-10">
		<Transition
			enter="ease-out duration-300"
			enterFrom="opacity-0"
			enterTo="opacity-100"
			leave="ease-in duration-200"
			leaveFrom="opacity-100"
			leaveTo="opacity-0"
		>
			<div class="bg-opacity-75 fixed inset-0 bg-gray-500 transition-opacity"></div>
		</Transition>
		<div class="fixed inset-0 z-10 overflow-y-auto">
			<div class="flex min-h-full items-end justify-center p-4 text-center sm:items-center sm:p-0">
				<Transition
					enter="ease-out duration-300"
					enterFrom="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
					enterTo="opacity-100 translate-y-0 sm:scale-100"
					leave="ease-in duration-200"
					leaveFrom="opacity-100 translate-y-0 sm:scale-100"
					leaveTo="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
				>
					<div
						class="relative transform overflow-hidden rounded-lg bg-white px-4 pt-5 pb-4 text-left shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-lg sm:p-6"
					>
						<div>
							<div class="space-y-12">
								<div class="">
									<div class="text-base leading-7 font-semibold text-gray-900">
										Provide a manual evaluation result
									</div>
									<p class="mt-1 text-sm leading-6 text-gray-600">
										Using this form, you can provide a manual evaluation result that will be
										considered as <ComplianceStatusSelect bind:status />
									</p>
									<div class="mt-2 grid grid-cols-1 gap-x-6 gap-y-8 sm:grid-cols-6">
										<div class="sm:col-span-4">
											<label for="name" class="block text-sm leading-6 font-medium text-gray-900">
												Name
											</label>
											<div class="mt-2">
												<input
													type="text"
													name="name"
													id="name"
													autocomplete="name"
													class="focus:ring-confirmate block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-xs ring-1 ring-gray-300 ring-inset placeholder:text-gray-400 focus:ring-2 focus:ring-inset sm:text-sm sm:leading-6"
													bind:value={name}
												/>
											</div>
										</div>

										<div class="col-span-full">
											<label for="about" class="block text-sm leading-6 font-medium text-gray-900">
												Comment
											</label>
											<div class="mt-2">
												<textarea
													id="about"
													name="about"
													rows="3"
													class="focus:ring-confirmate block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-xs ring-1 ring-gray-300 ring-inset placeholder:text-gray-400 focus:ring-2 focus:ring-inset sm:text-sm sm:leading-6"
													bind:value={comment}
												></textarea>
											</div>
											<p class="mt-3 text-sm leading-6 text-gray-600">
												Please consider that an auditor will judge the fulfillment of this control
												based on your comment.
											</p>
										</div>

										<div class="sm:col-span-4">
											<label
												for="validity"
												class="block text-sm leading-6 font-medium text-gray-900">Validity</label
											>
											<div class="mt-2">
												<input
													type="text"
													name="first-name"
													id="first-name"
													autocomplete="given-name"
													class="focus:ring-confirmate block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-xs ring-1 ring-gray-300 ring-inset placeholder:text-gray-400 focus:ring-2 focus:ring-inset sm:text-sm sm:leading-6"
													value="1 month"
													disabled
												/>
											</div>
										</div>
									</div>
									<div class="mt-5 sm:mt-6 sm:grid sm:grid-flow-row-dense sm:grid-cols-2 sm:gap-3">
										<button
											type="button"
											class="bg-confirmate hover:bg-confirmate-light focus-visible:outline-confirmate inline-flex w-full justify-center rounded-md px-3 py-2 text-sm font-semibold text-white shadow-xs focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-solid sm:col-start-2"
											onclick={submit}
										>
											Save
										</button>
										<button
											type="button"
											class="mt-3 inline-flex w-full justify-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-xs ring-1 ring-gray-300 ring-inset hover:bg-gray-50 sm:col-start-1 sm:mt-0"
											onclick={dialog.close}
										>
											Cancel
										</button>
									</div>
								</div>
							</div>
						</div>
					</div>
				</Transition>
			</div>
		</div>
	</div>
</Transition>
