<!--
@component Button

A versatile button component with multiple variants and sizes.
Supports primary, secondary, danger, and ghost variants in small, medium, and large sizes.
-->
<script lang="ts">
	import type { Snippet } from 'svelte';

	/**
	 * Props for the Button component
	 */
	interface Props {
		/** Content to render inside the button */
		children: Snippet;
		/** Visual style variant of the button @default 'primary' */
		variant?: 'primary' | 'secondary' | 'danger' | 'ghost';
		/** Size of the button @default 'md' */
		size?: 'sm' | 'md' | 'lg';
		/** Whether the button is disabled @default false */
		disabled?: boolean;
		/** HTML button type @default 'button' */
		type?: 'button' | 'submit' | 'reset';
		/** Additional CSS classes to apply */
		class?: string;
		/** Click event handler */
		onclick?: (event: MouseEvent) => void;
	}

	let {
		children,
		variant = 'primary',
		size = 'md',
		disabled = false,
		type = 'button',
		class: className = '',
		onclick
	}: Props = $props();

	const baseClasses =
		'inline-flex items-center justify-center rounded-md font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50';

	const variants = {
		primary: 'bg-blue-600 text-white hover:bg-blue-700 focus-visible:ring-blue-500',
		secondary: 'bg-gray-100 text-gray-900 hover:bg-gray-200 focus-visible:ring-gray-500',
		danger: 'bg-red-600 text-white hover:bg-red-700 focus-visible:ring-red-500',
		ghost: 'hover:bg-gray-100 hover:text-gray-900 focus-visible:ring-gray-500'
	};

	const sizes = {
		sm: 'h-8 px-3 text-sm',
		md: 'h-9 px-4 py-2',
		lg: 'h-10 px-8'
	};
</script>

<button
	{type}
	{disabled}
	{onclick}
	class="{baseClasses} {variants[variant]} {sizes[size]} {className}"
>
	{@render children()}
</button>
