# Personal Blog — Tổng quan dự án

## 1. Giới thiệu

Hệ thống blog cá nhân hiện đại, được xây dựng theo kiến trúc **JAMstack + API backend**. Blog phục vụ mục đích chia sẻ kinh nghiệm làm việc, review khóa học công nghệ, và câu chuyện cuộc sống với phong cách storytelling.

Hệ thống được thiết kế để:
- Tải nhanh (static-first, CDN-delivered)
- Scale được khi lượng truy cập tăng (stateless API, horizontal scaling)
- Chi phí thấp ban đầu (sử dụng free tier của các dịch vụ cloud)
- Dễ bảo trì và mở rộng

## 2. Kiến trúc hệ thống

```
┌─────────────────────────────────────────────────────────────┐
│                        Người dùng                            │
└─────────────────────────┬───────────────────────────────────┘
                          │
              ┌───────────▼───────────┐
              │   Cloudflare CDN      │
              │   (Cache + Security)  │
              └───────┬───────┬───────┘
                      │       │
         ┌────────────▼──┐  ┌─▼────────────────┐
         │   Frontend    │  │   Backend API     │
         │   (Static)    │  │   (Go / Gin)      │
         │               │  │                   │
         │ Astro 5       │  │ Fly.io            │
         │ Svelte 5      │  │ Auto-scale 1-3    │
         │ TailwindCSS   │  │ instances         │
         │               │  │                   │
         │ Cloudflare    │  └───────┬───────────┘
         │ Pages         │          │
         └───────────────┘    ┌─────▼─────┐
                              │           │
                    ┌─────────▼──┐  ┌─────▼──────┐
                    │ PostgreSQL │  │   Redis    │
                    │ (Neon)     │  │ (Upstash)  │
                    │            │  │            │
                    │ Data store │  │ Cache +    │
                    │            │  │ Rate limit │
                    └────────────┘  └────────────┘
```

## 3. Tech Stack

| Layer | Công nghệ | Lý do chọn |
|-------|-----------|------------|
| Frontend Framework | Astro 5 + Svelte 5 | Static generation tốt, interactive islands nhẹ |
| Styling | TailwindCSS | Utility-first, tree-shaking, bundle nhỏ |
| Backend | Go (Gin) | Performance cao, binary nhỏ, deploy dễ |
| Database | PostgreSQL (Neon) | Reliable, free tier tốt, serverless |
| Cache | Redis (Upstash) | REST-based, free tier, rate limiting |
| Search | Fuse.js (client-side) | Không cần server, phù hợp blog cá nhân |
| Comments | Internal (PostgreSQL) | Không phụ thuộc bên thứ 3, full control |
| Frontend Hosting | Cloudflare Pages | Free, global CDN, auto-build từ Git |
| Backend Hosting | Fly.io | Free tier, container support, auto-scale |

## 4. Tính năng chính

### 4.1 Quản lý nội dung
- Viết bài bằng Markdown/MDX
- Syntax highlighting (20+ ngôn ngữ, dual theme light/dark)
- Tự động tạo mục lục (TOC) từ headings
- Tính thời gian đọc ước tính
- Hỗ trợ draft (bài nháp không hiển thị công khai)

### 4.2 Phân loại & Tìm kiếm
- Categories (1 category/bài) và Tags (nhiều tags/bài)
- Pagination có thể cấu hình
- Tìm kiếm fuzzy client-side (< 100ms, hỗ trợ sai chính tả)

### 4.3 Giao diện
- Dark/Light mode (detect system preference, lưu localStorage)
- Responsive: mobile, tablet, desktop
- WCAG AA compliant (contrast ratio ≥ 4.5:1)

### 4.4 Tương tác
- Đếm lượt xem bài viết (chống duplicate 24h)
- Like bài viết (dedup 24h theo IP, localStorage persistence)
- Share tracking (Facebook, Twitter, LinkedIn, Copy Link) với platform analytics
- Bình luận nội bộ (không phụ thuộc GitHub/Giscus)
- Engagement counters (likes, comments, shares) hiển thị trên post detail và blog list
- Recommendation engine (bài viết nổi bật theo engagement score)

### 4.5 SEO & Performance
- Open Graph + Twitter Card meta tags
- JSON-LD structured data (BlogPosting schema)
- Sitemap.xml tự động
- RSS 2.0 feed (20 bài mới nhất)
- Image optimization (WebP, responsive srcset, lazy loading)
- Lighthouse Performance ≥ 90/100

### 4.6 Bảo mật & Ổn định
- Rate limiting (sliding window, Redis sorted sets)
- Fail-open khi Redis down
- Health check endpoint
- API response caching
- CORS configuration
- Security headers (X-Frame-Options, CSP, etc.)

## 5. API Endpoints

| Method | Path | Mô tả | Rate Limit |
|--------|------|--------|------------|
| GET | `/api/health` | Health check | Không giới hạn |
| POST | `/api/views/:slug` | Ghi nhận lượt xem | 100 req/min |
| GET | `/api/views/:slug` | Lấy view count | 100 req/min |
| GET | `/api/views?slugs=a,b,c` | Bulk view counts | 100 req/min |
| POST | `/api/engagement/like/:slug` | Ghi nhận like | 60 req/min |
| POST | `/api/engagement/share/:slug` | Ghi nhận share | 60 req/min |
| GET | `/api/engagement/:slug` | Lấy engagement counts | Không giới hạn |
| GET | `/api/engagement?slugs=a,b,c` | Bulk engagement counts | Không giới hạn |
| POST | `/api/comments/:slug` | Tạo comment mới | 30 req/min |
| GET | `/api/comments/:slug` | Lấy danh sách comments | Không giới hạn |
| GET | `/api/recommendations?limit=10` | Bài viết đề xuất | Không giới hạn |

**Lưu ý:** Newsletter endpoints đã bị vô hiệu hóa để đơn giản hóa project.

## 6. Database Schema

### post_views
| Column | Type | Mô tả |
|--------|------|-------|
| id | BIGSERIAL | Primary key |
| slug | VARCHAR(255) | Post slug (UNIQUE) |
| view_count | BIGINT | Tổng lượt xem |
| created_at | TIMESTAMPTZ | Thời gian tạo |
| updated_at | TIMESTAMPTZ | Cập nhật cuối |

### view_logs
| Column | Type | Mô tả |
|--------|------|-------|
| id | BIGSERIAL | Primary key |
| slug | VARCHAR(255) | Post slug |
| ip_hash | VARCHAR(64) | SHA-256 hash IP |
| viewed_at | TIMESTAMPTZ | Thời gian xem |

### post_engagement
| Column | Type | Mô tả |
|--------|------|-------|
| id | BIGSERIAL | Primary key |
| slug | VARCHAR(255) | Post slug (UNIQUE) |
| like_count | BIGINT | Tổng lượt like |
| comment_count | BIGINT | Tổng comments |
| share_count | BIGINT | Tổng lượt share |
| created_at | TIMESTAMPTZ | Thời gian tạo |
| updated_at | TIMESTAMPTZ | Cập nhật cuối |

### comments
| Column | Type | Mô tả |
|--------|------|-------|
| id | BIGSERIAL | Primary key |
| slug | VARCHAR(255) | Post slug |
| author_name | VARCHAR(100) | Tên người bình luận |
| content | TEXT | Nội dung bình luận |
| ip_hash | VARCHAR(64) | SHA-256 hash IP |
| created_at | TIMESTAMPTZ | Thời gian tạo |

### share_logs
| Column | Type | Mô tả |
|--------|------|-------|
| id | BIGSERIAL | Primary key |
| slug | VARCHAR(255) | Post slug |
| platform | VARCHAR(20) | Platform (facebook, twitter, linkedin, copy-link) |
| ip_hash | VARCHAR(64) | SHA-256 hash IP |
| shared_at | TIMESTAMPTZ | Thời gian share |

**Lưu ý:** Bảng `newsletter_subscribers` vẫn tồn tại trong database nhưng không được sử dụng.

## 7. Chiến lược Scaling

| Giai đoạn | Traffic | Cấu hình |
|-----------|---------|----------|
| Phase 1 | 0 - 10K views/tháng | 1 Fly.io instance, Neon free, Upstash free |
| Phase 2 | 10K - 100K views/tháng | 2-3 Fly.io instances, Neon Pro, Upstash pay-as-you-go |
| Phase 3 | 100K+ views/tháng | 5+ Fly.io instances, Neon read replicas, Upstash Pro |

## 8. Chi phí ước tính

### Phase 1 (Free tier)
| Dịch vụ | Chi phí |
|---------|---------|
| Cloudflare Pages | $0 |
| Fly.io (1 shared-cpu) | $0 |
| Neon PostgreSQL (0.5GB) | $0 |
| Upstash Redis (10K cmd/ngày) | $0 |
| **Tổng** | **$0/tháng** |

### Phase 2 (Scale up)
| Dịch vụ | Chi phí ước tính |
|---------|-----------------|
| Cloudflare Pages | $0 |
| Fly.io (2-3 instances) | ~$5-15/tháng |
| Neon Pro | ~$19/tháng |
| Upstash Pay-as-you-go | ~$5-10/tháng |
| **Tổng** | **~$30-45/tháng** |

## 9. Timeline phát triển

| Phase | Nội dung | Trạng thái |
|-------|----------|-----------|
| Setup | Project structure, Docker, configs | ✅ Hoàn thành |
| Content Engine | Markdown parsing, TOC, reading time | ✅ Hoàn thành |
| Frontend UI | Layout, theme, pages, search | ✅ Hoàn thành |
| Backend API | View count, newsletter, rate limit, caching | ✅ Hoàn thành |
| Integration | Frontend ↔ Backend connection | ✅ Hoàn thành |
| Deployment | Cloudflare Pages + Fly.io configs | ✅ Hoàn thành |
| Post Engagement | Like, share, comments, recommendations | ✅ Hoàn thành |
| **Go Live** | Cài đặt, test, deploy production | ⏳ Chờ cấp quyền |

## 10. Liên hệ

- **Repository**: [link repo]
- **Hướng dẫn cài đặt**: Xem file `docs/SETUP.md`
- **Spec chi tiết**: Xem thư mục `.kiro/specs/personal-blog/`
