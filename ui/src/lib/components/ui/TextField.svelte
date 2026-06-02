<script lang="ts">
	let {
		label,
		type = 'text',
		placeholder = '',
		value = $bindable('')
	}: {
		label: string;
		type?: 'text' | 'email' | 'url';
		placeholder?: string;
		value?: string;
	} = $props();
</script>

<label class="flex flex-col gap-1.5">
	<span class="text-sm font-medium text-gray-700">{label}</span>
	<input
		{type}
		{placeholder}
		bind:value
		class="rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-confirmate focus:outline-none focus:ring-1 focus:ring-confirmate"
	/>
</label>