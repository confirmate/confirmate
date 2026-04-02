<script lang="ts">
	import UserAvatar from '$lib/components/ui/UserAvatar.svelte';
	import SectionHeader from '$lib/components/ui/SectionHeader.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
</script>

<div>
	<SectionHeader title="Users" description="All users with access to this Confirmate instance." />

	<div class="mt-6">
		{#if data.users.length > 0}
			<div class="overflow-hidden rounded-lg border border-gray-200 bg-white">
				<table class="min-w-full divide-y divide-gray-200">
					<thead class="bg-gray-50">
						<tr>
							<th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wide text-gray-500">User</th>
							<th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wide text-gray-500">Roles</th>
							<th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wide text-gray-500">Status</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-100">
						{#each data.users as user}
							<tr class="hover:bg-gray-50">
								<td class="px-4 py-3">
									<UserAvatar {user} />
								</td>
								<td class="px-4 py-3">
									<div class="flex flex-wrap gap-1">
										{#each user.roles ?? [] as role}
											<span class="rounded-full bg-blue-50 px-2 py-0.5 text-xs font-medium text-blue-700">
												{role.replace('ROLE_', '').replaceAll('_', ' ').toLowerCase()}
											</span>
										{/each}
									</div>
								</td>
								<td class="px-4 py-3">
									{#if user.enabled}
										<span class="inline-flex items-center rounded-full bg-green-50 px-2 py-0.5 text-xs font-medium text-green-700">Active</span>
									{:else}
										<span class="inline-flex items-center rounded-full bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-500">Disabled</span>
									{/if}
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{:else}
			<EmptyState title="No users found" description="Users will appear here once they log in." />
		{/if}
	</div>
</div>
