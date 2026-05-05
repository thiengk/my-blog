<script>
  /**
   * Newsletter subscription form component.
   * Handles email validation, submission states, and API communication.
   *
   * @typedef {'idle' | 'loading' | 'success' | 'error'} FormStatus
   */

  // Configurable API URL - defaults to relative path for same-origin
  const API_URL = import.meta.env.PUBLIC_API_URL || '';

  // State
  let email = $state('');
  /** @type {FormStatus} */
  let status = $state('idle');
  let errorMessage = $state('');
  let validationError = $state('');
  let successMessage = $state('');

  /**
   * Validate email format using basic RFC 5322 regex.
   * @param {string} value - Email to validate
   * @returns {boolean} Whether the email is valid
   */
  function validateEmail(value) {
    if (!value || value.trim() === '') {
      validationError = 'Vui lòng nhập email';
      return false;
    }

    // Basic RFC 5322 email regex
    const emailRegex = /^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$/;

    if (!emailRegex.test(value.trim())) {
      validationError = 'Email không hợp lệ';
      return false;
    }

    validationError = '';
    return true;
  }

  /**
   * Handle form submission.
   * @param {SubmitEvent} event
   */
  async function handleSubmit(event) {
    event.preventDefault();

    // Client-side validation
    if (!validateEmail(email)) {
      return;
    }

    status = 'loading';
    errorMessage = '';

    try {
      const response = await fetch(`${API_URL}/api/newsletter/subscribe`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ email: email.trim() }),
      });

      const data = await response.json();

      if (response.ok && data.success) {
        status = 'success';
        successMessage = data.message || 'Đăng ký thành công! Cảm ơn bạn.';
        email = '';
      } else if (response.status === 409) {
        // Duplicate email
        status = 'error';
        errorMessage = data.message || 'Email này đã được đăng ký trước đó.';
      } else if (response.status === 422) {
        // Validation error from server
        status = 'error';
        errorMessage = data.message || 'Email không hợp lệ.';
      } else if (response.status === 429) {
        // Rate limited
        status = 'error';
        errorMessage = 'Bạn đã gửi quá nhiều yêu cầu. Vui lòng thử lại sau.';
      } else {
        status = 'error';
        errorMessage = data.message || 'Đã có lỗi xảy ra. Vui lòng thử lại sau.';
      }
    } catch (err) {
      status = 'error';
      errorMessage = 'Không thể kết nối đến server. Vui lòng thử lại sau.';
    }
  }

  /**
   * Clear validation error on input.
   */
  function handleInput() {
    if (validationError) {
      validationError = '';
    }
    // Reset error state when user starts typing again
    if (status === 'error') {
      status = 'idle';
      errorMessage = '';
    }
  }

  /**
   * Reset form to idle state (e.g., after success).
   */
  function resetForm() {
    status = 'idle';
    email = '';
    errorMessage = '';
    validationError = '';
    successMessage = '';
  }

  // Derived state
  let hasError = $derived(validationError !== '' || errorMessage !== '');
  let displayError = $derived(validationError || errorMessage);
  let isSubmitDisabled = $derived(status === 'loading');
</script>

<div class="newsletter-form w-full">
  {#if status === 'success'}
    <!-- Success State -->
    <div
      class="flex flex-col items-center gap-3 p-6 rounded-lg bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800"
      role="status"
      aria-live="polite"
    >
      <svg
        xmlns="http://www.w3.org/2000/svg"
        class="w-10 h-10 text-green-500 dark:text-green-400"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor"
        stroke-width="2"
        aria-hidden="true"
      >
        <path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
      <p class="text-sm font-medium text-green-700 dark:text-green-300 text-center">
        {successMessage}
      </p>
      <button
        onclick={resetForm}
        class="text-xs text-green-600 dark:text-green-400 underline underline-offset-2 hover:text-green-700 dark:hover:text-green-300 transition-colors"
      >
        Đăng ký email khác
      </button>
    </div>
  {:else}
    <!-- Form State (idle, loading, error) -->
    <form
      onsubmit={handleSubmit}
      class="flex flex-col sm:flex-row gap-3"
      novalidate
      aria-label="Đăng ký nhận bài viết mới qua email"
    >
      <div class="flex-1 relative">
        <label for="newsletter-email" class="sr-only">Địa chỉ email</label>
        <input
          id="newsletter-email"
          type="email"
          bind:value={email}
          oninput={handleInput}
          placeholder="email@example.com"
          class="input w-full {hasError ? 'border-red-500 dark:border-red-400 focus:ring-red-500 focus:border-red-500' : ''}"
          aria-label="Địa chỉ email để đăng ký newsletter"
          aria-describedby={hasError ? 'newsletter-error' : undefined}
          aria-invalid={hasError ? 'true' : undefined}
          disabled={status === 'loading'}
          autocomplete="email"
        />
      </div>

      <button
        type="submit"
        class="btn btn-primary whitespace-nowrap px-6 py-2.5 disabled:opacity-60 disabled:cursor-not-allowed"
        disabled={isSubmitDisabled}
        aria-label={status === 'loading' ? 'Đang gửi...' : 'Đăng ký nhận bài viết mới'}
      >
        {#if status === 'loading'}
          <svg
            class="animate-spin -ml-1 mr-2 h-4 w-4 text-white"
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
          Đăng ký
        {/if}
      </button>
    </form>

    <!-- Error Message -->
    {#if hasError}
      <p
        id="newsletter-error"
        class="mt-2 text-sm text-red-600 dark:text-red-400"
        role="alert"
        aria-live="assertive"
      >
        {displayError}
      </p>
    {/if}
  {/if}
</div>
