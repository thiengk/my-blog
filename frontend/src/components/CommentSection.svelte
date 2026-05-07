<script>
  /**
   * Comment Section component.
   * Displays existing comments and provides a form to submit new comments.
   * Replaces the Utterances-based Comments.svelte with an internal comment system.
   *
   * @typedef {'idle' | 'loading' | 'success' | 'error'} FormStatus
   * @typedef {'idle' | 'loading' | 'error'} FetchStatus
   * @typedef {{ id: number, slug: string, author_name: string, content: string, created_at: string }} Comment
   */

  /** @type {{ slug: string }} */
  let { slug } = $props();

  const API_URL = import.meta.env.PUBLIC_API_URL || '';

  // Form state
  let authorName = $state('');
  let content = $state('');
  /** @type {FormStatus} */
  let formStatus = $state('idle');
  let formError = $state('');
  let authorError = $state('');
  let contentError = $state('');

  // Comments list state
  /** @type {Comment[]} */
  let comments = $state([]);
  /** @type {FetchStatus} */
  let fetchStatus = $state('idle');
  let fetchError = $state('');

  // Derived state
  let isSubmitDisabled = $derived(formStatus === 'loading');
  let hasAuthorError = $derived(authorError !== '');
  let hasContentError = $derived(contentError !== '');
  let hasFormError = $derived(formError !== '');
  let commentCount = $derived(comments.length);

  /**
   * Validate author name field.
   * @param {string} value
   * @returns {boolean}
   */
  function validateAuthor(value) {
    const trimmed = value.trim();
    if (trimmed.length === 0) {
      authorError = 'Vui lòng nhập tên của bạn';
      return false;
    }
    if (trimmed.length > 100) {
      authorError = 'Tên không được vượt quá 100 ký tự';
      return false;
    }
    authorError = '';
    return true;
  }

  /**
   * Validate content field.
   * @param {string} value
   * @returns {boolean}
   */
  function validateContent(value) {
    const trimmed = value.trim();
    if (trimmed.length === 0) {
      contentError = 'Vui lòng nhập nội dung bình luận';
      return false;
    }
    if (trimmed.length > 5000) {
      contentError = 'Nội dung không được vượt quá 5000 ký tự';
      return false;
    }
    contentError = '';
    return true;
  }

  /**
   * Fetch comments for the current post slug.
   */
  async function fetchComments() {
    fetchStatus = 'loading';
    fetchError = '';

    try {
      const response = await fetch(`${API_URL}/api/comments/${slug}`);

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`);
      }

      const data = await response.json();
      comments = data.comments || [];
      fetchStatus = 'idle';
    } catch (err) {
      fetchStatus = 'error';
      fetchError = 'Không thể tải bình luận. Vui lòng thử lại sau.';
    }
  }

  /**
   * Handle comment form submission.
   * @param {SubmitEvent} event
   */
  async function handleSubmit(event) {
    event.preventDefault();

    // Client-side validation
    const isAuthorValid = validateAuthor(authorName);
    const isContentValid = validateContent(content);

    if (!isAuthorValid || !isContentValid) {
      return;
    }

    formStatus = 'loading';
    formError = '';

    try {
      const response = await fetch(`${API_URL}/api/comments/${slug}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          author_name: authorName.trim(),
          content: content.trim(),
        }),
      });

      const data = await response.json();

      if (response.ok || response.status === 201) {
        formStatus = 'success';
        // Add the new comment to the list without reload
        const newComment = data.comment || data;
        comments = [...comments, newComment];
        // Reset form fields
        authorName = '';
        content = '';
        // Reset form status after a short delay
        setTimeout(() => {
          formStatus = 'idle';
        }, 3000);
      } else if (response.status === 422) {
        formStatus = 'error';
        formError = data.message || 'Dữ liệu không hợp lệ. Vui lòng kiểm tra lại.';
      } else if (response.status === 429) {
        formStatus = 'error';
        formError = 'Bạn đã gửi quá nhiều bình luận. Vui lòng thử lại sau.';
      } else {
        formStatus = 'error';
        formError = data.message || 'Đã có lỗi xảy ra. Vui lòng thử lại sau.';
      }
    } catch (err) {
      formStatus = 'error';
      formError = 'Không thể kết nối đến server. Vui lòng thử lại sau.';
    }
  }

  /**
   * Clear author validation error on input.
   */
  function handleAuthorInput() {
    if (authorError) {
      authorError = '';
    }
    if (formStatus === 'error') {
      formStatus = 'idle';
      formError = '';
    }
  }

  /**
   * Clear content validation error on input.
   */
  function handleContentInput() {
    if (contentError) {
      contentError = '';
    }
    if (formStatus === 'error') {
      formStatus = 'idle';
      formError = '';
    }
  }

  /**
   * Format a date string for display.
   * @param {string} dateStr
   * @returns {string}
   */
  function formatDate(dateStr) {
    try {
      const date = new Date(dateStr);
      return date.toLocaleDateString('vi-VN', {
        year: 'numeric',
        month: 'long',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
      });
    } catch {
      return dateStr;
    }
  }

  // Fetch comments on mount
  $effect(() => {
    if (slug) {
      fetchComments();
    }
  });
</script>

<section class="mt-12" aria-label="Bình luận bài viết">
  <h2 class="text-xl font-bold text-gray-900 dark:text-gray-100 mb-6">
    Bình luận {#if commentCount > 0}<span class="text-base font-normal text-gray-500 dark:text-gray-400">({commentCount})</span>{/if}
  </h2>

  <!-- Comment Form -->
  <form
    onsubmit={handleSubmit}
    class="mb-8 p-4 rounded-lg border border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/50"
    novalidate
    aria-label="Viết bình luận"
  >
    <div class="mb-4">
      <label for="comment-author" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
        Tên của bạn
      </label>
      <input
        id="comment-author"
        type="text"
        bind:value={authorName}
        oninput={handleAuthorInput}
        placeholder="Nhập tên của bạn"
        class="w-full px-3 py-2 rounded-md border {hasAuthorError ? 'border-red-500 dark:border-red-400' : 'border-gray-300 dark:border-gray-600'} bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-colors"
        aria-describedby={hasAuthorError ? 'author-error' : undefined}
        aria-invalid={hasAuthorError ? 'true' : undefined}
        disabled={formStatus === 'loading'}
        maxlength="100"
      />
      {#if hasAuthorError}
        <p id="author-error" class="mt-1 text-sm text-red-600 dark:text-red-400" role="alert">
          {authorError}
        </p>
      {/if}
    </div>

    <div class="mb-4">
      <label for="comment-content" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
        Nội dung bình luận
      </label>
      <textarea
        id="comment-content"
        bind:value={content}
        oninput={handleContentInput}
        placeholder="Viết bình luận của bạn..."
        rows="4"
        class="w-full px-3 py-2 rounded-md border {hasContentError ? 'border-red-500 dark:border-red-400' : 'border-gray-300 dark:border-gray-600'} bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-colors resize-y"
        aria-describedby={hasContentError ? 'content-error' : undefined}
        aria-invalid={hasContentError ? 'true' : undefined}
        disabled={formStatus === 'loading'}
        maxlength="5000"
      ></textarea>
      {#if hasContentError}
        <p id="content-error" class="mt-1 text-sm text-red-600 dark:text-red-400" role="alert">
          {contentError}
        </p>
      {/if}
    </div>

    <!-- Form Error -->
    {#if hasFormError}
      <p class="mb-3 text-sm text-red-600 dark:text-red-400" role="alert" aria-live="assertive">
        {formError}
      </p>
    {/if}

    <!-- Success Message -->
    {#if formStatus === 'success'}
      <p class="mb-3 text-sm text-green-600 dark:text-green-400" role="status" aria-live="polite">
        Bình luận đã được gửi thành công!
      </p>
    {/if}

    <button
      type="submit"
      class="px-4 py-2 rounded-md bg-blue-600 hover:bg-blue-700 text-white font-medium text-sm transition-colors disabled:opacity-60 disabled:cursor-not-allowed"
      disabled={isSubmitDisabled}
      aria-label={formStatus === 'loading' ? 'Đang gửi bình luận...' : 'Gửi bình luận'}
    >
      {#if formStatus === 'loading'}
        <svg
          class="animate-spin inline-block -ml-1 mr-2 h-4 w-4 text-white"
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
          aria-hidden="true"
        >
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        Đang gửi...
      {:else}
        Gửi bình luận
      {/if}
    </button>
  </form>

  <!-- Comments List -->
  <div class="space-y-6">
    {#if fetchStatus === 'loading'}
      <!-- Loading State -->
      <div class="flex items-center justify-center py-8" role="status" aria-live="polite">
        <svg
          class="animate-spin h-6 w-6 text-gray-400 dark:text-gray-500"
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
          aria-hidden="true"
        >
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        <span class="ml-2 text-sm text-gray-500 dark:text-gray-400">Đang tải bình luận...</span>
      </div>
    {:else if fetchStatus === 'error'}
      <!-- Fetch Error State -->
      <div class="text-center py-8">
        <p class="text-sm text-red-600 dark:text-red-400" role="alert">{fetchError}</p>
        <button
          onclick={fetchComments}
          class="mt-2 text-sm text-blue-600 dark:text-blue-400 underline underline-offset-2 hover:text-blue-700 dark:hover:text-blue-300 transition-colors"
        >
          Thử lại
        </button>
      </div>
    {:else if comments.length === 0}
      <!-- Empty State -->
      <p class="text-center py-8 text-sm text-gray-500 dark:text-gray-400">
        Chưa có bình luận nào. Hãy là người đầu tiên bình luận!
      </p>
    {:else}
      <!-- Comments -->
      {#each comments as comment (comment.id)}
        <article class="p-4 rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800/30">
          <div class="flex items-center gap-2 mb-2">
            <span class="font-medium text-sm text-gray-900 dark:text-gray-100">
              {comment.author_name}
            </span>
            <span class="text-xs text-gray-500 dark:text-gray-400">
              {formatDate(comment.created_at)}
            </span>
          </div>
          <p class="text-sm text-gray-700 dark:text-gray-300 whitespace-pre-wrap break-words">
            {comment.content}
          </p>
        </article>
      {/each}
    {/if}
  </div>
</section>
