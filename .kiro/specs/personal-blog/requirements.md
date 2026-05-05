# Tài liệu Yêu cầu - Personal Blog

## Giới thiệu

Hệ thống blog cá nhân cho phép chia sẻ kinh nghiệm làm việc, đánh giá/review khóa học công nghệ, và câu chuyện cuộc sống. Blog được xây dựng theo phong cách storytelling, lấy cảm hứng từ "toidicodedao", với kiến trúc scale-ready và sử dụng các công nghệ mới (Astro, Svelte, Go, TailwindCSS) để vừa học vừa làm.

**Kiến trúc tổng quan:**
```
CDN (Cloudflare) → Static Frontend (Astro + Svelte + TailwindCSS)
                 → Go API (stateless, horizontal scaling)
                 → PostgreSQL (Neon) + Redis cache (Upstash)
```

## Thuật ngữ (Glossary)

- **Blog_System**: Toàn bộ hệ thống blog bao gồm frontend, backend API, và database
- **Frontend**: Ứng dụng Astro + Svelte + TailwindCSS, được deploy trên Cloudflare Pages
- **API_Server**: Backend Go (Gin/Echo) xử lý các request động (view count, newsletter, etc.)
- **Content_Engine**: Module xử lý nội dung Markdown/MDX thành HTML trong quá trình build
- **Cache_Layer**: Redis (Upstash) dùng để cache response và rate limiting
- **Database**: PostgreSQL (Neon) lưu trữ dữ liệu động (view count, subscribers, metadata)
- **CDN**: Cloudflare CDN phân phối static assets và cache trang
- **Search_Engine**: Module tìm kiếm client-side chạy trên trình duyệt
- **Comment_System**: Giscus - hệ thống bình luận dựa trên GitHub Discussions
- **Newsletter_Service**: Dịch vụ quản lý đăng ký nhận bài viết mới qua email
- **Rate_Limiter**: Module giới hạn số lượng request API dựa trên Redis
- **Image_Optimizer**: Module tối ưu hóa và lazy load hình ảnh
- **TOC_Generator**: Module tự động tạo mục lục từ heading trong bài viết
- **RSS_Generator**: Module tạo RSS feed từ danh sách bài viết
- **Theme_Manager**: Module quản lý chế độ Dark/Light mode

## Yêu cầu

### Yêu cầu 1: Quản lý và hiển thị nội dung bài viết

**User Story:** Là một blogger, tôi muốn viết bài bằng Markdown/MDX và hệ thống tự động render thành trang web đẹp, để tôi có thể tập trung vào nội dung thay vì lo về giao diện.

#### Tiêu chí chấp nhận (Acceptance Criteria)

1. WHEN một file Markdown/MDX được thêm vào thư mục content, THE Content_Engine SHALL parse file đó thành trang HTML hoàn chỉnh với đầy đủ formatting trong quá trình build.
2. WHEN một bài viết chứa code block, THE Content_Engine SHALL áp dụng syntax highlighting với theme phù hợp cho ngôn ngữ lập trình được chỉ định.
3. WHEN một bài viết được render, THE TOC_Generator SHALL tạo mục lục tự động từ các heading (h2, h3, h4) trong bài viết.
4. WHEN một bài viết được render, THE Content_Engine SHALL tính toán và hiển thị thời gian đọc ước tính dựa trên số từ trong bài (trung bình 200 từ/phút).
5. THE Content_Engine SHALL hỗ trợ frontmatter metadata bao gồm: title, description, date, categories, tags, cover image, và draft status.
6. WHEN một bài viết có draft status là true, THE Frontend SHALL không hiển thị bài viết đó trong danh sách công khai.
7. FOR ALL valid Markdown/MDX files, parsing rồi render rồi extract text content SHALL tạo ra nội dung tương đương với nội dung gốc (round-trip property cho content pipeline).

---

### Yêu cầu 2: Phân loại và tổ chức bài viết

**User Story:** Là một độc giả, tôi muốn duyệt bài viết theo chủ đề (categories) và nhãn (tags), để tôi có thể tìm nội dung phù hợp với sở thích.

#### Tiêu chí chấp nhận (Acceptance Criteria)

1. THE Frontend SHALL hiển thị danh sách tất cả categories có sẵn trên trang chuyên mục.
2. THE Frontend SHALL hiển thị danh sách tất cả tags có sẵn trên trang tags.
3. WHEN người dùng chọn một category, THE Frontend SHALL hiển thị tất cả bài viết thuộc category đó, sắp xếp theo ngày mới nhất.
4. WHEN người dùng chọn một tag, THE Frontend SHALL hiển thị tất cả bài viết có gắn tag đó, sắp xếp theo ngày mới nhất.
5. THE Content_Engine SHALL cho phép một bài viết thuộc đúng một category và nhiều tags.
6. WHEN danh sách bài viết vượt quá số lượng trên một trang, THE Frontend SHALL phân trang (pagination) với số bài mỗi trang có thể cấu hình.

---

### Yêu cầu 3: Tìm kiếm bài viết phía client

**User Story:** Là một độc giả, tôi muốn tìm kiếm bài viết theo từ khóa ngay trên trình duyệt, để tôi có thể nhanh chóng tìm được nội dung cần đọc mà không cần chờ server.

#### Tiêu chí chấp nhận (Acceptance Criteria)

1. THE Search_Engine SHALL xây dựng search index từ title, description, tags, và nội dung bài viết trong quá trình build.
2. WHEN người dùng nhập từ khóa vào ô tìm kiếm, THE Search_Engine SHALL trả về kết quả phù hợp trong thời gian dưới 100ms trên client.
3. THE Search_Engine SHALL hỗ trợ tìm kiếm fuzzy (cho phép sai chính tả nhẹ) và tìm kiếm theo prefix.
4. WHEN kết quả tìm kiếm được hiển thị, THE Frontend SHALL highlight từ khóa tìm kiếm trong tiêu đề và mô tả của kết quả.
5. IF không có kết quả nào phù hợp, THEN THE Frontend SHALL hiển thị thông báo "Không tìm thấy kết quả" kèm gợi ý tìm kiếm khác.

---

### Yêu cầu 4: Giao diện Dark/Light Mode và Responsive

**User Story:** Là một độc giả, tôi muốn chuyển đổi giữa chế độ sáng và tối, và đọc blog thoải mái trên mọi thiết bị, để trải nghiệm đọc luôn dễ chịu.

#### Tiêu chí chấp nhận (Acceptance Criteria)

1. THE Theme_Manager SHALL phát hiện system preference (prefers-color-scheme) của người dùng và áp dụng theme tương ứng khi truy cập lần đầu.
2. WHEN người dùng click nút chuyển theme, THE Theme_Manager SHALL chuyển đổi giữa dark mode và light mode ngay lập tức mà không reload trang.
3. THE Theme_Manager SHALL lưu lựa chọn theme của người dùng vào localStorage và áp dụng lại khi truy cập lần sau.
4. THE Frontend SHALL hiển thị đúng layout trên các kích thước màn hình: mobile (< 768px), tablet (768px - 1024px), và desktop (> 1024px).
5. WHILE người dùng đang ở dark mode, THE Frontend SHALL đảm bảo contrast ratio tối thiểu 4.5:1 cho text content theo chuẩn WCAG AA.
6. WHILE người dùng đang ở dark mode, THE Content_Engine SHALL áp dụng syntax highlighting theme tối cho code blocks.

---

### Yêu cầu 5: Hệ thống bình luận (Giscus)

**User Story:** Là một độc giả, tôi muốn bình luận và thảo luận về bài viết, để tôi có thể tương tác với tác giả và cộng đồng.

#### Tiêu chí chấp nhận (Acceptance Criteria)

1. THE Comment_System SHALL tích hợp Giscus widget vào cuối mỗi bài viết.
2. WHEN người dùng muốn bình luận, THE Comment_System SHALL yêu cầu đăng nhập GitHub thông qua Giscus.
3. WHEN một bình luận mới được tạo, THE Comment_System SHALL lưu bình luận đó vào GitHub Discussions của repository tương ứng.
4. THE Comment_System SHALL hỗ trợ Markdown formatting trong bình luận.
5. WHILE người dùng đang ở dark mode, THE Comment_System SHALL hiển thị Giscus widget với theme tối phù hợp.

---

### Yêu cầu 6: Đăng ký Newsletter

**User Story:** Là một độc giả thường xuyên, tôi muốn đăng ký nhận thông báo bài viết mới qua email, để tôi không bỏ lỡ nội dung hay.

#### Tiêu chí chấp nhận (Acceptance Criteria)

1. THE Frontend SHALL hiển thị form đăng ký newsletter với trường email trên trang chủ và cuối mỗi bài viết.
2. WHEN người dùng submit email hợp lệ, THE API_Server SHALL lưu email vào Database và trả về xác nhận thành công.
3. WHEN người dùng submit email không hợp lệ, THE Frontend SHALL hiển thị thông báo lỗi validation cụ thể.
4. IF email đã tồn tại trong hệ thống, THEN THE API_Server SHALL trả về thông báo rằng email đã được đăng ký trước đó.
5. THE API_Server SHALL validate email format theo chuẩn RFC 5322 trước khi lưu vào Database.
6. WHEN một subscriber muốn hủy đăng ký, THE Newsletter_Service SHALL cung cấp link unsubscribe trong mỗi email gửi đi.

---

### Yêu cầu 7: Đếm lượt xem bài viết (View Count)

**User Story:** Là tác giả blog, tôi muốn biết số lượt xem của mỗi bài viết, để tôi có thể đánh giá nội dung nào được quan tâm nhất.

#### Tiêu chí chấp nhận (Acceptance Criteria)

1. WHEN một người dùng truy cập bài viết, THE API_Server SHALL tăng view count của bài viết đó lên 1.
2. THE API_Server SHALL sử dụng Cache_Layer để batch update view count vào Database, tránh ghi trực tiếp mỗi request.
3. THE Frontend SHALL hiển thị số lượt xem trên mỗi bài viết (trang chi tiết và danh sách).
4. THE API_Server SHALL không đếm trùng lượt xem từ cùng một IP trong khoảng thời gian 24 giờ cho cùng một bài viết.
5. IF Database không khả dụng, THEN THE API_Server SHALL tiếp tục phục vụ trang với view count từ Cache_Layer và đồng bộ lại khi Database phục hồi.

---

### Yêu cầu 8: RSS Feed

**User Story:** Là một độc giả sử dụng RSS reader, tôi muốn subscribe RSS feed của blog, để tôi có thể đọc bài mới trong ứng dụng RSS yêu thích.

#### Tiêu chí chấp nhận (Acceptance Criteria)

1. THE RSS_Generator SHALL tạo file RSS 2.0 feed hợp lệ (valid XML) trong quá trình build.
2. THE RSS_Generator SHALL bao gồm 20 bài viết mới nhất trong feed, với đầy đủ title, description, link, pubDate, và category.
3. THE Frontend SHALL cung cấp link đến RSS feed trong thẻ `<head>` của mọi trang và hiển thị icon RSS trên giao diện.
4. FOR ALL bài viết được thêm vào feed, THE RSS_Generator SHALL đảm bảo XML output tuân thủ RSS 2.0 specification (round-trip: parse RSS XML rồi validate structure SHALL pass).

---

### Yêu cầu 9: Tối ưu hóa hình ảnh và hiệu suất

**User Story:** Là một độc giả, tôi muốn trang web tải nhanh và hình ảnh hiển thị mượt mà, để trải nghiệm đọc không bị gián đoạn.

#### Tiêu chí chấp nhận (Acceptance Criteria)

1. THE Image_Optimizer SHALL tự động chuyển đổi hình ảnh sang định dạng WebP/AVIF với nhiều kích thước (srcset) trong quá trình build.
2. THE Frontend SHALL áp dụng lazy loading cho tất cả hình ảnh nằm ngoài viewport ban đầu.
3. THE CDN SHALL cache tất cả static assets (HTML, CSS, JS, images) với cache headers phù hợp (immutable cho hashed assets, stale-while-revalidate cho HTML).
4. THE Frontend SHALL đạt điểm Lighthouse Performance tối thiểu 90/100 trên mobile.
5. THE Image_Optimizer SHALL tạo placeholder blur (LQIP - Low Quality Image Placeholder) cho mỗi hình ảnh để hiển thị trong khi ảnh gốc đang tải.
6. WHEN hình ảnh gốc có kích thước lớn hơn 2000px width, THE Image_Optimizer SHALL resize xuống tối đa 2000px width để giảm dung lượng.

---

### Yêu cầu 10: API Rate Limiting

**User Story:** Là tác giả blog, tôi muốn API backend được bảo vệ khỏi lạm dụng, để hệ thống luôn ổn định và không bị quá tải.

#### Tiêu chí chấp nhận (Acceptance Criteria)

1. THE Rate_Limiter SHALL giới hạn số request từ mỗi IP address dựa trên sliding window algorithm lưu trong Cache_Layer.
2. THE Rate_Limiter SHALL áp dụng giới hạn 100 requests/phút cho các API endpoint công khai.
3. WHEN một IP vượt quá giới hạn, THE API_Server SHALL trả về HTTP 429 (Too Many Requests) kèm header Retry-After chỉ thời gian chờ.
4. IF Cache_Layer không khả dụng, THEN THE Rate_Limiter SHALL cho phép request đi qua (fail-open) và ghi log cảnh báo.
5. THE Rate_Limiter SHALL hỗ trợ cấu hình giới hạn khác nhau cho từng nhóm endpoint (ví dụ: newsletter subscribe có giới hạn thấp hơn).
6. FOR ALL requests trong cùng một sliding window, THE Rate_Limiter SHALL đếm chính xác số request (invariant: count tại bất kỳ thời điểm nào SHALL bằng số request thực tế trong window đó).

---

### Yêu cầu 11: Database và Caching Strategy

**User Story:** Là tác giả blog, tôi muốn hệ thống có chiến lược database và caching hiệu quả, để blog có thể scale khi lượng truy cập tăng.

#### Tiêu chí chấp nhận (Acceptance Criteria)

1. THE Database SHALL có index trên các cột thường xuyên query: post slug, category, tag, created date, và view count.
2. THE Cache_Layer SHALL cache response của các API endpoint đọc (GET) với TTL có thể cấu hình cho từng endpoint.
3. WHEN dữ liệu trong Database thay đổi, THE API_Server SHALL invalidate cache tương ứng trong Cache_Layer.
4. THE Database SHALL sử dụng connection pooling với số connection tối đa có thể cấu hình.
5. THE API_Server SHALL implement health check endpoint trả về trạng thái kết nối Database và Cache_Layer.
6. WHILE hệ thống đang hoạt động, THE API_Server SHALL duy trì response time trung bình dưới 200ms cho các API endpoint đọc (có cache).

---

### Yêu cầu 12: SEO và Social Sharing

**User Story:** Là tác giả blog, tôi muốn bài viết được tối ưu cho công cụ tìm kiếm và hiển thị đẹp khi chia sẻ trên mạng xã hội, để thu hút nhiều độc giả hơn.

#### Tiêu chí chấp nhận (Acceptance Criteria)

1. THE Frontend SHALL tạo meta tags (Open Graph, Twitter Card) đầy đủ cho mỗi bài viết bao gồm: title, description, image, url, và type.
2. THE Frontend SHALL tạo sitemap.xml tự động trong quá trình build, bao gồm tất cả trang công khai.
3. THE Frontend SHALL sử dụng semantic HTML (article, header, nav, main, footer) cho cấu trúc trang.
4. THE Frontend SHALL tạo canonical URL cho mỗi trang để tránh duplicate content.
5. THE Content_Engine SHALL tạo structured data (JSON-LD) cho bài viết theo schema BlogPosting.
6. WHEN một bài viết mới được publish, THE RSS_Generator SHALL cập nhật sitemap.xml và RSS feed trong lần build tiếp theo.

---

### Yêu cầu 13: Deployment và Scalability

**User Story:** Là tác giả blog, tôi muốn hệ thống dễ deploy, miễn phí ban đầu, và có thể scale khi cần, để tôi không phải lo về hạ tầng khi blog phát triển.

#### Tiêu chí chấp nhận (Acceptance Criteria)

1. THE Frontend SHALL được deploy trên Cloudflare Pages với automatic build từ Git repository.
2. THE API_Server SHALL được deploy trên Fly.io hoặc Railway với Docker container.
3. THE API_Server SHALL là stateless, cho phép chạy nhiều instance song song mà không cần shared state ngoài Database và Cache_Layer.
4. THE Blog_System SHALL hoạt động trong free tier của tất cả dịch vụ (Cloudflare Pages, Fly.io/Railway, Neon PostgreSQL, Upstash Redis) cho giai đoạn đầu.
5. WHEN lượng truy cập tăng, THE API_Server SHALL có thể scale horizontal bằng cách thêm instance mới mà không cần thay đổi code.
6. THE Blog_System SHALL sử dụng environment variables cho tất cả cấu hình nhạy cảm (database URL, API keys, Redis URL).
7. IF một API_Server instance gặp lỗi, THEN THE Blog_System SHALL tự động route traffic sang các instance còn lại (health check + load balancing).
