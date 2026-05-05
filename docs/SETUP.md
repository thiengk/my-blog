# Hướng dẫn chạy dự án Local

## Yêu cầu hệ thống

| Công cụ | Phiên bản | Mục đích | Link tải |
|---------|-----------|----------|----------|
| **Node.js** | 20+ | Chạy frontend Astro | https://nodejs.org |
| **Go** | 1.22+ | Chạy backend API | https://go.dev/dl |
| **Docker Desktop** | Latest | Chạy PostgreSQL + Redis local | https://docker.com/products/docker-desktop |

## Các bước cài đặt

### Bước 1 — Clone và cấu hình environment

```bash
# Copy file environment variables
cp .env.example .env
```

Mở file `.env` và kiểm tra các giá trị mặc định. Với local development, các giá trị mặc định đã đủ để chạy.

### Bước 2 — Khởi động Database (PostgreSQL + Redis)

```bash
docker compose up -d
```

Lệnh này sẽ khởi động:
- **PostgreSQL 16** tại `localhost:5432` (user: `blog_user`, password: `blog_password`, db: `blog_dev`)
- **Redis 7** tại `localhost:6379`

Kiểm tra trạng thái:
```bash
docker compose ps
```

### Bước 3 — Setup Backend (Go API)

```bash
cd backend

# Tải dependencies
go mod tidy

# Chạy database migrations
make migrate

# Khởi động server
make run
```

Backend sẽ chạy tại: **http://localhost:8080**

Kiểm tra health: `curl http://localhost:8080/api/health`

### Bước 4 — Setup Frontend (Astro)

```bash
cd frontend

# Cài dependencies
npm install

# Khởi động dev server
npm run dev
```

Frontend sẽ chạy tại: **http://localhost:4321**

## Công cụ bổ sung (Optional)

| Công cụ | Mục đích | Cài đặt |
|---------|----------|---------|
| **golang-migrate** | Chạy DB migrations | `go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest` |
| **golangci-lint** | Lint Go code | https://golangci-lint.run/usage/install |

## Các lệnh thường dùng

### Frontend

| Lệnh | Mô tả |
|------|-------|
| `npm run dev` | Chạy dev server (hot reload) |
| `npm run build` | Build static site cho production |
| `npm run preview` | Preview bản build |

### Backend

| Lệnh | Mô tả |
|------|-------|
| `make run` | Chạy server |
| `make build` | Build binary |
| `make test` | Chạy unit tests |
| `make migrate` | Chạy migrations |
| `make migrate-down` | Rollback migration cuối |
| `make docker-build` | Build Docker image |
| `make lint` | Lint code |
| `make fmt` | Format code |

### Docker

| Lệnh | Mô tả |
|------|-------|
| `docker compose up -d` | Khởi động PostgreSQL + Redis |
| `docker compose down` | Dừng services |
| `docker compose logs -f` | Xem logs |
| `docker compose --profile full up -d` | Khởi động cả backend container |

## Xử lý lỗi thường gặp

### PostgreSQL connection refused
- Kiểm tra Docker đang chạy: `docker compose ps`
- Đợi health check pass: `docker compose logs postgres`

### Redis connection refused
- Tương tự PostgreSQL, kiểm tra Docker container

### Migration failed
- Đảm bảo PostgreSQL đã sẵn sàng trước khi chạy migrate
- Kiểm tra `DATABASE_URL` trong file `.env`

### Frontend build error
- Xóa cache: `rm -rf frontend/node_modules frontend/.astro`
- Cài lại: `cd frontend && npm install`

## Ports sử dụng

| Service | Port |
|---------|------|
| Frontend (Astro dev) | 4321 |
| Backend (Go API) | 8080 |
| PostgreSQL | 5432 |
| Redis | 6379 |
