<script>
  /**
   * ViewCounter Component
   * Displays view count for a blog post and records a new view on mount.
   * Handles graceful degradation when the API is unavailable.
   *
   * Validates: Requirements 7.1, 7.3
   */
  import { recordView, getViewCount } from '../lib/api';

  /** @type {{ slug: string }} */
  let { slug } = $props();

  /** @type {number | null} */
  let viewCount = $state(null);
  let isLoading = $state(true);

  $effect(() => {
    if (!slug) return;

    let cancelled = false;

    async function init() {
      // Record the view (fire-and-forget, errors handled internally)
      recordView(slug);

      // Fetch current view count
      try {
        const count = await getViewCount(slug);
        if (!cancelled) {
          viewCount = count;
          isLoading = false;
        }
      } catch {
        // Graceful degradation: show "--" if API unavailable
        if (!cancelled) {
          viewCount = null;
          isLoading = false;
        }
      }
    }

    init();

    return () => {
      cancelled = true;
    };
  });

  let displayText = $derived(
    isLoading ? '-- lượt xem' : viewCount !== null ? `${viewCount.toLocaleString('vi-VN')} lượt xem` : '-- lượt xem'
  );
</script>

<span class="view-counter">{displayText}</span>
