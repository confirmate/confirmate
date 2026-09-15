/** Triggers a browser download of base64-encoded file content (e.g. from a protobuf `bytes` field). */
export function downloadBase64File(base64: string, filename: string, mimeType: string) {
	const binary = atob(base64);
	const bytes = new Uint8Array(binary.length);
	for (let i = 0; i < binary.length; i++) {
		bytes[i] = binary.charCodeAt(i);
	}

	const url = URL.createObjectURL(new Blob([bytes], { type: mimeType }));
	const a = document.createElement('a');
	a.href = url;
	a.download = filename;
	document.body.appendChild(a);
	a.click();
	a.remove();
	URL.revokeObjectURL(url);
}
