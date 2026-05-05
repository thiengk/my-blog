<script>
  import Fuse from 'fuse.js';

  /**
   * @typedef {Object} SearchPost
   * @property {string} slug
   * @property {string} title
   * @property {string} description
   * @property {string[]} tags
   * @property {string} category
   * @property {string} content
   * @property {string} date
   */

  /**
   * @typedef {Object} SearchResult
   * @property {SearchPost} item
   * @property {number} score
   */

  // State
  let isOpen = $state(false);
  let query = $state('');
  let results = $state([]);
  let selectedIndex = $state(-1);
  let isLoading = $state(false);
  let searchIndex = $state(null);
  let fuse = $state(null);
  let inputRef = $state(null);
  let debounceTimer = $state(null);

  // Fuse.js configuration
  const fuseOptions = {
    keys: [
      { name: 'title', weight: 2 },
      { name: 'description', weight: 1.5 },
      { name: 'tags', weight: 1.5 },
      { name: 'content', weight: 1 },
    ],
    threshold: 0.3,
    includeScore: true,
    includeMatches: true,
    minMatchCharLength: 2,
    useExtendedSearch: true,
    ignoreLocation: true,
  };

  // Suggestions when no results found
  const suggestions = [
    'Thử tìm với từ khóa ngắn hơn',
    'Kiểm tra lại chính tả',
    'Thử tìm theo tag hoặc category',
    'Sử dụng từ khóa tiếng Việt không dấu',
  ];

  // Derived state
  let hasResults = $derived(results.length > 0);
  let showNoResults = $derived(query.length >= 2 && !hasResults && !isLoading);

  // Load search index on first open
  async function loadSearchIndex() {
    if (searchIndex) return;
    isLoading = true;
    try {
      const response = await fetch('/search.json');
      const data = await response.json();
      searchIndex = data.posts;
      fuse = new Fuse(searchIndex, fuseOptions);
    } catch (error) {
      console.error('Failed to load search index:', error);
    } finally {
      isLoading = false;
    }
  }

  // Open search modal
  async function openSearch() {
    isOpen = true;
    await loadSearchIndex();
    // Focus input after DOM update
    requestAnimationFrame(() => {
      if (inputRef) inputRef.focus();
    });
  }

  // Close search modal
  function closeSearch() {
    isOpen = false;
    query = '';
    results = [];
    selectedIndex = -1;
  }

  // Debounced search
  function handleInput(event) {
    const value = event.target.value;
    query = value;
    selectedIndex = -1;

    if (debounceTimer) clearTimeout(debounceTimer);

    if (value.length < 2) {
      results = [];
      return;
    }

    debounceTimer = setTimeout(() => {
      performSearch(value);
    }, 150);
  }

  // Perform search using Fuse.js
  function performSearch(searchQuery) {
    if (!fuse || !searchQuery) {
      results = [];
      return;
    }

    const startTime = performance.now();
    const fuseResults = fuse.search(searchQuery, { limit: 10 });
    const endTime = performance.now();

    // Log performance in development
    if (endTime - startTime > 100) {
      console.warn(`Search took ${(endTime - startTime).toFixed(1)}ms`);
    }

    results = fuseResults;
  }

  // Highlight matching text in results
  function highlight(text, searchQuery) {
    if (!searchQuery || searchQuery.length < 2 || !text) return text;

    // Escape special regex characters in the query
    const escaped = searchQuery.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
    const regex = new RegExp(`(${escaped})`, 'gi');
    return text.replace(regex, '<mark class="bg-yellow-200 dark:bg-yellow-800 text-inherit rounded px-0.5">$1</mark>');
  }

  // Format date for display
  function formatDate(dateStr) {
    const date = new Date(dateStr);
    return date.toLocaleDateString('vi-VN', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    });
  }

  // Keyboard navigation
  function handleKeydown(event) {
    if (event.key === 'Escape') {
      closeSearch();
      return;
    }

    if (!hasResults) return;

    if (event.key === 'ArrowDown') {
      event.preventDefault();
      selectedIndex = Math.min(selectedIndex + 1, results.length - 1);
    } else if (event.key === 'ArrowUp') {
      event.preventDefault();
      selectedIndex = Math.max(selectedIndex - 1, -1);
    } else if (event.key === 'Enter' && selectedIndex >= 0) {
      event.preventDefault();
      navigateToResult(results[selectedIndex].item.slug);
    }
  }

  // Navigate to selected result
  function navigateToResult(slug) {
    window.location.href = `/blog/${slug}`;
    closeSearch();
  }

  // Global keyboard shortcut (Ctrl+K or Cmd+K)
  function handleGlobalKeydown(event) {
    if ((event.metaKey || event.ctrlKey) && event.key === 'k') {
      event.preventDefault();
      if (isOpen) {
        closeSearch();
      } else {
        openSearch();
      }
    }
  }
</script>

<svelte:window onkeydown={handleGlobalKeydown} />

<!-- Search Trigger Button -->
<button
  onclick={openSearch}
  class="flex items-center justify-center w-10 h-10 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors text-gray-700 dark:text-gray-300"
  aria-label="Tìm kiếm (Ctrl+K)"
  title="Tìm kiếm (Ctrl+K)"
>
  <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
    <path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
  </svg>
</button>

<!-- Search Modal -->
{#if isOpen}
  <!-- Backdrop -->
  <div
    class="fixed inset-0 z-50 bg-black/50 backdrop-blur-sm"
    onclick={closeSearch}
    onkeydown={(e) => e.key === 'Escape' && closeSearch()}
    role="presentation"
  ></div>

  <!-- Search Dialog -->
  <div
    class="fixed inset-x-0 top-[10%] z-50 mx-auto w-full max-w-xl px-4"
    role="search"
    aria-label="Tìm kiếm bài viết"
  >
    <div class="rounded-xl bg-white dark:bg-gray-800 shadow-2xl border border-gray-200 dark:border-gray-700 overflow-hidden">
      <!-- Search Input -->
      <div class="flex items-center gap-3 px-4 border-b border-gray-200 dark:border-gray-700">
        <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5 text-gray-400 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
        </svg>
        <input
          bind:this={inputRef}
          type="text"
          value={query}
          oninput={handleInput}
          onkeydown={handleKeydown}
          placeholder="Tìm kiếm bài viết..."
          class="flex-1 py-4 bg-transparent text-gray-900 dark:text-gray-100 placeholder-gray-400 dark:placeholder-gray-500 outline-none text-base"
          aria-label="Nhập từ khóa tìm kiếm"
          aria-controls="search-results"
          aria-expanded={hasResults}
          autocomplete="off"
          spellcheck="false"
        />
        {#if isLoading}
          <div class="w-5 h-5 border-2 border-gray-300 dark:border-gray-600 border-t-primary-500 rounded-full animate-spin"></div>
        {:else if query}
          <button
            onclick={() => { query = ''; results = []; if (inputRef) inputRef.focus(); }}
            class="p-1 rounded hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-400"
            aria-label="Xóa tìm kiếm"
          >
            <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        {/if}
      </div>

      <!-- Results -->
      <div
        id="search-results"
        class="max-h-[60vh] overflow-y-auto"
        role="listbox"
        aria-label="Kết quả tìm kiếm"
      >
        {#if hasResults}
          <ul class="py-2">
            {#each results as result, index}
              <li
                role="option"
                aria-selected={index === selectedIndex}
              >
                <a
                  href="/blog/{result.item.slug}"
                  class="block px-4 py-3 hover:bg-gray-50 dark:hover:bg-gray-700/50 transition-colors {index === selectedIndex ? 'bg-gray-50 dark:bg-gray-700/50' : ''}"
                  onclick={closeSearch}
                >
                  <div class="flex items-start justify-between gap-2">
                    <div class="min-w-0 flex-1">
                      <!-- Title with highlight -->
                      <h3 class="text-sm font-semibold text-gray-900 dark:text-gray-100 truncate">
                        {@html highlight(result.item.title, query)}
                      </h3>
                      <!-- Description with highlight -->
                      <p class="mt-1 text-xs text-gray-600 dark:text-gray-400 line-clamp-2">
                        {@html highlight(result.item.description, query)}
                      </p>
                      <!-- Meta info -->
                      <div class="mt-1.5 flex items-center gap-2 text-xs text-gray-500 dark:text-gray-500">
                        <span class="inline-flex items-center px-1.5 py-0.5 rounded bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400">
                          {result.item.category}
                        </span>
                        <span>{formatDate(result.item.date)}</span>
                      </div>
                    </div>
                    <!-- Arrow icon -->
                    <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 text-gray-400 shrink-0 mt-1" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
                    </svg>
                  </div>
                </a>
              </li>
            {/each}
          </ul>
        {:else if showNoResults}
          <!-- No results message -->
          <div class="px-4 py-8 text-center">
            <svg xmlns="http://www.w3.org/2000/svg" class="w-12 h-12 mx-auto text-gray-300 dark:text-gray-600 mb-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <p class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Không tìm thấy kết quả cho "{query}"
            </p>
            <p class="text-xs text-gray-500 dark:text-gray-400 mb-3">Gợi ý:</p>
            <ul class="text-xs text-gray-500 dark:text-gray-400 space-y-1">
              {#each suggestions as suggestion}
                <li>• {suggestion}</li>
              {/each}
            </ul>
          </div>
        {:else if query.length > 0 && query.length < 2}
          <div class="px-4 py-6 text-center text-sm text-gray-500 dark:text-gray-400">
            Nhập ít nhất 2 ký tự để tìm kiếm
          </div>
        {/if}
      </div>

      <!-- Footer -->
      <div class="flex items-center justify-between px-4 py-2 border-t border-gray-200 dark:border-gray-700 text-xs text-gray-500 dark:text-gray-400">
        <div class="flex items-center gap-2">
          <kbd class="px-1.5 py-0.5 rounded border border-gray-300 dark:border-gray-600 bg-gray-50 dark:bg-gray-700 font-mono text-[10px]">↑↓</kbd>
          <span>di chuyển</span>
          <kbd class="px-1.5 py-0.5 rounded border border-gray-300 dark:border-gray-600 bg-gray-50 dark:bg-gray-700 font-mono text-[10px]">↵</kbd>
          <span>chọn</span>
        </div>
        <div class="flex items-center gap-1">
          <kbd class="px-1.5 py-0.5 rounded border border-gray-300 dark:border-gray-600 bg-gray-50 dark:bg-gray-700 font-mono text-[10px]">Esc</kbd>
          <span>đóng</span>
        </div>
      </div>
    </div>
  </div>
{/if}
