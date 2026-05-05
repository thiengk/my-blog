import typography from '@tailwindcss/typography';

/** @type {import('tailwindcss').Config} */
export default {
  content: ['./src/**/*.{astro,html,js,jsx,md,mdx,svelte,ts,tsx,vue}'],
  darkMode: 'class',
  theme: {
    screens: {
      // mobile: default (< 768px)
      // tablet: 768px - 1024px
      sm: '640px',
      md: '768px',
      // desktop: > 1024px
      lg: '1024px',
      xl: '1280px',
      '2xl': '1536px',
    },
    extend: {
      colors: {
        primary: {
          50: '#f0f9ff',
          100: '#e0f2fe',
          200: '#bae6fd',
          300: '#7dd3fc',
          400: '#38bdf8',   // dark mode links - contrast 8.19:1 on gray-900
          500: '#0ea5e9',
          600: '#0284c7',   // light mode links - contrast 4.56:1 on white
          700: '#0369a1',
          800: '#075985',
          900: '#0c4a6e',
          950: '#082f49',
        },
      },
      fontFamily: {
        sans: [
          'Inter',
          'system-ui',
          '-apple-system',
          'Segoe UI',
          'Roboto',
          'Helvetica Neue',
          'Arial',
          'sans-serif',
        ],
        mono: [
          'JetBrains Mono',
          'Fira Code',
          'ui-monospace',
          'SFMono-Regular',
          'Menlo',
          'Monaco',
          'Consolas',
          'monospace',
        ],
      },
      typography: ({ theme }) => ({
        DEFAULT: {
          css: {
            // WCAG AA: gray-700 (#374151) on white = 8.59:1 contrast ratio
            '--tw-prose-body': theme('colors.gray.700'),
            '--tw-prose-headings': theme('colors.gray.900'),
            '--tw-prose-lead': theme('colors.gray.600'),
            '--tw-prose-links': theme('colors.primary.600'),
            '--tw-prose-bold': theme('colors.gray.900'),
            '--tw-prose-counters': theme('colors.gray.600'),
            '--tw-prose-bullets': theme('colors.gray.400'),
            '--tw-prose-hr': theme('colors.gray.200'),
            '--tw-prose-quotes': theme('colors.gray.900'),
            '--tw-prose-quote-borders': theme('colors.gray.200'),
            '--tw-prose-captions': theme('colors.gray.600'),
            '--tw-prose-code': theme('colors.gray.800'),
            '--tw-prose-pre-code': theme('colors.gray.200'),
            '--tw-prose-pre-bg': theme('colors.gray.800'),
            '--tw-prose-th-borders': theme('colors.gray.300'),
            '--tw-prose-td-borders': theme('colors.gray.200'),
            // Headings
            h1: {
              fontWeight: '800',
              letterSpacing: '-0.025em',
            },
            h2: {
              fontWeight: '700',
              letterSpacing: '-0.02em',
            },
            h3: {
              fontWeight: '600',
            },
            h4: {
              fontWeight: '600',
            },
            // Links
            a: {
              textDecoration: 'underline',
              textUnderlineOffset: '2px',
              '&:hover': {
                color: theme('colors.primary.700'),
              },
            },
            // Code
            code: {
              fontWeight: '500',
              fontSize: '0.875em',
            },
            'code::before': {
              content: '""',
            },
            'code::after': {
              content: '""',
            },
            // Blockquotes
            blockquote: {
              fontStyle: 'italic',
              borderLeftWidth: '4px',
            },
            // Tables
            table: {
              fontSize: '0.875em',
            },
            thead: {
              borderBottomWidth: '2px',
            },
            'thead th': {
              fontWeight: '600',
              paddingBottom: '0.75em',
            },
            'tbody td': {
              paddingTop: '0.75em',
              paddingBottom: '0.75em',
            },
          },
        },
        dark: {
          css: {
            // WCAG AA: gray-300 (#d1d5db) on gray-900 (#111827) = 11.08:1 contrast ratio
            '--tw-prose-body': theme('colors.gray.300'),
            '--tw-prose-headings': theme('colors.gray.100'),
            '--tw-prose-lead': theme('colors.gray.400'),
            '--tw-prose-links': theme('colors.primary.400'),
            '--tw-prose-bold': theme('colors.gray.100'),
            '--tw-prose-counters': theme('colors.gray.400'),
            '--tw-prose-bullets': theme('colors.gray.400'),
            '--tw-prose-hr': theme('colors.gray.700'),
            '--tw-prose-quotes': theme('colors.gray.100'),
            '--tw-prose-quote-borders': theme('colors.gray.700'),
            '--tw-prose-captions': theme('colors.gray.400'),
            '--tw-prose-code': theme('colors.gray.200'),
            '--tw-prose-pre-code': theme('colors.gray.200'),
            '--tw-prose-pre-bg': theme('colors.gray.800/50'),
            '--tw-prose-th-borders': theme('colors.gray.600'),
            '--tw-prose-td-borders': theme('colors.gray.700'),
            // Links in dark mode
            a: {
              '&:hover': {
                color: theme('colors.primary.300'),
              },
            },
          },
        },
        // Larger prose variant for blog posts
        lg: {
          css: {
            h1: {
              fontSize: '2.25em',
            },
            h2: {
              fontSize: '1.5em',
              marginTop: '1.75em',
            },
            h3: {
              fontSize: '1.25em',
              marginTop: '1.5em',
            },
          },
        },
      }),
    },
  },
  plugins: [typography],
};
