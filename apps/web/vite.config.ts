import { fileURLToPath } from 'node:url';
import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [tailwindcss(), sveltekit()],
	resolve: {
		dedupe: ['svelte'],
		// packages/ui sits outside this app's node_modules, so its bare imports can't find these.
		// Pin them to the ESM entry (a kit.alias or default resolve picks the CJS build in dev).
		alias: [
			{ find: /^clsx$/, replacement: fileURLToPath(new URL('./node_modules/clsx/dist/clsx.mjs', import.meta.url)) },
			{
				find: /^tailwind-merge$/,
				replacement: fileURLToPath(new URL('./node_modules/tailwind-merge/dist/bundle-mjs.mjs', import.meta.url))
			}
		]
	},
	server: {
		fs: { allow: ['../..'] },
		port: 8082,
		strictPort: false
	}
});
