# Personal Blog

Blog cá nhân đơn giản được xây dựng với Astro (frontend) và Go (backend).

## Tính năng

✅ **Đã có:**
- Hiển thị danh sách bài viết với pagination
- Xem chi tiết bài viết (Markdown)
- Lọc theo categories và tags
- Tìm kiếm bài viết
- Dark mode
- View counter (đếm lượt xem)
- SEO optimization
- RSS feed
- Responsive design

## Công nghệ sử dụng

### Frontend
- **Astro** - Static site generator
- **Svelte** - Interactive components
- **TailwindCSS** - Styling
- **TypeScript** - Type safety

### Backend
- **Go (Gin)** - REST API
- **PostgreSQL** - Database
- **Redis** - Caching & rate limiting
- **Docker** - Containerization

## Cấu trúc thư mục

```
my-blog/
├── frontend/           # Astro frontend
│   ├── src/
│   │   ├── components/ # Svelte & Astro components
│   │   ├── content/    # Blog posts (Markdown)
│   │   ├── layouts/    # Page layouts
│   │   ├── pages/      # Routes
│   │   └── styles/     # Global styles
│   └── public/         # Static assets
│
├── backend/            # Go backend
│   ├── cmd/server/     # Main application
│   ├── internal/       # Internal packages
│   │   ├── config/     # Configuration
│   │   ├── database/   # Database connections
│   │   ├── handler/    # HTTP handlers
│   │   ├── middleware/ # Middlewares
│   │   └── service/    # Business logic
│   └── migrations/     # Database migrations
│
├── docs/               # Documentation
└── docker-compose.yml  # Docker services
```

## Hướng dẫn chạy

Xem chi tiết trong [docs/SETUP.md](docs/SETUP.md)

### Tóm tắt nhanh

1. **Khởi động database:**
   ```bash
   docker compose up -d
   ```

2. **Chạy backend:**
   ```bash
   cd backend
   go mod tidy
   make migrate
   make run
   ```

3. **Chạy frontend:**
   ```bash
   cd frontend
   npm install
   npm run dev
   ```

4. **Truy cập:**
   - Frontend: http://localhost:4321
   - Backend API: http://localhost:8080

## Viết bài mới

Tạo file `.md` trong `frontend/src/content/blog/`:

```markdown
---
title: "Tiêu đề bài viết"
description: "Mô tả ngắn"
date: 2026-05-06
category: "cong-nghe"
tags: ["tag1", "tag2"]
draft: false
---

## Nội dung bài viết

Viết nội dung ở đây...
```

## License

MIT
