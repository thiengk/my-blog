# Design Document: Meal Payment Scheduler

## Overview

Internal tool tích hợp vào blog hiện tại, cho phép nhóm làm blog quản lý lịch ăn sáng/trưa và tự động xếp lịch thanh toán công bằng. Hệ thống sử dụng thuật toán round-robin có trọng số dựa trên số lần thanh toán và thời gian thanh toán gần nhất.

Tính năng được triển khai dưới dạng:
- Frontend: Trang `/internal/meals` (Astro page + Svelte 5 interactive components)
- Backend: API endpoints `/api/meals/*` (Go/Gin, cùng server hiện tại)
- Database: Thêm tables vào PostgreSQL hiện có
- Auth: Shared secret (header-based cho API, sessionStorage cho frontend)

## Architecture

```
┌─────────────────────────────────────────┐
│         Frontend (Astro + Svelte 5)     │
│                                         │
│  /internal/meals                        │
│  ├── MealDashboard.svelte (main view)   │
│  ├── MemberManager.svelte              │
│  ├── PaymentHistory.svelte             │
│  └── MealAuth.svelte (gate)            │
│                                         │
└──────────────────┬──────────────────────┘
                   │ fetch + X-Group-Secret header
                   ▼
┌─────────────────────────────────────────┐
│         Backend (Go / Gin)              │
│                                         │
│  /api/meals/*                           │
│  ├── MealHandler (HTTP layer)           │
│  ├── MealService (business logic)       │
│  │   └── Scheduling algorithm           │
│  └── Auth middleware (X-Group-Secret)   │
│                                         │
└──────────────────┬──────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────┐
│         PostgreSQL (Neon)               │
│                                         │
│  meal_members                           │
│  meal_participations                    │
│  meal_payments                          │
│                                         │
└─────────────────────────────────────────┘
```

## Components and Interfaces

### Backend API Endpoints

| Method | Path | Mô tả | Auth |
|--------|------|--------|------|
| GET | `/api/meals/members` | Lấy danh sách thành viên | X-Group-Secret |
| POST | `/api/meals/members` | Thêm thành viên | X-Group-Secret |
| PUT | `/api/meals/members/:id` | Cập nhật thành viên (tên, status) | X-Group-Secret |
| DELETE | `/api/meals/members/:id` | Soft delete thành viên | X-Group-Secret |
| GET | `/api/meals/participations` | Lấy cấu hình tham gia bữa ăn | X-Group-Secret |
| PUT | `/api/meals/participations` | Cập nhật cấu hình tham gia | X-Group-Secret |
| GET | `/api/meals/next-payer` | Lấy người thanh toán tiếp theo (sáng + trưa) | X-Group-Secret |
| GET | `/api/meals/payments` | Lấy lịch sử thanh toán (paginated) | X-Group-Secret |
| POST | `/api/meals/payments` | Ghi nhận thanh toán | X-Group-Secret |
| DELETE | `/api/meals/payments/:id` | Undo thanh toán (trong 24h) | X-Group-Secret |
| GET | `/api/meals/stats` | Thống kê số lần thanh toán mỗi người | X-Group-Secret |

### Backend Service Interface

```go
// MealService defines the business logic for meal payment scheduling.
type MealService interface {
    // Members
    GetMembers(ctx context.Context) ([]MealMember, error)
    CreateMember(ctx context.Context, name string) (*MealMember, error)
    UpdateMember(ctx context.Context, id int64, name string, isActive bool) (*MealMember, error)
    DeleteMember(ctx context.Context, id int64) error

    // Participations
    GetParticipations(ctx context.Context) ([]MealParticipation, error)
    UpdateParticipations(ctx context.Context, memberId int64, breakfast bool, lunch bool) error

    // Scheduling
    GetNextPayer(ctx context.Context) (*NextPayerResult, error)

    // Payments
    GetPayments(ctx context.Context, limit int, offset int) ([]MealPayment, error)
    RecordPayment(ctx context.Context, memberId int64, mealType string, date time.Time) (*MealPayment, error)
    UndoPayment(ctx context.Context, paymentId int64) error

    // Stats
    GetStats(ctx context.Context) ([]MemberStats, error)
}
```

### Frontend Components

#### MealAuth.svelte
- Gate component kiểm tra sessionStorage cho `meal-group-secret`
- Nếu chưa auth: hiển thị form nhập mật khẩu nhóm
- Nếu đã auth: render children (dashboard)
- Lưu secret vào sessionStorage để gửi kèm mọi API request

#### MealDashboard.svelte
- Component chính, hiển thị:
  - Card "Người trả tiếp" cho bữa sáng và trưa
  - Nút "Xác nhận đã trả" để ghi nhận thanh toán nhanh
  - Bảng thống kê số lần trả của mỗi người
  - Lịch sử thanh toán gần nhất (10 records)

#### MemberManager.svelte
- CRUD thành viên (thêm, sửa tên, toggle active/inactive)
- Cấu hình tham gia bữa ăn (checkbox sáng/trưa cho mỗi thành viên)

#### PaymentHistory.svelte
- Danh sách lịch sử thanh toán với pagination
- Nút undo cho payment gần nhất (trong 24h)

## Data Models

### Database Tables

#### meal_members
| Column | Type | Mô tả |
|--------|------|-------|
| id | BIGSERIAL | Primary key |
| name | VARCHAR(100) | Tên thành viên (UNIQUE) |
| is_active | BOOLEAN | Trạng thái active (default: true) |
| created_at | TIMESTAMPTZ | Thời gian tạo |
| updated_at | TIMESTAMPTZ | Cập nhật cuối |
| deleted_at | TIMESTAMPTZ | Soft delete timestamp (NULL = active) |

#### meal_participations
| Column | Type | Mô tả |
|--------|------|-------|
| id | BIGSERIAL | Primary key |
| member_id | BIGINT | FK → meal_members.id |
| meal_type | VARCHAR(20) | 'breakfast' hoặc 'lunch' |
| is_participating | BOOLEAN | Có tham gia bữa này không (default: true) |
| created_at | TIMESTAMPTZ | Thời gian tạo |
| updated_at | TIMESTAMPTZ | Cập nhật cuối |

Constraint: UNIQUE(member_id, meal_type)

#### meal_payments
| Column | Type | Mô tả |
|--------|------|-------|
| id | BIGSERIAL | Primary key |
| member_id | BIGINT | FK → meal_members.id |
| meal_type | VARCHAR(20) | 'breakfast' hoặc 'lunch' |
| payment_date | DATE | Ngày thanh toán |
| created_at | TIMESTAMPTZ | Thời gian ghi nhận |

### Go Structs

```go
type MealMember struct {
    ID        int64      `json:"id"`
    Name      string     `json:"name"`
    IsActive  bool       `json:"is_active"`
    CreatedAt time.Time  `json:"created_at"`
    UpdatedAt time.Time  `json:"updated_at"`
    DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type MealParticipation struct {
    ID              int64  `json:"id"`
    MemberID        int64  `json:"member_id"`
    MemberName      string `json:"member_name"`
    MealType        string `json:"meal_type"`
    IsParticipating bool   `json:"is_participating"`
}

type MealPayment struct {
    ID          int64     `json:"id"`
    MemberID    int64     `json:"member_id"`
    MemberName  string    `json:"member_name"`
    MealType    string    `json:"meal_type"`
    PaymentDate string    `json:"payment_date"`
    CreatedAt   time.Time `json:"created_at"`
}

type NextPayerResult struct {
    Breakfast *NextPayer `json:"breakfast"`
    Lunch     *NextPayer `json:"lunch"`
}

type NextPayer struct {
    MemberID     int64  `json:"member_id"`
    MemberName   string `json:"member_name"`
    PaymentCount int64  `json:"payment_count"`
    LastPaidAt   *time.Time `json:"last_paid_at,omitempty"`
}

type MemberStats struct {
    MemberID       int64  `json:"member_id"`
    MemberName     string `json:"member_name"`
    BreakfastCount int64  `json:"breakfast_count"`
    LunchCount     int64  `json:"lunch_count"`
    TotalCount     int64  `json:"total_count"`
}
```

## Scheduling Algorithm

Thuật toán xếp lịch thanh toán công bằng cho mỗi bữa (breakfast/lunch):

```
Input: danh sách thành viên tham gia bữa đó (active + participating)
Output: thành viên được chọn thanh toán tiếp theo

1. Lấy tất cả thành viên active + participating cho meal_type
2. Với mỗi thành viên, đếm số lần thanh toán (payment_count) cho meal_type đó
3. Tìm min(payment_count) trong danh sách
4. Lọc ra các thành viên có payment_count == min
5. Nếu chỉ 1 người → chọn người đó
6. Nếu nhiều người cùng min:
   a. Lấy last_paid_at (thời gian thanh toán gần nhất) cho mỗi người
   b. Chọn người có last_paid_at xa nhất (NULL = chưa bao giờ trả → ưu tiên cao nhất)
   c. Nếu vẫn bằng nhau → chọn người có id nhỏ nhất (deterministic thay vì random)
```

SQL query tương đương:

```sql
SELECT m.id, m.name,
       COALESCE(COUNT(p.id), 0) AS payment_count,
       MAX(p.payment_date) AS last_paid_at
FROM meal_members m
JOIN meal_participations mp ON mp.member_id = m.id
LEFT JOIN meal_payments p ON p.member_id = m.id AND p.meal_type = $1
WHERE m.is_active = true
  AND m.deleted_at IS NULL
  AND mp.meal_type = $1
  AND mp.is_participating = true
GROUP BY m.id, m.name
ORDER BY payment_count ASC, last_paid_at ASC NULLS FIRST, m.id ASC
LIMIT 1;
```

## Error Handling

| Scenario | HTTP Status | Response |
|----------|-------------|----------|
| Missing/invalid X-Group-Secret | 401 | `{"error": "unauthorized"}` |
| Member name trùng | 409 | `{"error": "member already exists"}` |
| Member not found | 404 | `{"error": "member not found"}` |
| Invalid meal_type (not breakfast/lunch) | 400 | `{"error": "invalid meal type"}` |
| Undo payment quá 24h | 400 | `{"error": "can only undo payments within 24 hours"}` |
| Payment not found | 404 | `{"error": "payment not found"}` |
| No participants for meal | 200 | `{"breakfast": null}` hoặc `{"lunch": null}` |
| Database error | 500 | `{"error": "internal server error"}` |
| Empty member name | 400 | `{"error": "name is required"}` |

## Testing Strategy

### Backend Tests

1. **Unit tests cho MealService** (mock database):
   - Test scheduling algorithm với các scenarios: 1 người, nhiều người cùng count, thành viên mới (count=0)
   - Test undo payment logic (within 24h, expired)
   - Test member CRUD (duplicate name, soft delete)

2. **Integration tests** (test database):
   - Test full flow: create members → set participations → get next payer → record payment → verify next payer changes
   - Test concurrent payments

3. **Handler tests** (mock service):
   - Test auth middleware (valid/invalid/missing secret)
   - Test request validation
   - Test response format

### Frontend Tests

1. **Component tests** (vitest + testing-library):
   - MealAuth: test gate behavior, sessionStorage persistence
   - MealDashboard: test data display, confirm payment flow
   - MemberManager: test CRUD operations

2. **API integration** (mock fetch):
   - Test API calls include correct headers
   - Test error handling (401, 500)

## Configuration

Thêm environment variable:

```env
# Meal scheduler shared secret (required for meal API endpoints)
MEAL_GROUP_SECRET=your-shared-secret-here
```

Config struct addition:

```go
// In config.go
MealGroupSecret string // from MEAL_GROUP_SECRET env var
```
