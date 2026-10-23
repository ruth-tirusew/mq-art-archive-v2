import adapter from '@sveltejs/adapter-vercel';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
  preprocess: vitePreprocess(),
  kit: {
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
