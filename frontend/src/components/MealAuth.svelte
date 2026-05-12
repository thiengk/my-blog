<script>
  /**
   * MealAuth Component
   * Gate component that requires group password before showing meal scheduler content.
   * Validates by calling the API with the provided secret.
   */

  const API_URL = import.meta.env.PUBLIC_API_URL || '';
  const STORAGE_KEY = 'meal-group-secret';

  let authenticated = $state(false);
  let checking = $state(true);
  let secretInput = $state('');
  let error = $state('');
  let loading = $state(false);

  /**
   * Get stored secret from sessionStorage.
   * @returns {string|null}
   */
  function getStoredSecret() {
    try {
      return sessionStorage.getItem(STORAGE_KEY);
    } catch {
      return null;
    }
  }

  /**
   * Save secret to sessionStorage.
   * @param {string} secret
   */
  function saveSecret(secret) {
    try {
      sessionStorage.setItem(STORAGE_KEY, secret);
    } catch {
      // sessionStorage unavailable
    }
  }

  /**
   * Validate secret by calling the API.
   * @param {string} secret
   * @returns {Promise<boolean>}
   */
  async function validateSecret(secret) {
    try {
      const response = await fetch(`${API_URL}/api/meals/members`, {
        headers: { 'X-Group-Secret': secret },
      });
      return response.ok;
    } catch {
      return false;
    }
  }

  // Check stored secret on mount
  $effect(() => {
    const stored = getStoredSecret();
    if (stored) {
      validateSecret(stored).then((valid) => {
        if (valid) {
          authenticated = true;
        }
        checking = false;
      });
    } else {
      checking = false;
    }
  });

  /**
   * Handle form submission.
   */
  async function handleSubmit(e) {
    e.preventDefault();
    if (!secretInput.trim() || loading) return;

    loading = true;
    error = '';

    const valid = await validateSecret(secretInput);
    if (valid) {
      saveSecret(secretInput);
      authenticated = true;
    } else {
      error = 'Mật khẩu nhóm không đúng. Vui lòng thử lại.';
      secretInput = '';
    }
    loading = false;
  }

  /**
   * Get the current secret for API calls.
   * @returns {string}
   */
  export function getSecret() {
    return getStoredSecret() || '';
  }
</script>

{#if checking}
  <div class="meal-auth-loading">
    <p>Đang kiểm tra...</p>
  </div>
{:else if !authenticated}
  <div class="meal-auth-gate">
    <div class="meal-auth-card">
      <div class="meal-auth-icon">
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lock-icon" aria-hidden="true">
          <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
          <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
        </svg>
      </div>
      <h2>Lịch ăn nhóm</h2>
      <p class="meal-auth-desc">Nhập mật khẩu nhóm để truy cập</p>

      <form onsubmit={handleSubmit}>
        <input
          type="password"
          bind:value={secretInput}
          placeholder="Mật khẩu nhóm"
          aria-label="Nhập mật khẩu nhóm"
          disabled={loading}
          maxlength="128"
        />
        <button type="submit" disabled={!secretInput.trim() || loading} aria-label="Đăng nhập">
          {loading ? 'Đang xác thực...' : 'Đăng nhập'}
        </button>
      </form>

      {#if error}
        <p class="meal-auth-error" role="alert" aria-live="assertive">{error}</p>
      {/if}
    </div>
  </div>
{:else}
  <slot />
{/if}

<style>
  .meal-auth-loading {
    display: flex;
    justify-content: center;
    align-items: center;
    min-height: 300px;
    color: var(--text-color, #4a5568);
  }

  .meal-auth-gate {
    display: flex;
    justify-content: center;
    align-items: center;
    min-height: 400px;
    padding: 2rem;
  }

  .meal-auth-card {
    max-width: 400px;
    width: 100%;
    padding: 2rem;
    border: 1px solid var(--border-color, #e2e8f0);
    border-radius: 1rem;
    text-align: center;
    background: var(--card-bg, #ffffff);
  }

  .meal-auth-icon {
    margin-bottom: 1rem;
  }

  .lock-icon {
    width: 3rem;
    height: 3rem;
    color: var(--text-muted, #718096);
  }

  h2 {
    margin: 0 0 0.5rem;
    font-size: 1.5rem;
    color: var(--text-color, #1a202c);
  }

  .meal-auth-desc {
    margin: 0 0 1.5rem;
    color: var(--text-muted, #718096);
    font-size: 0.875rem;
  }

  form {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  input {
    padding: 0.75rem 1rem;
    border: 1px solid var(--border-color, #e2e8f0);
    border-radius: 0.5rem;
    font-size: 1rem;
    background: var(--input-bg, #ffffff);
    color: var(--text-color, #1a202c);
    outline: none;
    transition: border-color 0.2s;
  }

  input:focus {
    border-color: var(--primary-color, #3b82f6);
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
  }

  button {
    padding: 0.75rem 1rem;
    border: none;
    border-radius: 0.5rem;
    background: var(--primary-color, #3b82f6);
    color: #ffffff;
    font-size: 1rem;
    font-weight: 500;
    cursor: pointer;
    transition: background 0.2s;
  }

  button:hover:not(:disabled) {
    background: var(--primary-hover, #2563eb);
  }

  button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .meal-auth-error {
    margin-top: 0.75rem;
    color: #ef4444;
    font-size: 0.875rem;
  }

  :global(.dark) .meal-auth-card {
    background: var(--card-bg, #1e293b);
    border-color: var(--border-color, #334155);
  }

  :global(.dark) h2 {
    color: var(--text-color, #f1f5f9);
  }

  :global(.dark) input {
    background: var(--input-bg, #0f172a);
    border-color: var(--border-color, #334155);
    color: var(--text-color, #f1f5f9);
  }

  :global(.dark) .meal-auth-error {
    color: #f87171;
  }
</style>
