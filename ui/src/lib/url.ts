/**
 * Returns the correct href for a given path, accounting for hash routing.
 * With `router.type === 'hash'`, all internal links need a `#` prefix.
 */
export function href(path: string): string {
	return `#${path}`;
}
