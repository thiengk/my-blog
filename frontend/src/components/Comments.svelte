<script>
  /**
   * Giscus Comment System Wrapper
   * Integrates Giscus widget with theme synchronization.
   *
   * Validates: Requirements 5.1, 5.2, 5.3, 5.4, 5.5
   *
   * Configuration:
   * - Replace data-repo, data-repo-id, data-category, data-category-id
   *   with your actual GitHub repository and Discussions category values.
   * - Get these values from https://giscus.app
   */

  /** @type {HTMLDivElement | undefined} */
  let container = $state(undefined);

  /** @type {boolean} */
  let loaded = $state(false);

  /**
   * Detect current theme based on document's dark class.
   * @returns {'light' | 'dark_dimmed'}
   */
  function getGiscusTheme() {
    if (typeof document === 'undefined') return 'light';
    return document.documentElement.classList.contains('dark') ? 'dark_dimmed' : 'light';
  }

  /**
   * Send a message to the Giscus iframe to update its theme.
   * @param {'light' | 'dark_dimmed'} theme
   */
  function setGiscusTheme(theme) {
    const iframe = container?.querySelector('iframe.giscus-frame');
    if (!iframe) return;

    /** @type {HTMLIFrameElement} */ (iframe).contentWindow?.postMessage(
      { giscus: { setConfig: { theme } } },
      'https://giscus.app'
    );
  }

  /**
   * Load the Giscus script and inject it into the container.
   */
  function loadGiscus() {
    if (!container || loaded) return;

    const script = document.createElement('script');
    script.src = 'https://giscus.app/client.js';
    script.setAttribute('data-repo', 'username/repo-name');
    script.setAttribute('data-repo-id', 'R_placeholder');
    script.setAttribute('data-category', 'Blog Comments');
    script.setAttribute('data-category-id', 'DIC_placeholder');
    script.setAttribute('data-mapping', 'pathname');
    script.setAttribute('data-strict', '0');
    script.setAttribute('data-reactions-enabled', '1');
    script.setAttribute('data-emit-metadata', '0');
    script.setAttribute('data-input-position', 'top');
    script.setAttribute('data-theme', getGiscusTheme());
    script.setAttribute('data-lang', 'vi');
    script.setAttribute('data-loading', 'lazy');
    script.crossOrigin = 'anonymous';
    script.async = true;

    container.appendChild(script);
    loaded = true;
  }

  // Initialize Giscus on mount and watch for theme changes
  $effect(() => {
    if (!container) return;

    // Load Giscus script
    loadGiscus();

    // Observe theme changes on <html> element (dark class toggle)
    const observer = new MutationObserver((mutations) => {
      for (const mutation of mutations) {
        if (mutation.attributeName === 'class') {
          setGiscusTheme(getGiscusTheme());
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
  <div bind:this={container} class="giscus-container min-h-[260px]"></div>
</section>
