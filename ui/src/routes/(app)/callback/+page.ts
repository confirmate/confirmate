import { client, setToken } from '$lib/oauth';
import { ErrorResponse } from 'oidc-client-ts';
import type { PageLoad } from './$types';
import { error, redirect } from '@sveltejs/kit';

export const load = (async ({ url }) => {
	let res;
	try {
		res = await client.processSigninResponse(url.toString());
	} catch (err) {
		if (err instanceof ErrorResponse) {
			error(400, `error while fetching OAuth 2.0 response: ${err.error_description}`);
		} else {
			console.log(err);
			error(400, 'could not complete OAuth 2.0 flow');
		}
	}

	console.log('Successfully signed in');
	setToken(res.access_token);
	redirect(301, (res.userState as string) ?? '/');
}) satisfies PageLoad;
