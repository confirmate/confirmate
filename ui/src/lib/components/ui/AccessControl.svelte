<script lang="ts">
	import type { SchemaUser } from '$lib/api/openapi/orchestrator';
	import UserAvatar from './UserAvatar.svelte';
	import { Trash } from '@steeze-ui/heroicons';
	import { Icon } from '@steeze-ui/svelte-icon';

	let {
		readers = [],
		contributors = [],
		admins = [],
		onremove
	}: {
		readers?: SchemaUser[];
		contributors?: SchemaUser[];
		admins?: SchemaUser[];
		onremove?: (userId: string) => void;
	} = $props();

	let removing = $state<string | null>(null);

	async function handleRemove(userId: string) {
		if (removing) return;
		removing = userId;
		try {
			await onremove?.(userId);
		} finally {
			removing = null;
		}
	}

	const sections = $derived([
		{ label: 'Admins', description: 'Can manage permissions and delete this resource.', users: admins },
		{ label: 'Contributors', description: 'Can modify settings and data.', users: contributors },
		{ label: 'Readers', description: 'Can view this resource.', users: readers }
	]);
</script>

<div class="space-y-6">
	{#each sections as section}
		<div>
			<div class="flex items-baseline gap-2">
				<h4 class="text-sm font-medium text-gray-900">{section.label}</h4>
				<span class="text-xs text-gray-400">{section.description}</span>
			</div>
			<div class="mt-3">
				{#if section.users.length > 0}
					<ul class="space-y-2">
						{#each section.users as user}
							<li class="flex items-center justify-between rounded-md border border-gray-100 bg-gray-50 px-3 py-2">
								<UserAvatar {user} />
								{#each user.roles ?? [] as role}
									<span class="ml-auto shrink-0 rounded-full bg-blue-50 px-2 py-0.5 text-xs font-medium text-blue-700">
										{role.replace('ROLE_', '').replaceAll('_', ' ').toLowerCase()}
									</span>
								{/each}
								{#if onremove}
									<button
										type="button"
										onclick={() => handleRemove(user.id)}
										disabled={removing === user.id}
										title="Remove access"
										class="ml-2 shrink-0 rounded p-1 text-gray-300 hover:bg-red-50 hover:text-red-500 disabled:opacity-30"
									>
										<Icon src={Trash} class="h-4 w-4" />
									</button>
								{/if}
							</li>
						{/each}
					</ul>
				{:else}
					<p class="text-sm text-gray-400 italic">No {section.label.toLowerCase()} assigned.</p>
				{/if}
			</div>
		</div>
	{/each}

</div>
