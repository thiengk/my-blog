<script>
  /**
   * Utterances Comment System Wrapper
   * Integrates Utterances widget with theme synchronization.
   *
   * Validates: Requirements 5.1, 5.2, 5.3, 5.4, 5.5
   *
   * Configuration:
   * - Replace 'username/repo-name' with your actual GitHub repository
   * - Repository must be public
   * - Install Utterances app: https://github.com/apps/utterances
   *
   * Utterances is simpler than Giscus:
   * - Uses GitHub Issues instead of Discussions
   * - No need for repo-id or category-id
   * - Lightweight and easy to configure
   */

  /** @type {HTMLDivElement | undefined} */
  let container = $state(undefined);

  /** @type {boolean} */
  let loaded = $state(false);

  /**
   * Detect current theme based on document's dark class.
   * @returns {'github-light' | 'github-dark'}
   */
  function getUtterancesTheme() {
    if (typeof document === 'undefined') return 'github-light';
    return document.documentElement.classList.contains('dark') ? 'github-dark' : 'github-light';
  }

  /**
   * Send a message to the Utterances iframe to update its theme.
   * @param {'github-light' | 'github-dark'} theme
   */
  function setUtterancesTheme(theme) {
    const iframe = container?.querySelector('iframe.utterances-frame');
    if (!iframe) return;

    /** @type {HTMLIFrameElement} */ (iframe).contentWindow?.postMessage(
      { type: 'set-theme', theme },
      'https://utteranc.es'
    );
  }

  /**
   * Load the Utterances script and inject it into the container.
   */
  function loadUtterances() {
    if (!container || loaded) return;

    const script = document.createElement('script');
    script.src = 'https://utteranc.es/client.js';
    // TODO: Thay 'username/repo-name' bằng repository GitHub của bạn
    script.setAttribute('repo', 'username/repo-name');
    script.setAttribute('issue-term', 'pathname');
    script.setAttribute('label', 'blog-comment');
    script.setAttribute('theme', getUtterancesTheme());
    script.crossOrigin = 'anonymous';
    script.async = true;

    container.appendChild(script);
    loaded = true;
  }

  // Initialize Utterances on mount and watch for theme changes
  $effect(() => {
    if (!container) return;

    // Load Utterances script
    loadUtterances();

    // Observe theme changes on <html> element (dark class toggle)
    const observer = new MutationObserver((mutations) => {
      for (const mutation of mutations) {
        if (mutation.attributeName === 'class') {
          setUtterancesTheme(getUtterancesTheme());
        }
      }
    });

    observer.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ['class'],
    });

    return () => {
      observer.disconnect();
    };
  });
</script>

<section
  class="mt-12"
  aria-label="Bình luận bài viết"
>
  <h2 class="text-xl font-bold text-gray-900 dark:text-gray-100 mb-6">
    Bình luận
  </h2>
  <div bind:this={container} class="utterances-container min-h-[260px]"></div>
</section>
