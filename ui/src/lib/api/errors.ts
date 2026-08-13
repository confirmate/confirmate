import { error } from '@sveltejs/kit';

export function throwError(response: Response) {
	if (!response.ok) {
		throw error(response.status, response.statusText);
	}
	return response;
}
