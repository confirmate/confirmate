import { env } from '$env/dynamic/public';

export function trim(s: string, maxLength = 200): string {
	if (s.length < maxLength) {
		return s;
	} else {
		return s.substring(0, maxLength) + '[...]';
	}
}

export function short(resourceID: string) {
	// Split resource by / and take the last index
	const rr = resourceID.split('/');

	return rr[rr.length - 1];
}

/**
 * Combines the relative URL resource with the Confirmate backend URL.
 *
 * @param url the relative URL resource
 */
export function confirmatize(url: string): string {
	let backend = env.PUBLIC_CONFIRMATE_URL;
	if (backend?.endsWith('/')) {
		backend = backend.substring(0, backend.length);
	}

	if (!url.startsWith('/')) {
		url = `/${url}`;
	}

	return backend + url;
}
