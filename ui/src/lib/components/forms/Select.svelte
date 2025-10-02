<!--
@component Select

A form select dropdown component with label, validation, and help text support.
Provides a consistent interface for selecting from predefined options.
-->
<script lang="ts">
	/**
	 * Option structure for the select component
	 */
	interface Option {
		/** The value to submit with the form */
		value: string;
		/** The display text for the option */
		label: string;
		/** Whether this option is disabled @default false */
		disabled?: boolean;
	}

	/**
	 * Props for the Select component
	 */
	interface Props {
		/** Label text for the select */
		label: string;
		/** Unique identifier for the select */
		id: string;
		options: Option[];
		required?: boolean;
		disabled?: boolean;
		error?: string;
		help?: string;
		value?: string;
		placeholder?: string;
		class?: string;
	}

	let {
		label,
		id,
		options,
		required = false,
		disabled = false,
		error = '',
		help = '',
		value = $bindable(''),
		placeholder = 'Select an option...',
		class: className = ''
	}: Props = $props();
</script>

<div class="space-y-2">
	<label for={id} class="block text-sm font-medium text-gray-700">
		{label}
		{#if required}
			<span class="text-red-500">*</span>
		{/if}
	</label>

	<select
		{id}
		{required}
		{disabled}
		bind:value
		class="block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 disabled:bg-gray-50 disabled:text-gray-500 sm:text-sm {error
			? 'border-red-300 focus:border-red-500 focus:ring-red-500'
			: ''} {className}"
	>
		{#if placeholder}
			<option value="" disabled>{placeholder}</option>
		{/if}
		{#each options as option (option.value)}
			<option value={option.value} disabled={option.disabled}>
				{option.label}
			</option>
		{/each}
	</select>

	{#if error}
		<p class="text-sm text-red-600">{error}</p>
	{:else if help}
		<p class="text-sm text-gray-500">{help}</p>
	{/if}
</div>
