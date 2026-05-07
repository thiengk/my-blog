<script>
  /**
   * EngagementCounter Component
   * Hiển thị số lượng likes, comments, shares cho một bài viết.
   * Sử dụng cho cả post detail view và post list view.
   * Graceful degradation: hiển thị 0 khi API unavailable.
   *
   * Validates: Requirements 4.1, 4.2, 4.5
   */

  /** Base URL for API calls */
  const API_URL = import.meta.env.PUBLIC_API_URL || '';

  /** Default timeout for API requests (in milliseconds) */
  const DEFAULT_TIMEOUT = 5000;

  /**
   * @typedef {Object} Props
   * @property {string} slug - Post slug để fetch engagement counts
   * @property {number} [initialLikes] - Số likes ban đầu (từ bulk API)
   * @property {number} [initialComments] - Số comments ban đầu (từ bulk API)
   * @property {number} [initialShares] - Số shares ban đầu (từ bulk API)
   */

  /** @type {Props} */
  let {
    slug,
    initialLikes = undefined,
    initialComments = undefined,
    initialShares = undefined,
  } = $props();

  // State
  let likes = $state(initialLikes ?? 0);
  let comments = $state(initialComments ?? 0);
  let shares = $state(initialShares ?? 0);
  let isLoading = $state(true);

  /**
   * Kiểm tra xem đã có initial data từ props chưa.
   * Nếu có thì không cần fetch từ API.
   */
  let hasInitialData = $derived(
    initialLikes !== undefined ||
    initialComments !== undefined ||
    initialShares !== undefined
  );

  $effect(() => {
    if (!slug) return;

    // Nếu đã có initial data từ bulk API, không cần fetch lại
    if (hasInitialData) {
      likes = initialLikes ?? 0;
      comments = initialComments ?? 0;
      shares = initialShares ?? 0;
      isLoading = false;
      return;
    }

    let cancelled = false;

    async function fetchCounts() {
      try {
        const controller = new AbortController();
        const timeoutId = setTimeout(() => controller.abort(), DEFAULT_TIMEOUT);

        const response = await fetch(
          `${API_URL}/api/engagement/${encodeURIComponent(slug)}`,
          {
            method: 'GET',
            signal: controller.signal,
          }
        );

        clearTimeout(timeoutId);

        if (!cancelled) {
          if (response.ok) {
            const data = await response.json();
            likes = typeof data.likes === 'number' ? data.likes : 0;
            comments = typeof data.comments === 'number' ? data.comments : 0;
            shares = typeof data.shares === 'number' ? data.shares : 0;
          } else {
            // API trả về lỗi - graceful degradation
            likes = 0;
            comments = 0;
            shares = 0;
          }
          isLoading = false;
        }
      } catch {
        // API unavailable - graceful degradation: hiển thị 0
        if (!cancelled) {
          likes = 0;
          comments = 0;
          shares = 0;
          isLoading = false;
        }
      }
    }

    fetchCounts();

    return () => {
      cancelled = true;
    };
  });

  /**
   * Format số lượng theo locale Việt Nam.
   * @param {number} count
   * @returns {string}
   */
  function formatCount(count) {
    return count.toLocaleString('vi-VN');
  }

  // Derived display values
  let likesDisplay = $derived(isLoading ? '--' : formatCount(likes));
  let commentsDisplay = $derived(isLoading ? '--' : formatCount(comments));
  let sharesDisplay = $derived(isLoading ? '--' : formatCount(shares));
</script>

<div class="engagement-counter flex items-center gap-4 text-sm text-gray-600 dark:text-gray-400" aria-label="Thống kê tương tác bài viết">
  <!-- Likes -->
  <span class="inline-flex items-center gap-1" title="Lượt thích">
    <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2" aria-hidden="true">
      <path stroke-linecap="round" stroke-linejoin="round" d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z" />
    </svg>
    <span aria-label="{likes} lượt thích">{likesDisplay}</span>
  </span>

  <!-- Comments -->
  <span class="inline-flex items-center gap-1" title="Bình luận">
    <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2" aria-hidden="true">
      <path stroke-linecap="round" stroke-linejoin="round" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
    </svg>
    <span aria-label="{comments} bình luận">{commentsDisplay}</span>
  </span>

  <!-- Shares -->
  <span class="inline-flex items-center gap-1" title="Lượt chia sẻ">
    <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2" aria-hidden="true">
      <path stroke-linecap="round" stroke-linejoin="round" d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m0 2.684l6.632 3.316m-6.632-6l6.632-3.316m0 0a3 3 0 105.367-2.684 3 3 0 00-5.367 2.684zm0 9.316a3 3 0 105.368 2.684 3 3 0 00-5.368-2.684z" />
    </svg>
    <span aria-label="{shares} lượt chia sẻ">{sharesDisplay}</span>
  </span>
</div>
