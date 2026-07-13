import { orchestratorClient } from '$lib/api/client';
import type { PageLoad } from './$types';
import type { SchemaUser } from '$lib/api/openapi/orchestrator';

export const load = (async ({ parent, fetch, params }) => {
	const { toe } = await parent();
	const client = orchestratorClient(fetch);

	const [{ data: permsData }, { data: usersData }] = await Promise.all([
		client.GET('/v1/users/permissions', {
			params: {
				query: {
					'filter.objectType': 'OBJECT_TYPE_TARGET_OF_EVALUATION',
					'filter.objectId': params.id
				}
			}
		}),
		client.GET('/v1/users')
	]);

	const perms = permsData?.userPermissions ?? [];
	const allUsers = usersData?.users ?? [];

	const userById = new Map<string, SchemaUser>();
	for (const u of allUsers) {
		if (u.id) userById.set(u.id, u);
	}

	const collect = (level: string): SchemaUser[] =>
		perms
			.filter((p) => p.permission === level)
			.map((p) => userById.get(p.userId))
			.filter((u): u is SchemaUser => !!u);

	return {
		toe,
		toeId: params.id,
		allUsers,
		readers: collect('PERMISSION_READER'),
		contributors: collect('PERMISSION_CONTRIBUTOR'),
		admins: collect('PERMISSION_ADMIN')
	};
}) satisfies PageLoad;
