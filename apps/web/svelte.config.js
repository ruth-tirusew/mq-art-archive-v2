import adapter from '@sveltejs/adapter-vercel';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';
import { loadEnv } from 'vite';

// Origin of the API, the one cross-origin place this app is allowed to talk to (CSP connect-src).
const publicEnv = { ...loadEnv('', process.cwd(), 'PUBLIC_'), ...process.env };
let apiOrigin = null;
try {
  apiOrigin = new URL(publicEnv.PUBLIC_API_URL).origin;
} catch {
  // not configured: fall back to any https origin below rather than breaking the app
}

/**
 * Content Security Policy. SvelteKit adds nonces to its own scripts (mode: 'auto'), so no inline
 * script is allowed and an injected <script> would not run. Only enforced for production builds;
 * the dev server needs inline scripts and websockets for hot reload.
 */
const csp =
  process.env.NODE_ENV === 'production'
    ? {
        mode: 'auto',
        directives: {
          'default-src': ['self'],
          'script-src': ['self'],
          'style-src': ['self', 'unsafe-inline', 'https://fonts.googleapis.com'],
          'font-src': ['self', 'https://fonts.gstatic.com'],
          // Artists' and events' images come from many https hosts (Cloudinary, event sources).
          'img-src': ['self', 'data:', 'blob:', 'https:'],
          'connect-src': ['self', apiOrigin ?? 'https:', 'https://api.cloudinary.com'],
          'frame-ancestors': ['none'],
          'base-uri': ['self'],
          'form-action': ['self'],
          'object-src': ['none']
        }
      }
    : undefined;


/** @type {import('@sveltejs/kit').Config} */
const config = {
	compilerOptions: {
		runes: ({ filename }) =>
			filename.split(/[/\\]/).includes('node_modules') ? undefined : true
	},
	preprocess: vitePreprocess(),
	kit: {
		csp,
		alias: {
			$ui: '../../packages/ui/src',
			// packages/ui sits outside this app's node_modules; this keeps its type-only svelte imports resolving.
			'svelte/elements': './node_modules/svelte/elements'
		},
		adapter: adapter({
			runtime: 'nodejs22.x'
		})
	}
};

export default config;
