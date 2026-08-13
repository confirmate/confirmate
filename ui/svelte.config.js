import adapter from '@sveltejs/adapter-static';
import { dirname, relative, sep } from 'node:path';
import { fileURLToPath } from 'node:url';

// import.meta.dirname requires Node 20.11+; this repo's Dockerfiles still use Node 18 images,
// so derive the directory from import.meta.url instead for broader compatibility.
const currentDir = dirname(fileURLToPath(import.meta.url));

/** @type {import('@sveltejs/kit').Config} */
const config = {
	compilerOptions: {
		// defaults to rune mode for the project, except for `node_modules`. Can be removed in svelte 6.
		runes: ({ filename }) => {
			const relativePath = relative(currentDir, filename);
			const pathSegments = relativePath.toLowerCase().split(sep);
			const isExternalLibrary = pathSegments.includes('node_modules');

			return isExternalLibrary ? undefined : true;
		}
	},
	kit: {
		adapter: adapter({
			fallback: 'index.html'
		})
	}
};

export default config;
