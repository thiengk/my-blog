<script>
  /** @type {'light' | 'dark'} */
  let theme = $state('light');

  /**
   * Detect system color scheme preference.
   * @returns {'light' | 'dark'}
   */
  function detectSystemPreference() {
    if (typeof window === 'undefined') return 'light';
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
  }

  /**
   * Get the current theme from localStorage or system preference.
   * @returns {'light' | 'dark'}
   */
  function getCurrentTheme() {
    if (typeof localStorage !== 'undefined' && localStorage.getItem('theme')) {
      return /** @type {'light' | 'dark'} */ (localStorage.getItem('theme'));
    }
    return detectSystemPreference();
  }

  /**
   * Persist theme choice to localStorage.
   * @param {'light' | 'dark'} value
   */
  function persistTheme(value) {
    if (typeof localStorage !== 'undefined') {
      localStorage.setItem('theme', value);
    }
  }

  /**
   * Apply theme to the document by toggling the 'dark' class.
   * @param {'light' | 'dark'} value
   */
  function applyTheme(value) {
    if (typeof document !== 'undefined') {
      if (value === 'dark') {
        document.documentElement.classList.add('dark');
      } else {
        document.documentElement.classList.remove('dark');
      }
    }
  }

  /** Toggle between dark and light mode. */
  function toggleTheme() {
    theme = theme === 'dark' ? 'light' : 'dark';
    applyTheme(theme);
    persistTheme(theme);
  }

  // Initialize theme on mount
  $effect(() => {
    theme = getCurrentTheme();

    // Listen for system preference changes
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');

    /** @param {MediaQueryListEvent} e */
    function handleChange(e) {
      // Only react to system changes if user hasn't set a preference
      if (!localStorage.getItem('theme')) {
        theme = e.matches ? 'dark' : 'light';
        applyTheme(theme);
      }
    }

    mediaQuery.addEventListener('change', handleChange);

    return () => {
      mediaQuery.removeEventListener('change', handleChange);
    };
  });
</script>

<button
  onclick={toggleTheme}
  class="relative flex items-center justify-center w-9 h-9 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 dark:focus:ring-offset-gray-900"
  aria-label={theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
  title={theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
>
  <!-- Sun icon (shown in dark mode) -->
  <svg
    xmlns="http://www.w3.org/2000/svg"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    stroke-width="2"
    stroke-linecap="round"
    stroke-linejoin="round"
    class="w-5 h-5 text-gray-700 dark:text-gray-300 transition-transform duration-300 {theme === 'dark' ? 'scale-100 rotate-0' : 'scale-0 rotate-90'} absolute"
    aria-hidden="true"
  >
    <circle cx="12" cy="12" r="5" />
    <line x1="12" y1="1" x2="12" y2="3" />
    <line x1="12" y1="21" x2="12" y2="23" />
    <line x1="4.22" y1="4.22" x2="5.64" y2="5.64" />
    <line x1="18.36" y1="18.36" x2="19.78" y2="19.78" />
    <line x1="1" y1="12" x2="3" y2="12" />
    <line x1="21" y1="12" x2="23" y2="12" />
    <line x1="4.22" y1="19.78" x2="5.64" y2="18.36" />
    <line x1="18.36" y1="5.64" x2="19.78" y2="4.22" />
  </svg>

  <!-- Moon icon (shown in light mode) -->
  <svg
    xmlns="http://www.w3.org/2000/svg"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    stroke-width="2"
    stroke-linecap="round"
    stroke-linejoin="round"
    class="w-5 h-5 text-gray-700 dark:text-gray-300 transition-transform duration-300 {theme === 'light' ? 'scale-100 rotate-0' : 'scale-0 -rotate-90'} absolute"
    aria-hidden="true"
  >
    <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" />
  </svg>
</button>
