<script>
  /**
   * PasswordGate Component
   * Gates blog post content behind a password form.
   * Uses sessionStorage to persist unlock state within the browser session.
   * Content is only rendered after successful password verification (conditional rendering).
   *
   * Validates: Requirements 2.1, 2.2, 2.3, 2.4, 2.6, 2.7, 7.1, 7.2
   */
  import { isUnlockedInSession, saveUnlockState } from '../lib/password-utils';

  /**
   * @type {{
   *   password: string,
   *   slug: string,
   *   title: string,
   *   date: string,
   *   category?: string,
   *   tags?: string[],
   *   children?: import('svelte').Snippet
   * }}
   */
  let { password, slug, title, date, category, tags, children } = $props();

  let unlocked = $state(false);
  let error = $state('');
  let inputValue = $state('');
  let checking = $state(true);

  /** @type {HTMLInputElement | undefined} */
  let inputRef = $state(undefined);

  let isSubmitDisabled = $derived(inputValue.trim().length === 0);

  // Check session storage on mount
  $effect(() => {
    if (isUnlockedInSession(slug)) {
      unlocked = true;
    }
    checking = false;
  });

  // Auto-focus input when form becomes visible (after checking resolves)
  $effect(() => {
    if (!checking && !unlocked && inputRef) {
      inputRef.focus();
    }
  });

  /**
   * Handle form submission — compare input with password prop using strict equality.
   * @param {Event} e
   */
  function handleSubmit(e) {
    e.preventDefault();

    if (inputValue === password) {
      saveUnlockState(slug);
      unlocked = true;
      error = '';
    } else {
      error = 'Mật khẩu không đúng. Vui lòng thử lại.';
      inputValue = '';
      // Keep focus on input after incorrect attempt
      if (inputRef) {
        inputRef.focus();
      }
    }
  }

  /**
   * Handle keydown events on the input field.
   * Escape key clears the input.
   * @param {KeyboardEvent} e
   */
  function handleKeydown(e) {
    if (e.key === 'Escape') {
      inputValue = '';
    }
  }
</script>

{#if checking}
  <!-- Render nothing while checking session storage to prevent flash -->
{:else if unlocked}
  {#if children}
    {@render children()}
  {/if}
{:else}
  <div class="password-gate">
    <!-- Lock Icon -->
    <div class="password-gate-icon">
      <svg
        xmlns="http://www.w3.org/2000/svg"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
        class="lock-icon"
        aria-hidden="true"
      >
        <rect x="3" y="11" width="18" height="11" rx="2" ry="2" />
        <path d="M7 11V7a5 5 0 0 1 10 0v4" />
      </svg>
    </div>

    <!-- Post Title -->
    <h1 class="password-gate-title">{title}</h1>

    <!-- Metadata -->
    <div class="password-gate-meta">
      <time class="password-gate-date">{date}</time>
      {#if category}
        <span class="password-gate-separator">•</span>
        <span class="password-gate-category">{category}</span>
      {/if}
      {#if tags && tags.length > 0}
        <span class="password-gate-separator">•</span>
        <span class="password-gate-tags">
          {#each tags as tag}
            <span class="password-gate-tag">#{tag}</span>
          {/each}
        </span>
      {/if}
    </div>

    <!-- Protected Notice -->
    <p class="password-gate-notice">Bài viết này được bảo vệ bằng mật khẩu.</p>

    <!-- Password Form -->
    <form onsubmit={handleSubmit} class="password-gate-form">
      <div class="password-gate-input-group">
        <input
          bind:this={inputRef}
          bind:value={inputValue}
          type="password"
          maxlength="128"
          aria-label="Nhập mật khẩu để mở khóa bài viết này"
          placeholder="Nhập mật khẩu..."
          class="password-gate-input"
          onkeydown={handleKeydown}
        />
        <button
          type="submit"
          disabled={isSubmitDisabled}
          aria-label="Mở khóa bài viết"
          class="password-gate-submit"
        >
          Mở khóa
        </button>
      </div>

      <!-- Error Message — container always in DOM for screen reader announcements -->
      <div role="alert" aria-live="assertive" class="password-gate-error">
        {#if error}
          {error}
        {/if}
      </div>
    </form>
  </div>
{/if}

<style>
  .password-gate {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 3rem 1.5rem;
    text-align: center;
    max-width: 480px;
    margin: 2rem auto;
  }

  .password-gate-icon {
    margin-bottom: 1.5rem;
  }

  .lock-icon {
    width: 3rem;
    height: 3rem;
    color: #4b5563;
  }

  :global(.dark) .lock-icon {
    color: #d1d5db;
  }

  .password-gate-title {
    font-size: 1.5rem;
    font-weight: 700;
    color: #111827;
    margin: 0 0 0.75rem 0;
    line-height: 1.3;
  }

  :global(.dark) .password-gate-title {
    color: #f3f4f6;
  }

  .password-gate-meta {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
    font-size: 0.875rem;
    color: #4b5563;
    margin-bottom: 1.5rem;
  }

  :global(.dark) .password-gate-meta {
    color: #d1d5db;
  }

  .password-gate-separator {
    color: #d1d5db;
  }

  :global(.dark) .password-gate-separator {
    color: #4b5563;
  }

  .password-gate-tag {
    margin-right: 0.25rem;
  }

  .password-gate-notice {
    font-size: 0.875rem;
    color: #4b5563;
    margin: 0 0 1.5rem 0;
  }

  :global(.dark) .password-gate-notice {
    color: #d1d5db;
  }

  .password-gate-form {
    width: 100%;
  }

  .password-gate-input-group {
    display: flex;
    gap: 0.5rem;
    width: 100%;
  }

  .password-gate-input {
    flex: 1;
    padding: 0.625rem 0.875rem;
    border: 1px solid #d1d5db;
    border-radius: 0.5rem;
    font-size: 0.875rem;
    color: #111827;
    background: #ffffff;
    outline: none;
    transition: border-color 0.2s ease;
  }

  .password-gate-input:focus {
    border-color: #3b82f6;
    box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.2);
  }

  :global(.dark) .password-gate-input {
    background: #1f2937;
    border-color: #4b5563;
    color: #f3f4f6;
  }

  :global(.dark) .password-gate-input:focus {
    border-color: #60a5fa;
    box-shadow: 0 0 0 2px rgba(96, 165, 250, 0.2);
  }

  .password-gate-submit {
    padding: 0.625rem 1.25rem;
    border: none;
    border-radius: 0.5rem;
    font-size: 0.875rem;
    font-weight: 600;
    color: #ffffff;
    background: #2563eb;
    cursor: pointer;
    transition: background-color 0.2s ease, opacity 0.2s ease;
    white-space: nowrap;
  }

  .password-gate-submit:hover:not(:disabled) {
    background: #1d4ed8;
  }

  .password-gate-submit:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  :global(.dark) .password-gate-submit {
    background: #1d4ed8;
  }

  :global(.dark) .password-gate-submit:hover:not(:disabled) {
    background: #2563eb;
  }

  .password-gate-error {
    font-size: 0.875rem;
    color: #dc2626;
  }

  .password-gate-error:not(:empty) {
    margin-top: 0.75rem;
  }

  :global(.dark) .password-gate-error {
    color: #f87171;
  }
</style>
