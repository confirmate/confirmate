<script lang="ts">
	import SectionHeader from '$lib/components/ui/SectionHeader.svelte';
	import AccessControl from '$lib/components/ui/AccessControl.svelte';
	import type { PageProps } from './$types';
	import { orchestratorClient } from '$lib/api/client';
	import { invalidate } from '$app/navigation';

	let { data }: PageProps = $props();

	const client = orchestratorClient();

	let selectedUserId = $state('');
	let selectedPermission = $state<'PERMISSION_READER' | 'PERMISSION_CONTRIBUTOR' | 'PERMISSION_ADMIN'>('PERMISSION_CONTRIBUTOR');
	let saving = $state(false);
	let error = $state('');

	async function addUser() {
		if (!selectedUserId) return;
		saving = true;
		error = '';
		const { error: err } = await client.PUT(
			'/v1/users/permissions/{user_permission.object_type}/{user_permission.object_id}/users/{user_permission.user_id}',
			{
				params: {
					path: {
						'user_permission.object_type': 'OBJECT_TYPE_TARGET_OF_EVALUATION',
						'user_permission.object_id': data.toeId,
						'user_permission.user_id': selectedUserId
					}
				},
				body: {
					userId: selectedUserId,
					objectId: data.toeId,
					objectType: 'OBJECT_TYPE_TARGET_OF_EVALUATION',
					permission: selectedPermission
				}
			}
		);
		saving = false;
		if (err) {
			error = (err as { message?: string }).message ?? 'Failed to add user';
		} else {
			selectedUserId = '';
			invalidate((url) => url.pathname.includes('/access'));
		}
	}

	async function removeUser(userId: string) {
		error = '';
		const { error: err } = await client.DELETE(
			'/v1/users/permissions/{object_type}/{object_id}/users/{user_id}',
			{
				params: {
					path: {
						object_type: 'OBJECT_TYPE_TARGET_OF_EVALUATION',
						object_id: data.toeId,
						user_id: userId
					}
				}
			}
		);
		if (err) {
			error = (err as { message?: string }).message ?? 'Failed to remove user';
		} else {
			invalidate((url) => url.pathname.includes('/access'));
		}
	}

	const alreadyAssigned = $derived(
		new Set([
			...data.readers.map((u) => u.id),
			...data.contributors.map((u) => u.id),
			...data.admins.map((u) => u.id)
		])
	);

	const availableUsers = $derived(data.allUsers.filter((u) => !alreadyAssigned.has(u.id)));

	function displayName(u: { firstName?: string; lastName?: string; username?: string }): string {
		const full = [u.firstName, u.lastName].filter(Boolean).join(' ');
		return full || u.username || '';
	}
</script>

<div>
	<SectionHeader
		title="Access"
		description="Users with access to this target of evaluation."
	/>

	<div class="mt-6 max-w-xl space-y-8">
		<AccessControl
			readers={data.readers}
			contributors={data.contributors}
			admins={data.admins}
			onremove={removeUser}
		/>

		<div class="rounded-lg border border-gray-200 bg-white p-4">
			<h4 class="text-sm font-medium text-gray-900">Add user</h4>
			<div class="mt-3 flex items-end gap-3">
				<div class="flex-1">
					<label for="user-select" class="block text-xs font-medium text-gray-600 mb-1">User</label>
					<select
						id="user-select"
						bind:value={selectedUserId}
						class="w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
					>
						<option value="">Select a user…</option>
						{#each availableUsers as user}
							<option value={user.id}>{displayName(user)}</option>
						{/each}
					</select>
				</div>
				<div>
					<label for="perm-select" class="block text-xs font-medium text-gray-600 mb-1">Role</label>
					<select
						id="perm-select"
						bind:value={selectedPermission}
						class="rounded-md border border-gray-300 bg-white px-3 py-2 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
					>
						<option value="PERMISSION_READER">Reader</option>
						<option value="PERMISSION_CONTRIBUTOR">Contributor</option>
						<option value="PERMISSION_ADMIN">Admin</option>
					</select>
				</div>
				<button
					onclick={addUser}
					disabled={!selectedUserId || saving}
					class="rounded-md bg-confirmate px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-confirmate-light disabled:opacity-50 disabled:cursor-not-allowed"
				>
					{saving ? 'Adding…' : 'Add'}
				</button>
			</div>
			{#if error}
				<p class="mt-2 text-xs text-red-600">{error}</p>
			{/if}
		</div>
	</div>
</div>
