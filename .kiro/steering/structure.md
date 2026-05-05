# Project Structure

```
.kiro/
  steering/                    # AI steering rules
  specs/personal-blog/         # Spec documents (requirements, design, tasks)

frontend/                      # Astro + Svelte + TailwindCSS (static site)
  public/
    _headers                   # Cloudflare Pages cache headers
    _redirects                 # Cloudflare Pages redirects
    favicon.svg
  src/
    components/
      BlogCard.astro           # Blog post preview card
      BlogPostSchema.astro     # JSON-LD structured data
      Comments.svelte          # Giscus comment wrapper
      Footer.astro             # Site footer
      Navigation.svelte        # Responsive nav (mobile hamburger + desktop)
      NewsletterForm.svelte    # Email subscription form
      OptimizedImage.astro     # Image optimization wrapper
      Pagination.astro         # Page navigation
      Search.svelte            # Client-side search (Fuse.js)
      SEO.astro                # Meta tags (OG, Twitter Card, canonical)
      TableOfContents.astro    # Sticky TOC sidebar
      ThemeToggle.svelte       # Dark/Light mode toggle
      ViewCounter.svelte       # View count display + record
    content/
      config.ts                # Content collection schema (Zod)
      blog/                    # Markdown blog posts
    layouts/
      BaseLayout.astro         # Main layout (header, nav, main, footer)
    lib/
      api.ts                   # Backend API utility (view count, newsletter)
      image-utils.ts           # Image optimization utilities
      reading-time.ts          # Reading time calculation
      toc.ts                   # TOC generation from headings
    pages/
      index.astro              # Homepage
      blog/index.astro         # Blog listing (page 1)
      blog/[...page].astro     # Blog listing (paginated)
      blog/[...slug].astro     # Blog post detail
      categories/index.astro   # All categories
      categories/[category]/[...page].astro
      tags/index.astro         # All tags
      tags/[tag]/[...page].astro
      rss.xml.ts               # RSS feed endpoint
      search.json.ts           # Search index endpoint
    styles/
      global.css               # TailwindCSS + design system
  astro.config.mjs
  tailwind.config.mjs
  svelte.config.js
  tsconfig.json
  package.json

backend/                       # Go API server (Gin)
  cmd/server/
    main.go                    # Entry point, server setup, graceful shutdown
  internal/
    config/
      config.go                # Environment variable configuration
    database/
      postgres.go              # PostgreSQL connection pool (pgxpool)
      redis.go                 # Redis client (Upstash compatible)
    handler/
      health.go                # GET /api/health
      health_test.go
      newsletter.go            # POST /subscribe, /unsubscribe, GET /verify/:token
      viewcount.go             # POST/GET /api/views/:slug, GET /api/views
    middleware/
      cache.go                 # API response caching (Redis)
      cache_test.go
      ratelimit.go             # Sliding window rate limiter (Redis sorted sets)
    service/
      newsletter.go            # Newsletter business logic
      viewcount.go             # View count with batching
  migrations/
    000001_create_post_views.up.sql
    000001_create_post_views.down.sql
    000002_create_newsletter_subscribers.up.sql
    000002_create_newsletter_subscribers.down.sql
    000003_create_view_logs.up.sql
    000003_create_view_logs.down.sql
  Dockerfile                   # Multi-stage build (golang → alpine)
  fly.toml                     # Fly.io deployment config
  Makefile                     # Build, test, migrate commands
  go.mod
  go.sum

docker-compose.yml             # PostgreSQL 16 + Redis 7 (local dev)
.env.example                   # Environment variables template
.gitignore
```
