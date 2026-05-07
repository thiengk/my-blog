<script>
  /**
   * LikeButton Component
   * Displays like count and handles like action for a blog post.
   * Uses localStorage to persist liked state and optimistic UI updates.
   *
   * Validates: Requirements 2.1, 4.4
   */

  // Configurable API URL - defaults to relative path for same-origin
  const API_URL = import.meta.env.PUBLIC_API_URL || '';

  /** @type {{ slug: string, likeCount?: number }} */
  let { slug, likeCount = 0 } = $props();

  /** @type {'idle' | 'loading' | 'success' | 'error'} */
  let status = $state('idle');
  let count = $state(likeCount);
  let hasLiked = $state(false);
  let errorMessage = $state('');

  /**
   * Get the localStorage key for tracking liked posts.
   * @param {string} postSlug
   * @returns {string}
   */
  function getLikedKey(postSlug) {
    return `blog:liked:${postSlug}`;
  }

  /**
   * Check if the user has already liked this post (from localStorage).
   */
  function checkIfLiked() {
    try {
      const liked = localStorage.getItem(getLikedKey(slug));
      if (liked === 'true') {
        hasLiked = true;
      }
    } catch {
      // localStorage unavailable (e.g., private browsing) - allow liking
    }
  }

  /**
   * Save liked state to localStorage.
   */
  function saveLikedState() {
    try {
      localStorage.setItem(getLikedKey(slug), 'true');
    } catch {
      // localStorage unavailable - still allow the like to go through
    }
  }

  // Check liked state on mount
  $effect(() => {
    if (!slug) return;
    checkIfLiked();
  });

  /**
   * Handle like button click.
   * Sends POST request and performs optimistic UI update.
   */
  async function handleLike() {
    if (hasLiked || status === 'loading') return;

    status = 'loading';
    errorMessage = '';

    // Optimistic update
    count += 1;
    hasLiked = true;

    try {
      const response = await fetch(`${API_URL}/api/engagement/like/${slug}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        const data = await response.json();
        status = 'success';
        // Persist liked state
        saveLikedState();

        // If server says it wasn't counted (duplicate), keep the UI as-is
        // since localStorage already prevents re-clicking
        if (!data.counted) {
          // Revert optimistic update if not actually counted
          count -= 1;
        }
      } else if (response.status === 429) {
        // Rate limited - revert optimistic update
        count -= 1;
        hasLiked = false;
        status = 'error';
        errorMessage = 'Bạn đã thao tác quá nhanh. Vui lòng thử lại sau.';
      } else {
        // Other error - revert optimistic update
        count -= 1;
        hasLiked = false;
        status = 'error';
        errorMessage = 'Đã có lỗi xảy ra. Vui lòng thử lại.';
      }
    } catch {
      // Network error - revert optimistic update
      count -= 1;
      hasLiked = false;
      status = 'error';
      errorMessage = 'Không thể kết nối đến server. Vui lòng thử lại sau.';
    }
  }

  let isDisabled = $derived(hasLiked || status === 'loading');
  let buttonLabel = $derived(
    hasLiked ? 'Đã thích bài viết này' : 'Thích bài viết này'
  );
</script>

<div class="like-button-wrapper">
  <button
    type="button"
    onclick={handleLike}
    disabled={isDisabled}
    class="like-button {hasLiked ? 'liked' : ''} {status === 'loading' ? 'loading' : ''}"
    aria-label={buttonLabel}
    title={buttonLabel}
  >
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 24 24"
      fill={hasLiked ? 'currentColor' : 'none'}
      stroke="currentColor"
      stroke-width="2"
      class="like-icon"
      aria-hidden="true"
    >
      <path
        stroke-linecap="round"
        stroke-linejoin="round"
        d="M21 8.25c0-2.485-2.099-4.5-4.688-4.5-1.935 0-3.597 1.126-4.312 2.733-.715-1.607-2.377-2.733-4.313-2.733C5.1 3.75 3 5.765 3 8.25c0 7.22 9 12 9 12s9-4.78 9-12z"
      />
    </svg>
    <span class="like-count">{count.toLocaleString('vi-VN')}</span>
  </button>

  {#if status === 'error' && errorMessage}
    <p class="like-error" role="alert" aria-live="assertive">
      {errorMessage}
    </p>
  {/if}
</div>

<style>
  .like-button-wrapper {
    display: inline-flex;
    flex-direction: column;
    align-items: center;
    gap: 0.25rem;
  }

  .like-button {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.5rem 1rem;
    border: 1px solid var(--border-color, #e2e8f0);
    border-radius: 9999px;
    background: transparent;
    color: var(--text-color, #4a5568);
    font-size: 0.875rem;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .like-button:hover:not(:disabled) {
    border-color: #ef4444;
    color: #ef4444;
    background: rgba(239, 68, 68, 0.05);
  }

  .like-button:disabled {
    cursor: default;
    opacity: 0.8;
  }

  .like-button.liked {
    border-color: #ef4444;
    color: #ef4444;
    background: rgba(239, 68, 68, 0.05);
  }

  .like-button.loading {
    opacity: 0.7;
  }

  .like-icon {
    width: 1.25rem;
    height: 1.25rem;
    flex-shrink: 0;
  }

  .like-count {
    line-height: 1;
  }

  .like-error {
    margin-top: 0.25rem;
    font-size: 0.75rem;
    color: #ef4444;
  }

  :global(.dark) .like-button {
    border-color: var(--border-color, #4a5568);
    color: var(--text-color, #e2e8f0);
  }

  :global(.dark) .like-button:hover:not(:disabled) {
    border-color: #f87171;
    color: #f87171;
    background: rgba(248, 113, 113, 0.1);
  }

  :global(.dark) .like-button.liked {
    border-color: #f87171;
    color: #f87171;
    background: rgba(248, 113, 113, 0.1);
  }

  :global(.dark) .like-error {
    color: #f87171;
  }
</style>
