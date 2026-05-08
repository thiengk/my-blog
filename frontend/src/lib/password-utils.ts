/**
 * Password Protection Utilities
 *
 * Helper functions for password-protected blog posts.
 * Handles password classification and session-based unlock persistence.
 *
 * Validates: Requirements 1.2, 1.4, 4.1, 4.2, 4.4, 4.5
 */

/**
 * Kiểm tra bài viết có được bảo vệ bằng mật khẩu hay không.
 * Trả về true nếu password tồn tại, không rỗng, và không chỉ chứa whitespace.
 */
export function isProtectedPost(password: string | undefined): boolean {
  return password !== undefined && password.trim().length > 0;
}

/**
 * Tạo sessionStorage key cho bài viết.
 * Format: "protected-post:{slug}"
 */
export function getStorageKey(slug: string): string {
  return `protected-post:${slug}`;
}

/**
 * Kiểm tra sessionStorage có khả dụng hay không.
 * Sử dụng try/catch để xử lý trường hợp bị chặn (private browsing, etc.)
 */
export function isSessionStorageAvailable(): boolean {
  try {
    const testKey = '__test__';
    sessionStorage.setItem(testKey, '1');
    sessionStorage.removeItem(testKey);
    return true;
  } catch {
    return false;
  }
}

/**
 * Kiểm tra bài viết đã được mở khóa trong session hiện tại.
 * Trả về false nếu sessionStorage không khả dụng.
 */
export function isUnlockedInSession(slug: string): boolean {
  if (!isSessionStorageAvailable()) {
    return false;
  }
  try {
    return sessionStorage.getItem(getStorageKey(slug)) === 'true';
  } catch {
    return false;
  }
}

/**
 * Lưu trạng thái mở khóa vào sessionStorage.
 * No-op nếu sessionStorage không khả dụng.
 */
export function saveUnlockState(slug: string): void {
  if (!isSessionStorageAvailable()) {
    return;
  }
  try {
    sessionStorage.setItem(getStorageKey(slug), 'true');
  } catch {
    // No-op: sessionStorage unavailable
  }
}
