# Implementation Plan: Post Engagement

## Overview

Triển khai hệ thống tương tác bài viết bao gồm: bình luận nội bộ, like, share tracking, engagement counters, và recommendation engine. Backend sử dụng Go + Gin, frontend sử dụng Svelte 5 + Astro. Kiến trúc theo pattern hiện có: Redis batch + flush + dedup, PostgreSQL là source of truth.

## Tasks

- [x] 1. Tạo database migrations và data models
  - [x] 1.1 Tạo migration cho bảng post_engagement
    - Tạo file `backend/migrations/000004_create_post_engagement.up.sql` với schema bảng `post_engagement` (id, slug, like_count, comment_count, share_count, created_at, updated_at)
    - Tạo file `backend/migrations/000004_create_post_engagement.down.sql`
    - Tạo indexes: unique trên slug, computed index cho engagement score
    - _Requirements: 7.1_

  - [x] 1.2 Tạo migration cho bảng comments
    - Tạo file `backend/migrations/000005_create_comments.up.sql` với schema bảng `comments` (id, slug, author_name, content, ip_hash, created_at)
    - Tạo file `backend/migrations/000005_create_comments.down.sql`
    - Tạo indexes: idx_comments_slug, idx_comments_slug_created
    - _Requirements: 1.7, 7.5_

  - [x] 1.3 Tạo migration cho bảng share_logs
    - Tạo file `backend/migrations/000006_create_share_logs.up.sql` với schema bảng `share_logs` (id, slug, platform, ip_hash, shared_at)
    - Tạo file `backend/migrations/000006_create_share_logs.down.sql`
    - Tạo index: idx_share_logs_slug_platform
    - _Requirements: 3.4_

- [x] 2. Implement EngagementService (likes & shares)
  - [x] 2.1 Tạo EngagementService interface và struct
    - Tạo file `backend/internal/service/engagement.go`
    - Định nghĩa interface `EngagementService` với các methods: RecordLike, RecordShare, GetCounts, GetBulkCounts, FlushBatch
    - Định nghĩa struct `EngagementCounts` (likes, comments, shares)
    - Implement constructor `NewEngagementService(db *pgxpool.Pool, redis *redis.Client)`
    - Implement helper `hashIP` (tái sử dụng pattern từ viewcount.go)
    - _Requirements: 2.1, 3.1, 7.1_

  - [x] 2.2 Implement RecordLike và RecordShare
    - Implement `RecordLike`: check dedup key `like:seen:{slug}:{ip_hash}` (TTL 24h), nếu chưa tồn tại thì SET dedup key và INCR `like:batch:{slug}`
    - Implement `RecordShare`: check dedup key `share:seen:{slug}:{ip_hash}` (TTL 24h), nếu chưa tồn tại thì SET dedup key, INCR `share:batch:{slug}`, và INSERT vào share_logs với platform info
    - Xử lý fallback khi Redis unavailable: ghi trực tiếp vào PostgreSQL
    - _Requirements: 2.1, 2.2, 2.4, 2.5, 3.1, 3.2, 3.4, 3.5, 6.5, 7.3_

  - [x] 2.3 Implement GetCounts và GetBulkCounts
    - Implement `GetCounts`: check Redis cache `engagement:count:{slug}` trước, fallback query PostgreSQL, cache kết quả với TTL 5 phút, cộng thêm pending batch counts
    - Implement `GetBulkCounts`: lặp qua danh sách slugs, gọi GetCounts cho từng slug, giới hạn tối đa 50 slugs
    - _Requirements: 2.3, 3.3, 4.1, 4.3, 7.6_

  - [x] 2.4 Implement FlushBatch
    - SCAN Redis cho `like:batch:*` và `share:batch:*`
    - GETDEL từng batch key
    - UPSERT vào bảng post_engagement trong PostgreSQL
    - Nếu DB write fail, put count back vào Redis (retry on next cycle)
    - _Requirements: 2.5, 3.5, 7.2, 7.4_

  - [x]* 2.5 Viết unit tests cho EngagementService
    - Test RecordLike: trường hợp like mới, duplicate like trong 24h
    - Test RecordShare: trường hợp share mới, duplicate share, platform tracking
    - Test GetCounts: cache hit, cache miss, pending batch counts
    - Test FlushBatch: flush thành công, DB failure retry
    - _Requirements: 2.1, 2.2, 3.1, 3.2, 7.2, 7.4_

- [x] 3. Implement CommentService
  - [x] 3.1 Tạo CommentService interface và struct
    - Tạo file `backend/internal/service/comment.go`
    - Định nghĩa interface `CommentService` với methods: CreateComment, GetComments, GetCommentCount
    - Định nghĩa structs: `Comment`, `CreateCommentInput`
    - Implement constructor `NewCommentService(db *pgxpool.Pool, redis *redis.Client)`
    - _Requirements: 1.1, 1.7_

  - [x] 3.2 Implement CreateComment
    - Validate author_name (1-100 ký tự) và content (1-5000 ký tự)
    - Trả về validation error nếu input không hợp lệ
    - Hash IP bằng SHA-256
    - INSERT vào bảng comments trong PostgreSQL (direct write, không qua Redis buffer)
    - INCR comment count trong Redis cache `comment:count:{slug}`
    - Cập nhật comment_count trong bảng post_engagement
    - _Requirements: 1.1, 1.3, 1.4, 1.5, 1.6, 1.7, 6.5, 7.5_

  - [x] 3.3 Implement GetComments và GetCommentCount
    - `GetComments`: query bảng comments WHERE slug = ?, ORDER BY created_at ASC
    - `GetCommentCount`: check Redis cache `comment:count:{slug}` trước, fallback COUNT(*) từ PostgreSQL
    - _Requirements: 1.2, 4.1_

  - [x]* 3.4 Viết unit tests cho CommentService
    - Test CreateComment: tạo thành công, validation errors (empty author, empty content, quá dài)
    - Test GetComments: trả về đúng thứ tự chronological
    - Test GetCommentCount: cache hit, cache miss
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5_

- [x] 4. Implement RecommendationService
  - [x] 4.1 Tạo RecommendationService interface và struct
    - Tạo file `backend/internal/service/recommendation.go`
    - Định nghĩa interface `RecommendationService` với methods: GetTopPosts, RecalculateRankings
    - Định nghĩa structs: `RankedPost`, `RecommendationConfig`
    - Implement constructor với configurable weights (default: like=1, comment=2, share=3)
    - _Requirements: 5.1, 5.2_

  - [x] 4.2 Implement GetTopPosts và RecalculateRankings
    - `GetTopPosts`: check Redis sorted set `recommendations:top` trước (TTL 5 phút), nếu cache miss thì gọi RecalculateRankings
    - `RecalculateRankings`: query PostgreSQL tính engagement_score = (like_count × like_weight) + (comment_count × comment_weight) + (share_count × share_weight), sort DESC, tiebreaker by created_at DESC
    - Lưu kết quả vào Redis sorted set với TTL 5 phút
    - Giới hạn maximum 50 results
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 5.6, 5.7_

  - [x]* 4.3 Viết unit tests cho RecommendationService
    - Test GetTopPosts: cache hit, cache miss + recalculate
    - Test scoring formula với các weights khác nhau
    - Test tiebreaker khi cùng score
    - Test limit parameter (default 10, max 50)
    - _Requirements: 5.1, 5.2, 5.6, 5.7_

- [x] 5. Checkpoint - Đảm bảo tất cả service tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 6. Implement API Handlers
  - [x] 6.1 Tạo EngagementHandler
    - Tạo file `backend/internal/handler/engagement.go`
    - Implement `NewEngagementHandler(svc EngagementService)`
    - Implement `RegisterRoutes` đăng ký: POST `/api/engagement/like/:slug`, POST `/api/engagement/share/:slug`, GET `/api/engagement/:slug`, GET `/api/engagement`
    - POST like: extract IP, gọi service.RecordLike, trả về `{counted: bool}`
    - POST share: extract IP + platform từ request body, gọi service.RecordShare
    - GET single: gọi service.GetCounts, trả về engagement counts
    - GET bulk: parse query param `slugs`, validate max 50, gọi service.GetBulkCounts
    - _Requirements: 2.1, 2.3, 3.1, 3.3, 4.1, 4.3, 7.6_

  - [x] 6.2 Tạo CommentHandler
    - Tạo file `backend/internal/handler/comment.go`
    - Implement `NewCommentHandler(svc CommentService)`
    - Implement `RegisterRoutes` đăng ký: POST `/api/comments/:slug`, GET `/api/comments/:slug`
    - POST: parse JSON body (author_name, content), extract IP, gọi service.CreateComment, trả về 201 Created
    - GET: gọi service.GetComments, trả về danh sách comments
    - _Requirements: 1.1, 1.2_

  - [x] 6.3 Tạo RecommendationHandler
    - Tạo file `backend/internal/handler/recommendation.go`
    - Implement `NewRecommendationHandler(svc RecommendationService)`
    - Implement `RegisterRoutes` đăng ký: GET `/api/recommendations`
    - Parse query param `limit` (default 10, max 50), gọi service.GetTopPosts
    - _Requirements: 5.3, 5.7_

  - [x]* 6.4 Viết unit tests cho handlers
    - Test EngagementHandler: like success, like duplicate, share success, get counts, bulk counts
    - Test CommentHandler: create comment success, validation error, get comments
    - Test RecommendationHandler: get recommendations với limit
    - _Requirements: 1.1, 2.1, 3.1, 5.7_

- [x] 7. Cấu hình Rate Limiting và wiring trong main.go
  - [x] 7.1 Thêm rate limit configs cho engagement endpoints
    - Thêm rate limit config `engagement` (60 req/min) cho like và share endpoints
    - Thêm rate limit config `comments` (30 req/min) cho comment creation endpoint
    - Sử dụng existing `RateLimiter.Middleware()` pattern
    - _Requirements: 6.1, 6.2, 6.3, 6.4_

  - [x] 7.2 Wire services và handlers trong main.go
    - Khởi tạo EngagementService, CommentService, RecommendationService
    - Khởi tạo EngagementHandler, CommentHandler, RecommendationHandler
    - Đăng ký routes với rate limiting middleware
    - Thiết lập background goroutine cho FlushBatch (interval 60s) tương tự view count flush
    - _Requirements: 6.1, 6.2, 6.3, 7.2_

- [x] 8. Checkpoint - Đảm bảo backend hoạt động đúng
  - Ensure all tests pass, ask the user if questions arise.

- [x] 9. Implement Frontend Components
  - [x] 9.1 Tạo LikeButton.svelte component
    - Tạo file `frontend/src/components/LikeButton.svelte`
    - Hiển thị like count hiện tại
    - Xử lý click: gửi POST request tới `/api/engagement/like/:slug`
    - Optimistic UI update khi like thành công
    - Lưu trạng thái đã like vào localStorage, disable button nếu đã like
    - Xử lý error states
    - _Requirements: 2.1, 4.4_

  - [x] 9.2 Tạo ShareButtons.svelte component
    - Tạo file `frontend/src/components/ShareButtons.svelte`
    - Hiển thị share count và các platform buttons (Facebook, Twitter, LinkedIn, Copy Link)
    - Khi click: mở share dialog của platform tương ứng và gửi POST request tới `/api/engagement/share/:slug`
    - Copy Link: copy URL vào clipboard
    - _Requirements: 3.1, 3.4_

  - [x] 9.3 Tạo EngagementCounter.svelte component
    - Tạo file `frontend/src/components/EngagementCounter.svelte`
    - Hiển thị likes, comments, shares counts
    - Sử dụng cho cả post detail view và post list view
    - Graceful degradation: hiển thị 0 khi API unavailable
    - _Requirements: 4.1, 4.2, 4.5_

  - [x] 9.4 Tạo CommentSection.svelte component (thay thế Comments.svelte)
    - Tạo file `frontend/src/components/CommentSection.svelte`
    - Form nhập comment: fields author name và content
    - Client-side validation: author (1-100 chars), content (1-5000 chars)
    - Hiển thị danh sách comments theo thứ tự chronological
    - Loading states và error handling
    - Sau khi submit thành công, thêm comment mới vào danh sách không cần reload
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5_

  - [x] 9.5 Tạo RecommendedPosts.svelte component
    - Tạo file `frontend/src/components/RecommendedPosts.svelte`
    - Fetch GET `/api/recommendations?limit=10`
    - Hiển thị danh sách bài viết được đề xuất với engagement score
    - Sử dụng trên homepage hoặc sidebar
    - _Requirements: 5.3, 5.7_

- [x] 10. Tích hợp Frontend Components vào Astro pages
  - [x] 10.1 Tích hợp engagement components vào blog post layout
    - Thêm LikeButton, ShareButtons, EngagementCounter vào blog post detail page
    - Thay thế Comments.svelte bằng CommentSection.svelte
    - Truyền slug prop cho các components
    - _Requirements: 1.1, 1.2, 2.1, 3.1, 4.1_

  - [x] 10.2 Tích hợp EngagementCounter vào blog list và RecommendedPosts
    - Thêm EngagementCounter vào BlogCard.astro (sử dụng bulk API)
    - Thêm RecommendedPosts component vào homepage hoặc sidebar
    - _Requirements: 4.2, 5.3_

- [x] 11. Final checkpoint - Đảm bảo toàn bộ hệ thống hoạt động
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- Tasks đánh dấu `*` là optional và có thể bỏ qua để ra MVP nhanh hơn
- Mỗi task tham chiếu đến requirements cụ thể để đảm bảo traceability
- Checkpoints đảm bảo validation từng giai đoạn
- Backend Go sử dụng pattern tương tự viewcount.go đã có sẵn (Redis batch + flush + dedup)
- Frontend Svelte 5 components tích hợp vào Astro static site
- Rate limiting tái sử dụng cơ chế sliding window Redis đã có

## Task Dependency Graph

```json
{
  "waves": [
    { "id": 0, "tasks": ["1.1", "1.2", "1.3"] },
    { "id": 1, "tasks": ["2.1", "3.1", "4.1"] },
    { "id": 2, "tasks": ["2.2", "2.3", "3.2", "3.3", "4.2"] },
    { "id": 3, "tasks": ["2.4", "2.5", "3.4", "4.3"] },
    { "id": 4, "tasks": ["6.1", "6.2", "6.3"] },
    { "id": 5, "tasks": ["6.4", "7.1"] },
    { "id": 6, "tasks": ["7.2"] },
    { "id": 7, "tasks": ["9.1", "9.2", "9.3", "9.4", "9.5"] },
    { "id": 8, "tasks": ["10.1", "10.2"] }
  ]
}
```
