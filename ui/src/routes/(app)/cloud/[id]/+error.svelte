<script>
	import { redirectLogin } from '$lib/oauth';
	import { page } from '$app/stores';

	// We want to automatically redirect to the login page if we get an access
	// denied from the backend
	page.subscribe((p) => {
		if (p.status == 401) {
			redirectLogin('/');
		}
	});
</script>

{#if $page.status == 404}
	<p>Whatever you tried to find does not exist</p>
{:else}
	Seems like you encountered an error! The error was:
	<ul>
		<li>{$page.status}</li>
		<li>{$page.error?.message}</li>
	</ul>
{/if}
