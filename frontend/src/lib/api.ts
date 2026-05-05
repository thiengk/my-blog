/**
 * API utility for communicating with the Go backend.
 * Handles view count recording, fetching, and graceful degradation.
 *
 * Validates: Requirements 7.1, 7.3
 */

/** Base URL for API calls - configurable via environment variable */
const API_URL = import.meta.env.PUBLIC_API_URL || '';

/** Default timeout for API requests (in milliseconds) */
const DEFAULT_TIMEOUT = 5000;

/**
 * Create an AbortController with a timeout.
 * @param ms - Timeout in milliseconds
 * @returns AbortSignal that will abort after the specified timeout
 */
function createTimeoutSignal(ms: number = DEFAULT_TIMEOUT): AbortSignal {
  const controller = new AbortController();
  setTimeout(() => controller.abort(), ms);
  return controller.signal;
}

/**
 * Record a page view for a blog post.
 * Fire-and-forget: errors are silently caught for graceful degradation.
 *
 * @param slug - The blog post slug to record a view for
 */
export async function recordView(slug: string): Promise<void> {
  try {
    await fetch(`${API_URL}/api/views/${encodeURIComponent(slug)}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      signal: createTimeoutSignal(),
    });
  } catch {
    // Graceful degradation: silently ignore errors
    // View recording is non-critical functionality
  }
}

/**
 * Get the view count for a single blog post.
 * Returns 0 if the API is unavailable (graceful degradation).
 *
 * @param slug - The blog post slug to get the view count for
 * @returns The view count, or 0 if unavailable
 */
export async function getViewCount(slug: string): Promise<number> {
  try {
    const response = await fetch(
      `${API_URL}/api/views/${encodeURIComponent(slug)}`,
      {
        method: 'GET',
        signal: createTimeoutSignal(),
      }
    );

    if (!response.ok) {
      return 0;
    }

    const data = await response.json();
    return typeof data.count === 'number' ? data.count : 0;
  } catch {
    // Graceful degradation: return 0 when API is unavailable
    return 0;
  }
}

/**
 * Get view counts for multiple blog posts in a single request.
 * Returns an empty object if the API is unavailable (graceful degradation).
 *
 * @param slugs - Array of blog post slugs to get view counts for
 * @returns Record mapping slug to view count, or empty object if unavailable
 */
export async function getBulkViewCounts(
  slugs: string[]
): Promise<Record<string, number>> {
  if (slugs.length === 0) {
    return {};
  }

  try {
    const params = new URLSearchParams({
      slugs: slugs.map(s => encodeURIComponent(s)).join(','),
    });

    const response = await fetch(`${API_URL}/api/views?${params.toString()}`, {
      method: 'GET',
      signal: createTimeoutSignal(),
    });

    if (!response.ok) {
      return {};
    }

    const data = await response.json();
    if (data.counts && typeof data.counts === 'object') {
      return data.counts as Record<string, number>;
    }

    return {};
  } catch {
    // Graceful degradation: return empty counts when API is unavailable
    return {};
  }
}
