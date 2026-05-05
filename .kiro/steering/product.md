# Product Overview

Blog cá nhân chia sẻ kinh nghiệm làm việc, review khóa học công nghệ, và câu chuyện cuộc sống. Lấy cảm hứng từ phong cách storytelling, kiến trúc scale-ready.

## Tính năng chính

- **Content Engine**: Viết bài bằng Markdown/MDX, tự động render HTML với syntax highlighting, TOC, reading time
- **Phân loại**: Categories (1 per post) và Tags (nhiều per post), pagination
- **Tìm kiếm**: Client-side search với Fuse.js (fuzzy, prefix matching, < 100ms)
- **Dark/Light Mode**: Detect system preference, toggle không reload, lưu localStorage
- **Bình luận**: Giscus (GitHub Discussions), sync theme
- **Newsletter**: Đăng ký email, verification token, unsubscribe
- **View Count**: Đếm lượt xem, batch update qua Redis, chống duplicate 24h
- **SEO**: Open Graph, Twitter Card, JSON-LD, canonical URL, sitemap, RSS feed
- **Image Optimization**: WebP, responsive srcset, lazy loading, LQIP blur placeholder
- **Rate Limiting**: Sliding window (Redis sorted sets), fail-open, per-endpoint config
- **Caching**: Redis cache cho API responses, write-through invalidation

## Kiến trúc

```
CDN (Cloudflare Pages) → Static Frontend (Astro + Svelte + TailwindCSS)
                       → Go API (stateless, horizontal scaling)
                       → PostgreSQL (Neon) + Redis (Upstash)
```

## Deployment

- Frontend: Cloudflare Pages (static, auto-build từ Git)
- Backend: Fly.io (Docker container, auto-scale 1-3 instances)
- Database: Neon PostgreSQL (serverless)
- Cache: Upstash Redis (REST-based)
