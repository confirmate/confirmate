<!--
@component Input

A form input component with label, validation, and help text support.
Supports various input types, validation states, and custom content via snippets.
-->
<script lang="ts">
	import type { Snippet } from 'svelte';

	/**
	 * Props for the Input component
	 */
	interface Props {
		/** Label text for the input */
		label: string;
		/** Unique identifier for the input */
		id: string;
		/** HTML input type @default 'text' */
		type?: 'text' | 'email' | 'password' | 'number' | 'url';
		/** Placeholder text */
		placeholder?: string;
		/** Whether the input is required @default false */
		required?: boolean;
		/** Whether the input is disabled @default false */
		disabled?: boolean;
		/** Error message to display */
		error?: string;
		/** Help text to display below the input */
		help?: string;
		/** Input value (bindable) */
		value?: string | number;
		/** Additional CSS classes */
		class?: string;
		/** Optional content to render in the input (e.g., icons) */
		children?: Snippet;
	}

	let {
		label,
		id,
		type = 'text',
		placeholder = '',
		required = false,
		disabled = false,
		error = '',
		help = '',
		value = $bindable(''),
		class: className = '',
		children
	}: Props = $props();
</script>

<div class="space-y-2">
	<label for={id} class="block text-sm font-medium text-gray-700">
		{label}
		{#if required}
			<span class="text-red-500">*</span>
		{/if}
	</label>

	<div class="relative">
		<input
			{id}
			{type}
			{placeholder}
			{required}
			{disabled}
			bind:value
			class="block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 disabled:bg-gray-50 disabled:text-gray-500 sm:text-sm {error
				? 'border-red-300 focus:border-red-500 focus:ring-red-500'
				: ''} {className}"
		/>

		{#if children}
			<div class="absolute inset-y-0 right-0 flex items-center pr-3">
				{@render children()}
			</div>
		{/if}
	</div>

	{#if error}
		<p class="text-sm text-red-600">{error}</p>
	{:else if help}
		<p class="text-sm text-gray-500">{help}</p>
	{/if}
</div>
