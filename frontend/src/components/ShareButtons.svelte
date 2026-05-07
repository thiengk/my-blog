<script>
  /**
   * ShareButtons Component
   * Displays share count and platform share buttons (Facebook, Twitter, LinkedIn, Copy Link).
   * Records share actions via the engagement API.
   *
   * Validates: Requirements 3.1, 3.4
   */

  const API_URL = import.meta.env.PUBLIC_API_URL || '';

  /** @type {{ slug: string, shareCount: number, url: string }} */
  let { slug, shareCount = 0, url } = $props();

  /** @type {number} */
  let currentShareCount = $state(shareCount);
  /** @type {string} */
  let copyStatus = $state('idle');

  /**
   * Record a share action to the backend API.
   * @param {string} platform - The platform being shared to
   */
  async function recordShare(platform) {
    try {
      await fetch(`${API_URL}/api/engagement/share/${encodeURIComponent(slug)}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ platform }),
      });
      currentShareCount += 1;
    } catch {
      // Graceful degradation: still open share dialog even if tracking fails
    }
  }

  /**
   * Share on Facebook.
   */
  function shareOnFacebook() {
    const shareUrl = `https://www.facebook.com/sharer/sharer.php?u=${encodeURIComponent(url)}`;
    window.open(shareUrl, '_blank', 'width=600,height=400');
    recordShare('facebook');
  }

  /**
   * Share on Twitter.
   */
  function shareOnTwitter() {
    const shareUrl = `https://twitter.com/intent/tweet?url=${encodeURIComponent(url)}`;
    window.open(shareUrl, '_blank', 'width=600,height=400');
    recordShare('twitter');
  }

  /**
   * Share on LinkedIn.
   */
  function shareOnLinkedIn() {
    const shareUrl = `https://www.linkedin.com/sharing/share-offsite/?url=${encodeURIComponent(url)}`;
    window.open(shareUrl, '_blank', 'width=600,height=400');
    recordShare('linkedin');
  }

  /**
   * Copy link to clipboard.
   */
  async function copyLink() {
    try {
      await navigator.clipboard.writeText(url);
      copyStatus = 'copied';
      recordShare('copy-link');
      setTimeout(() => {
        copyStatus = 'idle';
      }, 2000);
    } catch {
      copyStatus = 'error';
      setTimeout(() => {
        copyStatus = 'idle';
      }, 2000);
    }
  }

  let displayCount = $derived(currentShareCount.toLocaleString('vi-VN'));
</script>

<div class="share-buttons flex items-center gap-2 flex-wrap">
  <span class="text-sm text-gray-600 dark:text-gray-400">
    {displayCount} lượt chia sẻ
  </span>

  <div class="flex items-center gap-1">
    <!-- Facebook -->
    <button
      onclick={shareOnFacebook}
      class="share-btn p-2 rounded-lg hover:bg-blue-50 dark:hover:bg-blue-900/20 text-blue-600 dark:text-blue-400 transition-colors"
      aria-label="Chia sẻ lên Facebook"
      title="Chia sẻ lên Facebook"
    >
      <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24" aria-hidden="true">
        <path d="M24 12.073c0-6.627-5.373-12-12-12s-12 5.373-12 12c0 5.99 4.388 10.954 10.125 11.854v-8.385H7.078v-3.47h3.047V9.43c0-3.007 1.792-4.669 4.533-4.669 1.312 0 2.686.235 2.686.235v2.953H15.83c-1.491 0-1.956.925-1.956 1.874v2.25h3.328l-.532 3.47h-2.796v8.385C19.612 23.027 24 18.062 24 12.073z"/>
      </svg>
    </button>

    <!-- Twitter -->
    <button
      onclick={shareOnTwitter}
      class="share-btn p-2 rounded-lg hover:bg-sky-50 dark:hover:bg-sky-900/20 text-sky-500 dark:text-sky-400 transition-colors"
      aria-label="Chia sẻ lên Twitter"
      title="Chia sẻ lên Twitter"
    >
      <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24" aria-hidden="true">
        <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z"/>
      </svg>
    </button>

    <!-- LinkedIn -->
    <button
      onclick={shareOnLinkedIn}
      class="share-btn p-2 rounded-lg hover:bg-blue-50 dark:hover:bg-blue-900/20 text-blue-700 dark:text-blue-300 transition-colors"
      aria-label="Chia sẻ lên LinkedIn"
      title="Chia sẻ lên LinkedIn"
    >
      <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24" aria-hidden="true">
        <path d="M20.447 20.452h-3.554v-5.569c0-1.328-.027-3.037-1.852-3.037-1.853 0-2.136 1.445-2.136 2.939v5.667H9.351V9h3.414v1.561h.046c.477-.9 1.637-1.85 3.37-1.85 3.601 0 4.267 2.37 4.267 5.455v6.286zM5.337 7.433c-1.144 0-2.063-.926-2.063-2.065 0-1.138.92-2.063 2.063-2.063 1.14 0 2.064.925 2.064 2.063 0 1.139-.925 2.065-2.064 2.065zm1.782 13.019H3.555V9h3.564v11.452zM22.225 0H1.771C.792 0 0 .774 0 1.729v20.542C0 23.227.792 24 1.771 24h20.451C23.2 24 24 23.227 24 22.271V1.729C24 .774 23.2 0 22.222 0h.003z"/>
      </svg>
    </button>

    <!-- Copy Link -->
    <button
      onclick={copyLink}
      class="share-btn p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-600 dark:text-gray-400 transition-colors"
      aria-label={copyStatus === 'copied' ? 'Đã sao chép liên kết' : 'Sao chép liên kết'}
      title={copyStatus === 'copied' ? 'Đã sao chép!' : 'Sao chép liên kết'}
    >
      {#if copyStatus === 'copied'}
        <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5 text-green-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2" aria-hidden="true">
          <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
        </svg>
      {:else}
        <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2" aria-hidden="true">
          <path stroke-linecap="round" stroke-linejoin="round" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
        </svg>
      {/if}
    </button>
  </div>
</div>
