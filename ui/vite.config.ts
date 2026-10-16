import adapter from '@sveltejs/adapter-static';
import { dirname, relative, sep } from 'node:path';
import { fileURLToPath } from 'node:url';
import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';

// import.meta.dirname requires Node 20.11+; this repo's Dockerfiles still use Node 18 images,
// so derive the directory from import.meta.url instead for broader compatibility.
const currentDir = dirname(fileURLToPath(import.meta.url));

export default defineConfig({
	plugins: [
		tailwindcss(),
		sveltekit({
			compilerOptions: {
				// defaults to rune mode for the project, except for `node_modules`. Can be removed in svelte 6.
				runes: ({ filename }) => {
					const relativePath = relative(currentDir, filename);
					const pathSegments = relativePath.toLowerCase().split(sep);
					const isExternalLibrary = pathSegments.includes('node_modules');

					return isExternalLibrary ? undefined : true;
				}
			},
			adapter: adapter({ fallback: 'index.html' }),
			paths: {
				// Optional path prefix for serving the UI behind a path-based reverse proxy
				// (e.g. UI_BASE_PATH=/proxy/5173). Must start with a slash and not end with one.
				base: (process.env.UI_BASE_PATH ?? '') as '' | `/${string}`
			}
		})
	],
	server: {
		proxy: {
			'/v1/auth/': 'http://localhost:8080',
			'/v1/discovery/': 'http://localhost:8080',
			'/v1experimental/evidence/': 'http://localhost:8080',
			'/v1/evidence_store/': 'http://localhost:8080',
			'/v1/assessment/': 'http://localhost:8080',
			'/v1/evaluation/': 'http://localhost:8080',
			'/v1/orchestrator/': 'http://localhost:8080',
			'/v1/users': 'http://localhost:8080'
		}
	}
});
