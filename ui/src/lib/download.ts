/**
 * Fetches `url` with the given headers and triggers a browser download of the raw response body.
 * Used for binary file downloads (e.g. report exports) that return the file directly rather than
 * being wrapped in a JSON envelope, so the generated openapi clients (which always parse JSON)
 * can't be used.
 */
export async function downloadFile(url: string, headers: HeadersInit, fallbackFilename: string) {
	const response = await fetch(url, { headers });
	if (!response.ok) {
		throw new Error(`request to ${url} failed with status ${response.status}`);
	}

	const filename = filenameFromContentDisposition(response.headers.get('Content-Disposition')) ?? fallbackFilename;
	const blob = await response.blob();

	const blobUrl = URL.createObjectURL(blob);
	const a = document.createElement('a');
	a.href = blobUrl;
	a.download = filename;
	document.body.appendChild(a);
	a.click();
	a.remove();
	URL.revokeObjectURL(blobUrl);
}

function filenameFromContentDisposition(header: string | null): string | undefined {
	return header?.match(/filename="?([^";]+)"?/)?.[1];
}
