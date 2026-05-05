import { defineConfig } from 'astro/config';
import svelte from '@astrojs/svelte';
import tailwind from '@astrojs/tailwind';
import sitemap from '@astrojs/sitemap';
import rehypeSlug from 'rehype-slug';
import rehypeAutolinkHeadings from 'rehype-autolink-headings';

// https://astro.build/config
export default defineConfig({
  site: 'https://blog.example.com',
  output: 'static',
  integrations: [
    svelte(),
    tailwind({
      applyBaseStyles: false,
    }),
    sitemap(),
  ],
  image: {
    // Use Astro's built-in sharp service for image optimization
    service: {
      entrypoint: 'astro/assets/services/sharp',
      config: {
        // Default quality for all image formats
        quality: 80,
      },
    },
    // Default responsive widths for srcset generation
    // Images will be generated at these widths during build
    domains: [],
    remotePatterns: [],
  },
  markdown: {
    rehypePlugins: [
      rehypeSlug,
      [
        rehypeAutolinkHeadings,
        {
          behavior: 'prepend',
          properties: {
            className: ['anchor-link'],
            ariaLabel: 'Link to section',
          },
          content: {
            type: 'element',
            tagName: 'span',
            properties: { className: ['anchor-icon'] },
            children: [{ type: 'text', value: '#' }],
          },
        },
      ],
    ],
    shikiConfig: {
      themes: {
        light: 'github-light',
        dark: 'github-dark',
      },
    },
  },
});
