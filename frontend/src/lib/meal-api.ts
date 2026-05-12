/**
 * Meal Scheduler API utility.
 * Handles all API calls for the internal meal payment scheduler feature.
 * Automatically includes X-Group-Secret header from sessionStorage.
 */

const API_URL = import.meta.env.PUBLIC_API_URL || '';
const STORAGE_KEY = 'meal-group-secret';

/**
 * Get the stored group secret from sessionStorage.
 */
function getSecret(): string {
  try {
    return sessionStorage.getItem(STORAGE_KEY) || '';
  } catch {
    return '';
  }
}

/**
 * Clear stored secret and reload page (force re-auth).
 */
function clearAuth(): void {
  try {
    sessionStorage.removeItem(STORAGE_KEY);
  } catch {
    // ignore
  }
  window.location.reload();
}

/**
 * Generic fetch wrapper that adds auth header and handles errors.
 * @param path - API path (e.g., '/api/meals/members')
 * @param options - Fetch options
 * @returns Parsed JSON response
 */
export async function mealFetch(path: string, options: RequestInit = {}): Promise<any> {
  const secret = getSecret();
  if (!secret) {
    clearAuth();
    throw new Error('Chưa đăng nhập');
  }

  const response = await fetch(`${API_URL}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      'X-Group-Secret': secret,
      ...(options.headers || {}),
    },
  });

  if (response.status === 401) {
    clearAuth();
    throw new Error('Phiên đăng nhập hết hạn');
  }

  if (!response.ok) {
    const data = await response.json().catch(() => ({}));
    throw new Error(data.error || `Lỗi ${response.status}`);
  }

  return response.json();
}
