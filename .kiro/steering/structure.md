# Project Structure

```
.kiro/
  steering/                    # AI steering rules

frontend/                      # Astro + Svelte + TailwindCSS (static site)
  public/
    favicon.svg
  src/
    components/
      AudioTranscriber.svelte  # Main component: upload MP3 → transcribe → show text
    layouts/
      BaseLayout.astro         # Main layout (header, main, footer, dark mode)
    pages/
      index.astro              # Homepage (single page app)
    styles/
      global.css               # TailwindCSS base + component styles
  astro.config.mjs
  tailwind.config.mjs
  package.json
```
