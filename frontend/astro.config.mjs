import { defineConfig } from 'astro/config';
import svelte from '@astrojs/svelte';
import tailwind from '@astrojs/tailwind';

import cloudflare from "@astrojs/cloudflare";

export default defineConfig({
  site: 'https://speech-to-text.pages.dev',
  output: 'static',

  integrations: [
    svelte(),
    tailwind({
      applyBaseStyles: false,
    }),
  ],

  vite: {
    ssr: {
      external: ['@xenova/transformers'],
    },
    optimizeDeps: {
      exclude: ['@xenova/transformers'],
    },
  },

  adapter: cloudflare()
});