# Implementation Plan

- [x] 1. Database migration và config



  - [x] 1.1 Tạo migration file cho meal tables

    - Tạo `backend/migrations/000007_create_meal_tables.up.sql` với 3 tables: meal_members, meal_participations, meal_payments
    - Tạo `backend/migrations/000007_create_meal_tables.down.sql` để rollback
    - Thêm indexes cho foreign keys và query patterns (member_id, meal_type, payment_date)
    - _Requirements: 1.1, 1.2, 1.5, 2.1, 2.2, 4.1_

  - [x] 1.2 Thêm MEAL_GROUP_SECRET vào config


    - Thêm field `MealGroupSecret` vào `internal/config/config.go`
    - Đọc từ env var `MEAL_GROUP_SECRET`
    - Cập nhật `.env.example` với biến mới
    - _Requirements: 6.5, 6.6_

- [x] 2. Backend service layer



  - [x] 2.1 Tạo `internal/service/meal.go` với interface và struct


    - Định nghĩa `MealService` interface với tất cả methods
    - Định nghĩa Go structs: MealMember, MealParticipation, MealPayment, NextPayerResult, NextPayer, MemberStats
    - Implement constructor `NewMealService(db *pgxpool.Pool) MealService`
    - _Requirements: 1.1, 2.1, 3.1, 4.1_

  - [x] 2.2 Implement member CRUD trong meal service


    - `GetMembers`: query tất cả members active (deleted_at IS NULL)
    - `CreateMember`: insert member mới, check duplicate name (return error 409 nếu trùng)
    - `UpdateMember`: update tên và is_active status
    - `DeleteMember`: soft delete (set deleted_at = NOW())
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5_

  - [x] 2.3 Implement participation management


    - `GetParticipations`: query tất cả participations JOIN member name
    - `UpdateParticipations`: upsert participation record cho member + meal_type
    - Auto-create participation records khi thêm member mới (default: participating = true cho cả breakfast và lunch)
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5_

  - [x] 2.4 Implement scheduling algorithm (GetNextPayer)


    - Query members active + participating cho mỗi meal_type
    - Count payments, get last_paid_at
    - Order by: payment_count ASC, last_paid_at ASC NULLS FIRST, id ASC
    - Return NextPayerResult với breakfast và lunch payer
    - Handle case: không có participants → return null cho meal đó
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5_

  - [x] 2.5 Implement payment recording và undo


    - `RecordPayment`: insert payment record với member_id, meal_type, payment_date
    - `UndoPayment`: delete payment nếu created_at trong vòng 24h, return error nếu quá hạn
    - `GetPayments`: query payments JOIN member name, ORDER BY payment_date DESC, paginated (limit + offset)
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5_


  - [x] 2.6 Implement GetStats

    - Query count payments GROUP BY member_id, meal_type
    - Return breakfast_count, lunch_count, total_count cho mỗi member active
    - _Requirements: 3.6, 5.2_

- [x] 3. Backend service tests


  - [x] 3.1 Viết unit tests cho meal service


    - Tạo `internal/service/meal_test.go`
    - Test scheduling algorithm: 1 member, multiple members equal count, new member (count=0), no participants
    - Test undo payment: within 24h (success), expired (error)
    - Test create member: duplicate name (error), valid name (success)
    - Test soft delete: member không còn xuất hiện trong GetMembers
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 4.5, 1.4, 1.5_

- [x] 4. Backend handler và auth middleware


  - [x] 4.1 Tạo auth middleware cho meal endpoints


    - Tạo `internal/middleware/meal_auth.go`
    - Check header `X-Group-Secret` against config `MealGroupSecret`
    - Return 401 nếu missing hoặc sai
    - _Requirements: 6.5, 6.6_

  - [x] 4.2 Tạo `internal/handler/meal.go` với tất cả endpoints


    - Implement MealHandler struct với MealService dependency
    - GET `/api/meals/members` → GetMembers
    - POST `/api/meals/members` → CreateMember (body: `{"name": "..."}`)
    - PUT `/api/meals/members/:id` → UpdateMember (body: `{"name": "...", "is_active": true}`)
    - DELETE `/api/meals/members/:id` → DeleteMember
    - GET `/api/meals/participations` → GetParticipations
    - PUT `/api/meals/participations` → UpdateParticipations (body: `{"member_id": 1, "breakfast": true, "lunch": false}`)
    - GET `/api/meals/next-payer` → GetNextPayer
    - GET `/api/meals/payments?limit=10&offset=0` → GetPayments
    - POST `/api/meals/payments` → RecordPayment (body: `{"member_id": 1, "meal_type": "breakfast", "date": "2026-05-12"}`)
    - DELETE `/api/meals/payments/:id` → UndoPayment
    - GET `/api/meals/stats` → GetStats
    - _Requirements: 1.1, 2.1, 3.5, 4.1, 4.4, 5.2_

  - [x] 4.3 Register meal routes trong main.go


    - Import meal service và handler
    - Tạo MealService instance với dbPool
    - Tạo MealHandler instance
    - Register routes với auth middleware group
    - _Requirements: 6.5_

- [x] 5. Backend handler tests


  - [x] 5.1 Viết handler tests cho meal endpoints


    - Tạo `internal/handler/meal_test.go`
    - Test auth middleware: valid secret (200), invalid secret (401), missing header (401)
    - Test request validation: empty name (400), invalid meal_type (400), invalid id (400)
    - Test response format cho mỗi endpoint
    - _Requirements: 6.5, 6.6_

- [x] 6. Frontend auth gate component


  - [x] 6.1 Tạo `src/components/MealAuth.svelte`


    - Svelte 5 runes: `$state` cho `authenticated`, `secret`, `error`
    - On mount: check sessionStorage key `meal-group-secret`
    - Form nhập mật khẩu nhóm, validate bằng cách gọi GET `/api/meals/members` với secret
    - Nếu 200: lưu secret vào sessionStorage, set authenticated = true
    - Nếu 401: hiển thị error "Mật khẩu nhóm không đúng"
    - Export function `getSecret()` để các component con dùng khi gọi API
    - _Requirements: 6.1, 6.2, 6.3, 6.4_

- [x] 7. Frontend dashboard component


  - [x] 7.1 Tạo `src/components/MealDashboard.svelte`


    - Fetch và hiển thị next payer cho breakfast và lunch (card lớn, nổi bật)
    - Nút "Xác nhận đã trả" cho mỗi bữa → gọi POST `/api/meals/payments` với next payer info
    - Nút "Người khác trả" → dropdown chọn member khác (override)
    - Hiển thị bảng stats: tên, số lần trả sáng, trưa, tổng
    - Hiển thị 10 payments gần nhất với nút undo cho payment mới nhất (< 24h)
    - Auto-refresh data sau mỗi action (confirm, undo)
    - _Requirements: 3.5, 4.1, 4.2, 4.3, 4.5, 5.1, 5.2, 5.3, 5.4_

  - [x] 7.2 Tạo `src/components/MemberManager.svelte`


    - Form thêm thành viên mới (input name + submit)
    - Danh sách thành viên với toggle active/inactive
    - Checkbox breakfast/lunch participation cho mỗi thành viên
    - Nút xóa (soft delete) với confirm dialog
    - Hiển thị error messages (duplicate name, etc.)
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 2.2, 2.3, 2.5_

- [x] 8. Frontend page và integration


  - [x] 8.1 Tạo page `src/pages/internal/meals.astro`


    - Import MealAuth và MealDashboard components
    - Sử dụng layout hiện tại của blog
    - MealAuth wraps toàn bộ content (gate pattern)
    - Bên trong gate: tabs hoặc sections cho Dashboard và Quản lý thành viên
    - Responsive layout (mobile-friendly)
    - Không link trang này từ navigation chính (internal only, truy cập bằng URL trực tiếp)
    - _Requirements: 5.5, 6.1, 6.2_

  - [x] 8.2 Tạo `src/lib/meal-api.ts` utility cho API calls


    - Helper function `mealFetch(path, options)` tự động thêm X-Group-Secret header từ sessionStorage
    - Typed functions: getMembers, createMember, updateMember, deleteMember
    - Typed functions: getParticipations, updateParticipations
    - Typed functions: getNextPayer, getPayments, recordPayment, undoPayment, getStats
    - Error handling: nếu 401 → clear sessionStorage, redirect về auth form
    - _Requirements: 6.1, 6.5_

- [x] 9. Final verification

  - [x] 9.1 Verify backend builds và tests pass



    - Run `go build ./...` trong backend directory
    - Run `go test ./...` để verify tất cả tests pass
    - _Requirements: all_


  - [x] 9.2 Verify frontend builds

    - Run `npm run build` trong frontend directory
    - Verify trang `/internal/meals` được generate đúng
    - _Requirements: all_
