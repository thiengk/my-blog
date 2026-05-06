---
title: "Test Newsletter - Đăng ký nhận bài viết mới"
description: "Bài viết test tính năng newsletter và đăng ký email để nhận thông báo bài viết mới."
date: 2026-05-06
category: "cong-nghe"
tags: ["test", "newsletter", "email"]
draft: false
coverImage: "https://images.unsplash.com/photo-1557200134-90327ee9fafa?w=800&h=400&fit=crop"
---

## Giới thiệu

Đây là bài viết test để kiểm tra tính năng **Newsletter** - đăng ký nhận bài viết mới qua email.

## Tính năng Newsletter

Blog này đã tích hợp tính năng đăng ký newsletter với các đặc điểm:

### ✅ Đã hoàn thành

- **Đăng ký email**: Form đăng ký đơn giản và dễ sử dụng
- **Validation**: Kiểm tra định dạng email hợp lệ
- **Backend API**: Go + PostgreSQL để lưu trữ subscribers
- **Rate limiting**: Bảo vệ khỏi spam và abuse
- **Dark mode support**: Form tự động thích ứng với theme

### 🔄 Đang phát triển

- **Email verification**: Xác nhận email qua link
- **Gửi email tự động**: Thông báo khi có bài viết mới
- **Unsubscribe**: Hủy đăng ký dễ dàng

## Cách sử dụng

1. Cuộn xuống cuối bài viết
2. Nhập email vào form "Đăng ký nhận bài viết mới"
3. Click nút "Đăng ký"
4. Nhận thông báo thành công!

## Công nghệ sử dụng

### Frontend
- **Astro**: Static site generator
- **Svelte**: Component framework cho form
- **TailwindCSS**: Styling

### Backend
- **Go (Gin)**: REST API
- **PostgreSQL**: Database lưu trữ subscribers
- **Redis**: Caching và rate limiting

## Code Example

Đây là cách validate email trong Go:

```go
func ValidateEmail(email string) bool {
    if len(email) == 0 || len(email) > 320 {
        return false
    }
    emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    return emailRegex.MatchString(email)
}
```

Và trong Svelte:

```javascript
function validateEmail(value) {
    const emailRegex = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;
    return emailRegex.test(value.trim());
}
```

## Kết luận

Tính năng newsletter đã hoạt động tốt! Hãy thử đăng ký để nhận thông báo khi có bài viết mới nhé.

### Các bước tiếp theo

- [ ] Tích hợp email service (SendGrid, Mailgun, hoặc AWS SES)
- [ ] Tạo email template đẹp mắt
- [ ] Thêm tính năng quản lý subscribers
- [ ] Analytics và tracking

---

**Cảm ơn bạn đã đọc!** Đừng quên đăng ký newsletter ở phía dưới để không bỏ lỡ bài viết tiếp theo! 📧
