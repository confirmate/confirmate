import { confirmatize } from './util';

export interface StartDiscoveryResponse {
	successful: boolean;
}

export interface HttpEndpoint {
	transportEncryption?: TransportEncryption;
}

export interface TransportEncryption {
	algorithm: string;
	enabled: boolean;
	enforced: boolean;
	tlsVersion: string;
}

export async function startDiscovery(): Promise<boolean> {
	const apiUrl = confirmatize(`/v1/discovery/start`);

	return fetch(apiUrl, {
		method: 'POST',
		headers: {
			Authorization: `Bearer ${localStorage.token}`
		}
	})
		.then((res) => res.json())
		.then((response: StartDiscoveryResponse) => {
			return response.successful;
		});
}
