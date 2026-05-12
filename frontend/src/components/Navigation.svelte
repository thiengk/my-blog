<script>
  let isMenuOpen = $state(false);

  function toggleMenu() {
    isMenuOpen = !isMenuOpen;
  }

  function closeMenu() {
    isMenuOpen = false;
  }

  function handleKeydown(event) {
    if (event.key === 'Escape' && isMenuOpen) {
      closeMenu();
    }
  }
</script>

<svelte:window onkeydown={handleKeydown} />

<nav class="relative" aria-label="Main navigation">
  <!-- Desktop Navigation -->
  <ul class="hidden md:flex items-center gap-6 lg:gap-8">
    <li>
      <a
        href="/"
        class="text-gray-700 dark:text-gray-300 hover:text-primary-600 dark:hover:text-primary-400 transition-colors font-medium"
      >
        Home
      </a>
    </li>
    <li>
      <a
        href="/blog"
        class="text-gray-700 dark:text-gray-300 hover:text-primary-600 dark:hover:text-primary-400 transition-colors font-medium"
      >
        Blog
      </a>
    </li>
    <li>
      <a
        href="/categories"
        class="text-gray-700 dark:text-gray-300 hover:text-primary-600 dark:hover:text-primary-400 transition-colors font-medium"
      >
        Categories
      </a>
    </li>
    <li>
      <a
        href="/tags"
        class="text-gray-700 dark:text-gray-300 hover:text-primary-600 dark:hover:text-primary-400 transition-colors font-medium"
      >
        Tags
      </a>
    </li>
    <li>
      <a
        href="/internal/meals"
        class="text-gray-700 dark:text-gray-300 hover:text-primary-600 dark:hover:text-primary-400 transition-colors font-medium"
      >
        🍜 Lịch ăn
      </a>
    </li>
  </ul>

  <!-- Mobile Hamburger Button -->
  <button
    class="md:hidden flex flex-col justify-center items-center w-10 h-10 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
    onclick={toggleMenu}
    aria-label={isMenuOpen ? 'Close menu' : 'Open menu'}
    aria-expanded={isMenuOpen}
    aria-controls="mobile-menu"
  >
    <span
      class="block w-5 h-0.5 bg-gray-700 dark:bg-gray-300 transition-transform duration-300 {isMenuOpen ? 'rotate-45 translate-y-1.5' : ''}"
    ></span>
    <span
      class="block w-5 h-0.5 bg-gray-700 dark:bg-gray-300 transition-opacity duration-300 mt-1 {isMenuOpen ? 'opacity-0' : ''}"
    ></span>
    <span
      class="block w-5 h-0.5 bg-gray-700 dark:bg-gray-300 transition-transform duration-300 mt-1 {isMenuOpen ? '-rotate-45 -translate-y-1.5' : ''}"
    ></span>
  </button>

  <!-- Mobile Menu -->
  {#if isMenuOpen}
    <div
      id="mobile-menu"
      class="absolute top-full right-0 mt-2 w-48 bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700 py-2 z-50 md:hidden"
      role="menu"
    >
      <a
        href="/"
        class="block px-4 py-2 text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 hover:text-primary-600 dark:hover:text-primary-400 transition-colors"
        role="menuitem"
        onclick={closeMenu}
      >
        Home
      </a>
      <a
        href="/blog"
        class="block px-4 py-2 text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 hover:text-primary-600 dark:hover:text-primary-400 transition-colors"
        role="menuitem"
        onclick={closeMenu}
      >
        Blog
      </a>
      <a
        href="/categories"
        class="block px-4 py-2 text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 hover:text-primary-600 dark:hover:text-primary-400 transition-colors"
        role="menuitem"
        onclick={closeMenu}
      >
        Categories
      </a>
      <a
        href="/tags"
        class="block px-4 py-2 text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 hover:text-primary-600 dark:hover:text-primary-400 transition-colors"
        role="menuitem"
        onclick={closeMenu}
      >
        Tags
      </a>
      <a
        href="/internal/meals"
        class="block px-4 py-2 text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 hover:text-primary-600 dark:hover:text-primary-400 transition-colors"
        role="menuitem"
        onclick={closeMenu}
      >
        🍜 Lịch ăn
      </a>
    </div>
  {/if}
</nav>
