# Kế hoạch Triển khai: Personal Blog

## Tổng quan

Triển khai hệ thống blog cá nhân theo kiến trúc JAMstack + API backend. Frontend sử dụng Astro + Svelte + TailwindCSS deploy trên Cloudflare Pages, backend Go (Gin) deploy trên Fly.io, với PostgreSQL (Neon) và Redis (Upstash) làm data layer. Các task được sắp xếp theo thứ tự: thiết lập project → content engine → frontend components → backend API → tích hợp → deployment.

## Tasks

- [ ] 1. Thiết lập cấu trúc project và cấu hình cơ bản
  - [x] 1.1 Khởi tạo Astro project với Svelte và TailwindCSS
    - Tạo Astro project mới với `create astro`
    - Cài đặt và cấu hình integration: `@astrojs/svelte`, `@astrojs/tailwind`
    - Cấu hình `astro.config.mjs` với output static, site URL
    - Thiết lập TailwindCSS config với dark mode class strategy
    - Tạo cấu trúc thư mục: `src/content/`, `src/components/`, `src/layouts/`, `src/pages/`
    - _Yêu cầu: 1.1, 4.4, 13.1_

  - [x] 1.2 Khởi tạo Go backend project
    - Tạo Go module với `go mod init`
    - Cài đặt dependencies: Gin, pgx (PostgreSQL driver), go-redis
    - Tạo cấu trúc thư mục: `cmd/server/`, `internal/handler/`, `internal/service/`, `internal/middleware/`, `internal/config/`
    - Tạo file `main.go` với basic Gin server setup
    - Tạo `internal/config/config.go` đọc environment variables (DB URL, Redis URL, port, rate limit configs)
    - _Yêu cầu: 13.2, 13.3, 13.6_

  - [x] 1.3 Thiết lập Docker và environment configuration
    - Tạo `Dockerfile` cho Go backend (multi-stage build)
    - Tạo `docker-compose.yml` cho local development (PostgreSQL + Redis)
    - Tạo `.env.example` với tất cả environment variables cần thiết
    - Tạo `Makefile` với các commands: build, run, test, migrate
    - _Yêu cầu: 13.2, 13.4, 13.6_

- [ ] 2. Content Engine - Xử lý nội dung Markdown/MDX
  - [x] 2.1 Cấu hình Astro Content Collections
    - Định nghĩa content collection schema trong `src/content/config.ts` với Zod validation
    - Schema bao gồm: title, description, date, updatedDate, category, tags, coverImage, draft, slug
    - Tạo sample blog posts trong `src/content/blog/` để test
    - _Yêu cầu: 1.5, 1.6, 2.5_

  - [x] 2.2 Implement Content Pipeline (parsing, TOC, reading time)
    - Cấu hình remark/rehype plugins cho Markdown processing
    - Implement plugin tạo TOC tự động từ headings (h2, h3, h4)
    - Implement utility tính reading time (200 từ/phút)
    - Cấu hình syntax highlighting với Shiki (hỗ trợ dual theme cho dark/light mode)
    - _Yêu cầu: 1.1, 1.2, 1.3, 1.4, 4.6_

  - [ ]* 2.3 Viết unit tests cho Content Pipeline
    - Test TOC generation từ các heading levels khác nhau
    - Test reading time calculation với các độ dài bài viết khác nhau
    - Test frontmatter validation (valid/invalid cases)
    - Test draft filtering logic
    - _Yêu cầu: 1.3, 1.4, 1.5, 1.6_

- [ ] 3. Frontend - Layout và Theme System
  - [x] 3.1 Tạo Base Layout và Navigation
    - Tạo `src/layouts/BaseLayout.astro` với HTML structure semantic (header, nav, main, footer)
    - Implement responsive navigation (mobile hamburger menu, desktop nav bar)
    - Thêm meta tags cơ bản (viewport, charset, description)
    - Thêm RSS feed link trong `<head>`
    - _Yêu cầu: 4.4, 8.3, 12.3_

  - [x] 3.2 Implement Theme Manager (Dark/Light Mode)
    - Tạo `src/components/ThemeToggle.svelte` component
    - Implement logic: detect system preference (`prefers-color-scheme`)
    - Implement toggle giữa dark/light mode không reload trang
    - Lưu preference vào localStorage, áp dụng lại khi revisit
    - Thêm inline script trong `<head>` để tránh flash of unstyled content (FOUC)
    - _Yêu cầu: 4.1, 4.2, 4.3_

  - [x] 3.3 Thiết lập TailwindCSS Design System
    - Cấu hình color palette cho cả light và dark mode
    - Đảm bảo contrast ratio tối thiểu 4.5:1 cho text (WCAG AA)
    - Tạo typography styles cho prose content (headings, paragraphs, lists, code)
    - Cấu hình responsive breakpoints: mobile (<768px), tablet (768-1024px), desktop (>1024px)
    - _Yêu cầu: 4.4, 4.5_

- [ ] 4. Frontend - Trang Blog và Phân loại
  - [x] 4.1 Tạo trang danh sách bài viết (Blog Index)
    - Tạo `src/pages/blog/index.astro` hiển thị danh sách bài viết (không draft)
    - Implement pagination với số bài mỗi trang có thể cấu hình
    - Hiển thị card cho mỗi bài: title, description, date, category, tags, cover image, reading time
    - Sắp xếp theo ngày mới nhất
    - _Yêu cầu: 1.6, 2.3, 2.4, 2.6_

  - [x] 4.2 Tạo trang chi tiết bài viết (Blog Post)
    - Tạo `src/pages/blog/[...slug].astro` dynamic route
    - Render Markdown content với full formatting
    - Hiển thị TOC sidebar (sticky trên desktop)
    - Hiển thị metadata: date, category, tags, reading time, view count placeholder
    - _Yêu cầu: 1.1, 1.2, 1.3, 1.4, 7.3_

  - [x] 4.3 Tạo trang Categories và Tags
    - Tạo `src/pages/categories/index.astro` liệt kê tất cả categories
    - Tạo `src/pages/categories/[category].astro` hiển thị bài viết theo category
    - Tạo `src/pages/tags/index.astro` liệt kê tất cả tags
    - Tạo `src/pages/tags/[tag].astro` hiển thị bài viết theo tag
    - Sắp xếp bài viết theo ngày mới nhất, có pagination
    - _Yêu cầu: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6_

- [ ] 5. Checkpoint - Kiểm tra frontend cơ bản
  - Ensure all tests pass, ask the user if questions arise.
  - Xác nhận: content rendering, navigation, theme toggle, responsive layout, categories/tags hoạt động đúng.

- [ ] 6. Frontend - Search và Interactive Components
  - [x] 6.1 Implement Client-side Search
    - Tạo search index JSON trong quá trình build (title, description, tags, content)
    - Tạo `src/components/Search.svelte` component sử dụng Fuse.js
    - Cấu hình Fuse.js với fuzzy search và prefix matching
    - Implement highlight từ khóa trong kết quả
    - Hiển thị "Không tìm thấy kết quả" kèm gợi ý khi không có match
    - Đảm bảo search response dưới 100ms trên client
    - _Yêu cầu: 3.1, 3.2, 3.3, 3.4, 3.5_

  - [x] 6.2 Implement Newsletter Form Component
    - Tạo `src/components/NewsletterForm.svelte`
    - Implement client-side email validation (RFC 5322 basic)
    - Hiển thị states: idle, loading, success, error
    - Hiển thị thông báo lỗi validation cụ thể
    - Đặt form trên trang chủ và cuối mỗi bài viết
    - _Yêu cầu: 6.1, 6.3_

  - [x] 6.3 Tích hợp Giscus Comment System
    - Tạo `src/components/Comments.svelte` wrapper cho Giscus widget
    - Cấu hình Giscus với GitHub repository và Discussions category
    - Implement theme switching cho Giscus (sync với blog theme)
    - Đặt component cuối mỗi bài viết
    - _Yêu cầu: 5.1, 5.2, 5.3, 5.4, 5.5_

  - [ ]* 6.4 Viết unit tests cho Search component
    - Test fuzzy search với các query khác nhau
    - Test highlight logic
    - Test empty results handling
    - _Yêu cầu: 3.2, 3.3, 3.4, 3.5_

- [ ] 7. Frontend - SEO, RSS, và Image Optimization
  - [x] 7.1 Implement SEO Components
    - Tạo `src/components/SEO.astro` component cho meta tags
    - Generate Open Graph tags (og:title, og:description, og:image, og:url, og:type)
    - Generate Twitter Card tags
    - Implement canonical URL cho mỗi trang
    - Tạo JSON-LD structured data (BlogPosting schema) cho bài viết
    - _Yêu cầu: 12.1, 12.4, 12.5_

  - [x] 7.2 Generate Sitemap và RSS Feed
    - Cấu hình `@astrojs/sitemap` integration cho sitemap.xml tự động
    - Implement RSS feed generation (RSS 2.0 valid XML)
    - RSS bao gồm 20 bài mới nhất với: title, description, link, pubDate, category
    - Hiển thị RSS icon trên giao diện
    - _Yêu cầu: 8.1, 8.2, 8.3, 12.2, 12.6_

  - [x] 7.3 Implement Image Optimization
    - Cấu hình Astro Image integration (`@astrojs/image` hoặc built-in `astro:assets`)
    - Tự động convert sang WebP/AVIF với multiple sizes (srcset)
    - Implement lazy loading cho images ngoài viewport
    - Generate LQIP (blur placeholder) cho mỗi image
    - Resize images lớn hơn 2000px width xuống max 2000px
    - _Yêu cầu: 9.1, 9.2, 9.5, 9.6_

  - [ ]* 7.4 Viết tests cho RSS feed generation
    - Validate RSS output là XML hợp lệ
    - Validate tuân thủ RSS 2.0 specification
    - Test số lượng bài viết trong feed (max 20)
    - _Yêu cầu: 8.1, 8.4_

- [ ] 8. Backend - Database Schema và Connection
  - [x] 8.1 Tạo Database Migration Files
    - Tạo migration tool setup (golang-migrate hoặc goose)
    - Tạo migration cho bảng `post_views` (slug, view_count, timestamps)
    - Tạo migration cho bảng `newsletter_subscribers` (email, status, verification_token, timestamps)
    - Tạo migration cho bảng `view_logs` (slug, ip_hash, viewed_at)
    - Tạo indexes theo design: slug, category, email, status, token, viewed_at
    - _Yêu cầu: 11.1_

  - [x] 8.2 Implement Database Connection Pool
    - Tạo `internal/database/postgres.go` với pgxpool connection
    - Cấu hình connection pooling (max connections từ env var)
    - Implement graceful shutdown cho DB connections
    - Tạo `internal/database/redis.go` với Upstash Redis client
    - _Yêu cầu: 11.4_

- [ ] 9. Backend - Rate Limiter Middleware
  - [x] 9.1 Implement Sliding Window Rate Limiter
    - Tạo `internal/middleware/ratelimit.go`
    - Implement sliding window algorithm sử dụng Redis sorted sets
    - Hỗ trợ cấu hình khác nhau cho từng endpoint group
    - Default: 100 req/min cho public endpoints, 10 req/min cho newsletter
    - Trả về HTTP 429 với header `Retry-After` khi vượt limit
    - Implement fail-open khi Redis không khả dụng (log warning)
    - _Yêu cầu: 10.1, 10.2, 10.3, 10.4, 10.5_

  - [ ]* 9.2 Viết unit tests cho Rate Limiter
    - Test sliding window counting chính xác
    - Test fail-open behavior khi Redis down
    - Test different rate limit configs cho different endpoints
    - Test HTTP 429 response format
    - _Yêu cầu: 10.1, 10.2, 10.3, 10.4, 10.5, 10.6_

- [ ] 10. Backend - View Count Service
  - [x] 10.1 Implement View Count Handler và Service
    - Tạo `internal/service/viewcount.go` với ViewCountService interface
    - Implement `RecordView`: check duplicate bằng Redis key `view:seen:{slug}:{ip_hash}` (TTL 24h)
    - Implement batch increment: tăng `view:batch:{slug}` trong Redis
    - Implement `GetCount`: đọc từ cache `view:count:{slug}`, fallback DB
    - Implement `FlushBatch`: cron job đẩy batch counts từ Redis vào PostgreSQL mỗi 5 phút
    - Tạo `internal/handler/viewcount.go` với endpoints: POST /api/views/:slug, GET /api/views/:slug, GET /api/views
    - _Yêu cầu: 7.1, 7.2, 7.3, 7.4, 7.5_

  - [ ]* 10.2 Viết unit tests cho View Count Service
    - Test duplicate detection (same IP within 24h)
    - Test batch flush logic
    - Test cache fallback khi DB down
    - Test bulk view count retrieval
    - _Yêu cầu: 7.1, 7.2, 7.4, 7.5_

- [ ] 11. Backend - Newsletter Service
  - [x] 11.1 Implement Newsletter Handler và Service
    - Tạo `internal/service/newsletter.go` với NewsletterService interface
    - Implement `Subscribe`: validate email (RFC 5322), check duplicate, save to DB, generate verification token
    - Implement `Unsubscribe`: update status to 'unsubscribed'
    - Implement `VerifyEmail`: verify token, update status to 'active'
    - Trả về thông báo phù hợp khi email đã tồn tại
    - Tạo `internal/handler/newsletter.go` với endpoints: POST /api/newsletter/subscribe, POST /api/newsletter/unsubscribe, GET /api/newsletter/verify/:token
    - _Yêu cầu: 6.2, 6.4, 6.5, 6.6_

  - [ ]* 11.2 Viết unit tests cho Newsletter Service
    - Test email validation (valid/invalid formats)
    - Test duplicate email handling
    - Test verification token flow
    - Test unsubscribe flow
    - _Yêu cầu: 6.2, 6.4, 6.5_

- [ ] 12. Backend - Health Check và Caching
  - [x] 12.1 Implement Health Check Endpoint
    - Tạo `internal/handler/health.go`
    - Check PostgreSQL connection status
    - Check Redis connection status
    - Trả về JSON với status của từng service
    - Endpoint: GET /api/health (không rate limit)
    - _Yêu cầu: 11.5_

  - [x] 12.2 Implement API Response Caching
    - Tạo `internal/middleware/cache.go` middleware
    - Cache GET responses trong Redis với TTL cấu hình được
    - Implement cache invalidation khi data thay đổi (write-through)
    - Cache keys: `cache:views:{slug}`, `cache:views:bulk:{hash}`
    - _Yêu cầu: 11.2, 11.3, 11.6_

- [ ] 13. Checkpoint - Kiểm tra backend hoàn chỉnh
  - Ensure all tests pass, ask the user if questions arise.
  - Xác nhận: API endpoints hoạt động, rate limiting, view count batching, newsletter flow, health check, caching.

- [ ] 14. Tích hợp Frontend-Backend
  - [x] 14.1 Kết nối View Count từ Frontend
    - Tạo `src/lib/api.ts` utility cho API calls
    - Implement gọi POST /api/views/:slug khi user truy cập bài viết
    - Implement hiển thị view count trên trang chi tiết và danh sách bài viết
    - Fetch bulk view counts cho trang danh sách
    - Handle graceful degradation khi API không khả dụng
    - _Yêu cầu: 7.1, 7.3_

  - [x] 14.2 Kết nối Newsletter Form với Backend
    - Wire NewsletterForm.svelte gọi POST /api/newsletter/subscribe
    - Handle response states: success, duplicate email, validation error
    - Implement loading state và error display
    - _Yêu cầu: 6.1, 6.2, 6.3, 6.4_

  - [ ]* 14.3 Viết integration tests cho Frontend-Backend flow
    - Test view count increment và display
    - Test newsletter subscribe flow end-to-end
    - Test error handling khi API down
    - _Yêu cầu: 7.1, 7.3, 6.2_

- [ ] 15. Performance và CDN Configuration
  - [x] 15.1 Cấu hình Cloudflare Pages và Cache Headers
    - Tạo `_headers` file cho Cloudflare Pages
    - Set immutable cache cho hashed assets (JS, CSS, images)
    - Set stale-while-revalidate cho HTML pages
    - Cấu hình Cloudflare Pages build settings
    - _Yêu cầu: 9.3, 9.4, 13.1_

  - [x] 15.2 Cấu hình Fly.io Deployment
    - Tạo `fly.toml` configuration file
    - Cấu hình health check endpoint cho Fly.io
    - Cấu hình auto-scaling rules
    - Setup environment variables trên Fly.io
    - _Yêu cầu: 13.2, 13.3, 13.5, 13.7_

- [ ] 16. Checkpoint cuối - Kiểm tra toàn bộ hệ thống
  - Ensure all tests pass, ask the user if questions arise.
  - Xác nhận: frontend build thành công, backend chạy ổn định, tích hợp hoạt động, deployment configs đúng, SEO/RSS/sitemap generate đúng.

## Ghi chú

- Các task đánh dấu `*` là optional và có thể bỏ qua để ra MVP nhanh hơn
- Mỗi task tham chiếu đến yêu cầu cụ thể để đảm bảo traceability
- Checkpoints đảm bảo kiểm tra incremental sau mỗi giai đoạn lớn
- Frontend và backend có thể phát triển song song sau khi thiết lập xong cấu trúc ban đầu
- Tất cả cấu hình nhạy cảm sử dụng environment variables, không hardcode
