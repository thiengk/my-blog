<script>
  /**
   * RecommendedPosts Component
   * Fetches and displays top recommended posts sorted by engagement score.
   * Used on homepage or sidebar to highlight popular content.
   *
   * Validates: Requirements 5.3, 5.7
   */

  const API_URL = import.meta.env.PUBLIC_API_URL || '';

  /** @type {{ limit?: number }} */
  let { limit = 10 } = $props();

  /**
   * @typedef {Object} RankedPost
   * @property {string} slug
   * @property {number} engagement_score
   * @property {number} likes
   * @property {number} comments
   * @property {number} shares
   */

  /** @type {RankedPost[]} */
  let posts = $state([]);
  let isLoading = $state(true);
  let error = $state('');

  $effect(() => {
    let cancelled = false;

    async function fetchRecommendations() {
      isLoading = true;
      error = '';

      try {
        const response = await fetch(`${API_URL}/api/recommendations?limit=${limit}`);

        if (!response.ok) {
          throw new Error(`HTTP ${response.status}`);
        }

        const data = await response.json();

        if (!cancelled) {
          posts = data.posts || [];
          isLoading = false;
        }
      } catch (err) {
        if (!cancelled) {
          posts = [];
          error = 'Không thể tải danh sách bài viết đề xuất.';
          isLoading = false;
        }
      }
    }

    fetchRecommendations();

    return () => {
      cancelled = true;
    };
  });
</script>

<aside class="recommended-posts" aria-label="Bài viết đề xuất">
  <h3 class="text-lg font-semibold mb-3">Bài viết nổi bật</h3>

  {#if isLoading}
    <p class="text-sm text-gray-500 dark:text-gray-400">Đang tải...</p>
  {:else if error}
    <p class="text-sm text-red-600 dark:text-red-400">{error}</p>
  {:else if posts.length === 0}
    <p class="text-sm text-gray-500 dark:text-gray-400">Chưa có bài viết đề xuất.</p>
  {:else}
    <ul class="space-y-2">
      {#each posts as post (post.slug)}
        <li class="flex items-center justify-between gap-2 text-sm">
          <a
            href={`/blog/${post.slug}`}
            class="text-blue-600 dark:text-blue-400 hover:underline truncate"
          >
            {post.slug}
          </a>
          <span
            class="shrink-0 text-xs text-gray-500 dark:text-gray-400"
            title="Điểm tương tác"
          >
            ⭐ {post.engagement_score.toFixed(0)}
          </span>
        </li>
      {/each}
    </ul>
  {/if}
</aside>
