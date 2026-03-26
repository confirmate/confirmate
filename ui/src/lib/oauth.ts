import { goto } from '$app/navigation';
import { env } from '$env/dynamic/public';
import { type OidcClientSettings, OidcClient } from 'oidc-client-ts';

const settings: OidcClientSettings = {
	authority: env.PUBLIC_OAUTH_AUTHORITY,
	client_id: env.PUBLIC_OAUTH_CLIENT_ID,
	redirect_uri: env.PUBLIC_OAUTH_REDIRECT_URI,
	scope: 'profile'
};

export const client = new OidcClient(settings);

export async function redirectLogin(backTo = '/') {
	client.createSigninRequest({ state: backTo }).then((req) => {
		window.location.href = req.url;
	});
}

export function setToken(token: string) {
	localStorage.setItem('token', token);
}
