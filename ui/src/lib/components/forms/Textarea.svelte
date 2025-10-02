<!--
@component Textarea

A multi-line text input component with label, validation, and help text support.
Similar to Input but for longer text content.
-->
<script lang="ts">
	import type { Snippet } from 'svelte';

	/**
	 * Props for the Textarea component
	 */
	interface Props {
		/** Label text for the textarea */
		label: string;
		/** Unique identifier for the textarea */
		id: string;
		/** Number of visible text lines @default 3 */
		rows?: number;
		/** Placeholder text */
		placeholder?: string;
		/** Whether the textarea is required @default false */
		required?: boolean;
		/** Whether the textarea is disabled @default false */
		disabled?: boolean;
		/** Error message to display */
		error?: string;
		/** Help text to display below the textarea */
		help?: string;
		/** Textarea value (bindable) */
		value?: string;
		/** Additional CSS classes */
		class?: string;
	}

	let {
		label,
		id,
		rows = 3,
		placeholder = '',
		required = false,
		disabled = false,
		error = '',
		help = '',
		value = $bindable(''),
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

	<textarea
		{id}
		{rows}
		{placeholder}
		{required}
		{disabled}
		bind:value
		class="block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 disabled:bg-gray-50 disabled:text-gray-500 sm:text-sm {error
			? 'border-red-300 focus:border-red-500 focus:ring-red-500'
			: ''} {className}"
	></textarea>

	{#if error}
		<p class="text-sm text-red-600">{error}</p>
	{:else if help}
		<p class="text-sm text-gray-500">{help}</p>
	{/if}
</div>
