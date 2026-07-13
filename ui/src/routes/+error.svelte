<script lang="ts">
	import { page } from '$app/state';
</script>

<div class="flex min-h-screen items-center justify-center bg-gray-50">
	<div class="text-center">
		<p class="text-6xl font-bold text-confirmate">{page.status}</p>
		<h1 class="mt-4 text-xl font-semibold text-gray-900">
			{page.status === 404 ? 'Page not found' : 'Something went wrong'}
		</h1>
		<p class="mt-2 text-sm text-gray-500">
			{page.error?.message ?? 'An unexpected error occurred.'}
		</p>
		<a
			href="/"
			class="mt-6 inline-block rounded-md bg-confirmate px-4 py-2 text-sm font-medium text-white hover:bg-confirmate-light"
		>
			Back to Dashboard
		</a>
	</div>
</div>
