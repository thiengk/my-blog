# Tech Stack

## Frontend (`frontend/`)

- **Framework**: Astro 5 (static site generation)
- **UI Components**: Svelte 5 (interactive islands)
- **CSS**: TailwindCSS 3.4 + @tailwindcss/typography
- **Search**: Fuse.js 7 (client-side fuzzy search)
- **Markdown**: rehype-slug, rehype-autolink-headings, Shiki (syntax highlighting)
- **Build output**: Static HTML (Cloudflare Pages)

## Backend (`backend/`)

- **Language**: Go 1.22
- **Framework**: Gin (HTTP router)
- **Database**: PostgreSQL via pgx/v5 (connection pooling)
- **Cache**: Redis via go-redis/v9
- **Migrations**: golang-migrate (sequential SQL files)
- **CORS**: gin-contrib/cors

## Infrastructure

- **Frontend hosting**: Cloudflare Pages
- **Backend hosting**: Fly.io (Docker)
- **Database**: Neon PostgreSQL (serverless)
- **Cache/Rate Limit**: Upstash Redis
- **Comments**: Giscus (GitHub Discussions)

## Commands

### Frontend

```bash
cd frontend
npm install          # Cài dependencies
npm run dev          # Dev server (localhost:4321)
npm run build        # Build static site
npm run preview      # Preview build
```

### Backend

```bash
cd backend
go mod tidy                    # Sync dependencies
make run                       # Chạy server (localhost:8080)
make build                     # Build binary
make test                      # Chạy tests
make migrate                   # Chạy migrations
make docker-build              # Build Docker image
docker compose up              # Khởi động PostgreSQL + Redis local
docker compose --profile full up  # Khởi động cả backend container
```
