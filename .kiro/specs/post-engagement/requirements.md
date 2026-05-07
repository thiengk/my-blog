# Requirements Document

## Introduction

Tính năng Post Engagement mở rộng hệ thống blog hiện tại bằng cách thêm khả năng tương tác cho bài viết: hệ thống bình luận nội bộ (thay thế Utterances/Giscus), đếm lượt like, comment, share, và xếp hạng bài viết theo mức độ tương tác để đề xuất (recommend) các bài viết phổ biến nhất ở đầu danh sách.

Hệ thống hiện tại đã có view counter với kiến trúc Redis batch + PostgreSQL. Tính năng mới sẽ mở rộng kiến trúc này để hỗ trợ thêm các loại tương tác khác.

## Glossary

- **Engagement_Service**: Service backend (Go) xử lý logic nghiệp vụ cho likes, comments, shares và tính toán engagement score
- **Comment_Service**: Service backend xử lý CRUD operations cho bình luận
- **Recommendation_Engine**: Module tính toán và xếp hạng bài viết dựa trên engagement score
- **Engagement_Score**: Điểm tương tác tổng hợp của một bài viết, được tính từ tổng likes, comments, và shares với trọng số cấu hình được
- **Post_Slug**: Định danh duy nhất của bài viết dưới dạng URL-friendly string
- **Engagement_API**: Các REST API endpoints phục vụ tính năng engagement
- **Comment_Widget**: Svelte component hiển thị và quản lý bình luận trên frontend
- **Engagement_Counter**: Svelte component hiển thị số lượng likes, comments, shares trên frontend
- **Redis_Cache**: Lớp cache Redis lưu trữ engagement counts và recommendation rankings
- **IP_Hash**: SHA-256 hash của địa chỉ IP người dùng, dùng để chống duplicate actions

## Requirements

### Requirement 1: Hệ thống bình luận nội bộ

**User Story:** As a reader, I want to leave comments on blog posts, so that I can share my thoughts and engage with the content.

#### Acceptance Criteria

1. WHEN a reader submits a comment with author name and content, THE Comment_Service SHALL create the comment and associate it with the specified Post_Slug
2. WHEN a reader views a blog post, THE Comment_Widget SHALL display all approved comments for that Post_Slug in chronological order (oldest first)
3. THE Comment_Service SHALL validate that comment content has a minimum length of 1 character and a maximum length of 5000 characters
4. THE Comment_Service SHALL validate that author name has a minimum length of 1 character and a maximum length of 100 characters
5. IF a comment submission contains empty author name or empty content, THEN THE Comment_Service SHALL return a validation error with a descriptive message
6. WHEN a comment is successfully created, THE Comment_Service SHALL increment the comment count for the associated Post_Slug
7. THE Comment_Service SHALL store each comment with author name, content, Post_Slug, creation timestamp, and IP_Hash

### Requirement 2: Hệ thống Like

**User Story:** As a reader, I want to like a blog post, so that I can express appreciation for the content.

#### Acceptance Criteria

1. WHEN a reader sends a like request for a Post_Slug, THE Engagement_Service SHALL record the like and increment the like count for that post
2. WHEN a reader has already liked a Post_Slug within 24 hours (based on IP_Hash), THE Engagement_Service SHALL reject the duplicate like and return a message indicating the post was already liked
3. WHEN a reader requests the like count for a Post_Slug, THE Engagement_Service SHALL return the current total like count
4. THE Engagement_Service SHALL use Redis_Cache to check for duplicate likes within the 24-hour window
5. THE Engagement_Service SHALL batch like counts in Redis and periodically flush to PostgreSQL (consistent with existing view count pattern)

### Requirement 3: Hệ thống Share tracking

**User Story:** As a blog owner, I want to track how many times posts are shared, so that I can understand content virality.

#### Acceptance Criteria

1. WHEN a reader clicks a share button for a Post_Slug, THE Engagement_Service SHALL record the share action and increment the share count for that post
2. WHEN a reader has already shared the same Post_Slug within 24 hours (based on IP_Hash), THE Engagement_Service SHALL reject the duplicate share
3. WHEN a reader requests the share count for a Post_Slug, THE Engagement_Service SHALL return the current total share count
4. THE Engagement_Service SHALL support tracking shares to specific platforms (facebook, twitter, linkedin, copy-link)
5. THE Engagement_Service SHALL batch share counts in Redis and periodically flush to PostgreSQL

### Requirement 4: Engagement Counter hiển thị

**User Story:** As a reader, I want to see the number of likes, comments, and shares on each post, so that I can gauge the popularity of the content.

#### Acceptance Criteria

1. WHEN a reader views a blog post, THE Engagement_Counter SHALL display the current like count, comment count, and share count for that Post_Slug
2. WHEN a reader views the blog post list, THE Engagement_Counter SHALL display engagement counts for each post in the list
3. THE Engagement_API SHALL provide a bulk endpoint that returns engagement counts (likes, comments, shares) for multiple Post_Slugs in a single request
4. THE Engagement_Counter SHALL update the displayed like count immediately after a successful like action without requiring a page reload
5. IF the Engagement_API is unavailable, THEN THE Engagement_Counter SHALL display counts as 0 and remain functional without blocking page rendering

### Requirement 5: Recommendation dựa trên Engagement Score

**User Story:** As a reader, I want to see the most engaging posts recommended at the top of the list, so that I can discover popular content easily.

#### Acceptance Criteria

1. THE Recommendation_Engine SHALL calculate Engagement_Score for each post using the formula: score = (likes × like_weight) + (comments × comment_weight) + (shares × share_weight)
2. THE Recommendation_Engine SHALL use configurable weights with default values: like_weight = 1, comment_weight = 2, share_weight = 3
3. WHEN a reader views the blog post list with recommendation mode enabled, THE Recommendation_Engine SHALL return posts sorted by Engagement_Score in descending order
4. THE Recommendation_Engine SHALL cache the ranked post list in Redis_Cache with a time-to-live of 5 minutes
5. WHEN the cached ranking expires, THE Recommendation_Engine SHALL recalculate rankings from the current engagement data in PostgreSQL
6. IF two posts have the same Engagement_Score, THEN THE Recommendation_Engine SHALL sort them by creation date (newest first) as a tiebreaker
7. THE Engagement_API SHALL provide an endpoint that returns the top N recommended posts (default N = 10, maximum N = 50)

### Requirement 6: API Rate Limiting và Anti-abuse

**User Story:** As a blog owner, I want to prevent abuse of the engagement system, so that engagement metrics remain authentic.

#### Acceptance Criteria

1. THE Engagement_API SHALL enforce rate limiting of 60 requests per minute per IP for like and share endpoints
2. THE Engagement_API SHALL enforce rate limiting of 30 requests per minute per IP for comment creation endpoint
3. THE Engagement_API SHALL use the existing Redis-based sliding window rate limiting mechanism
4. IF a client exceeds the rate limit, THEN THE Engagement_API SHALL return HTTP 429 status with a Retry-After header
5. THE Engagement_Service SHALL use IP_Hash (SHA-256) for duplicate detection to preserve user privacy (consistent with existing view count pattern)

### Requirement 7: Engagement Data Persistence

**User Story:** As a blog owner, I want engagement data to be reliably stored, so that metrics are not lost during system restarts or failures.

#### Acceptance Criteria

1. THE Engagement_Service SHALL store engagement counts (likes, comments, shares) in PostgreSQL as the source of truth
2. THE Engagement_Service SHALL use Redis_Cache as a write buffer with periodic batch flush to PostgreSQL (flush interval configurable, default 60 seconds)
3. IF Redis_Cache is unavailable, THEN THE Engagement_Service SHALL fall back to direct PostgreSQL writes for engagement recording
4. IF a PostgreSQL flush fails, THEN THE Engagement_Service SHALL retain the pending counts in Redis_Cache and retry on the next flush cycle
5. THE Comment_Service SHALL store comments directly in PostgreSQL (no Redis buffering for comment content)
6. THE Engagement_Service SHALL provide a bulk counts endpoint that returns likes, comments, and shares for up to 50 Post_Slugs in a single request
