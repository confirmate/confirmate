// Used only for the authorize redirect (full navigation, not fetch)
const AUTHORIZE_BASE = import.meta.env.VITE_AUTH_BASE ?? '';
const CLIENT_ID = 'ui';
const RETURN_TO_KEY = 'auth_return_to';
const CODE_VERIFIER_KEY = 'auth_code_verifier';
const TOKEN_KEY = 'token';

function redirectUri(): string {
	return `${window.location.origin}/auth/callback`;
}

async function generatePKCE(): Promise<{ verifier: string; challenge: string }> {
	const array = new Uint8Array(32);
	crypto.getRandomValues(array);
	const verifier = btoa(String.fromCharCode(...array))
		.replace(/\+/g, '-')
		.replace(/\//g, '_')
		.replace(/=/g, '');

	const encoder = new TextEncoder();
	const data = encoder.encode(verifier);
	const digest = await crypto.subtle.digest('SHA-256', data);
	const challenge = btoa(String.fromCharCode(...new Uint8Array(digest)))
		.replace(/\+/g, '-')
		.replace(/\//g, '_')
		.replace(/=/g, '');

	return { verifier, challenge };
}

export function getToken(): string | null {
	if (typeof localStorage === 'undefined') return null;
	return localStorage.getItem(TOKEN_KEY);
}

export function setToken(token: string): void {
	localStorage.setItem(TOKEN_KEY, token);
}

export function clearToken(): void {
	localStorage.removeItem(TOKEN_KEY);
}

export function isAuthenticated(): boolean {
	return !!getToken();
}

export async function login(returnTo = '/dashboard/'): Promise<void> {
	sessionStorage.setItem(RETURN_TO_KEY, returnTo);

	const { verifier, challenge } = await generatePKCE();
	sessionStorage.setItem(CODE_VERIFIER_KEY, verifier);

	const params = new URLSearchParams({
		response_type: 'code',
		client_id: CLIENT_ID,
		redirect_uri: redirectUri(),
		code_challenge: challenge,
		code_challenge_method: 'S256'
	});

	window.location.href = `${AUTHORIZE_BASE}/v1/auth/authorize?${params}`;
}

export function logout(): void {
	clearToken();
	sessionStorage.removeItem(CODE_VERIFIER_KEY);
	sessionStorage.removeItem(RETURN_TO_KEY);
	const returnTo = encodeURIComponent(window.location.origin + '/');
	window.location.replace(`${AUTHORIZE_BASE}/v1/auth/logout?return_to=${returnTo}`);
}

export async function exchangeCode(code: string): Promise<void> {
	const verifier = sessionStorage.getItem(CODE_VERIFIER_KEY);
	if (!verifier) throw new Error('No PKCE code verifier found');

	const body = new URLSearchParams({
		grant_type: 'authorization_code',
		code,
		client_id: CLIENT_ID,
		redirect_uri: redirectUri(),
		code_verifier: verifier
	});

	// Use relative URL so the request goes through the Vite proxy (avoids CORS)
	const res = await fetch(`/v1/auth/token`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
		body: body.toString()
	});

	if (!res.ok) {
		throw new Error(`Token exchange failed: ${res.status} ${await res.text()}`);
	}

	const data = await res.json();
	if (!data.access_token) throw new Error('No access_token in response');
	sessionStorage.removeItem(CODE_VERIFIER_KEY);
	setToken(data.access_token);
}

export function getReturnTo(): string {
	return sessionStorage.getItem(RETURN_TO_KEY) ?? '/dashboard/';
}

export function clearReturnTo(): void {
	sessionStorage.removeItem(RETURN_TO_KEY);
}
