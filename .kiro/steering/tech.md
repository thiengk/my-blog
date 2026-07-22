# Tech Stack

## Frontend (`frontend/`)

- **Framework**: Astro 5 (static site generation)
- **UI Components**: Svelte 5 (interactive islands)
- **CSS**: TailwindCSS 3.4
- **AI/ML**: Transformers.js (Xenova) — chạy Whisper model trên browser
- **Build output**: Static HTML (Cloudflare Pages)

## Infrastructure

- **Frontend hosting**: Cloudflare Pages (push to main = auto deploy)

## Commands

### Frontend

```bash
cd frontend
npm install          # Cài dependencies
npm run dev          # Dev server (localhost:4321)
npm run build        # Build static site
npm run preview      # Preview build
```
