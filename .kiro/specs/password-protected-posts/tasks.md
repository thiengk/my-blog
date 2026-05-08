# Implementation Plan: Password Protected Posts

## Overview

Implement client-side password protection for blog posts using Astro frontmatter configuration and a Svelte 5 PasswordGate component. The feature adds a `password` field to the content schema, creates a password gate UI that blocks content until the correct password is entered, persists unlock state in sessionStorage, and displays lock indicators in blog listings. No backend changes required.

## Tasks

- [x] 1. Set up utility functions and content schema
  - [x] 1.1 Add `password` field to blog content schema
    - Modify `src/content/config.ts` to add optional `password` field with `z.string().max(128).optional()`
    - _Requirements: 1.1_

  - [x] 1.2 Create `src/lib/password-utils.ts` with helper functions
    - Implement `isProtectedPost(password: string | undefined): boolean` — returns true if password is defined, non-empty, and not whitespace-only
    - Implement `getStorageKey(slug: string): string` — returns `"protected-post:{slug}"`
    - Implement `isUnlockedInSession(slug: string): boolean` — checks sessionStorage for unlock state, returns false if sessionStorage unavailable
    - Implement `saveUnlockState(slug: string): void` — saves `"true"` to sessionStorage, no-op if unavailable
    - Implement `isSessionStorageAvailable(): boolean` — tests sessionStorage availability with try/catch
    - _Requirements: 1.2, 1.4, 4.1, 4.2, 4.4, 4.5_

  - [x]* 1.3 Write property tests for `password-utils.ts`
    - **Property 1: isProtectedPost classification**
    - **Property 5: Session storage round-trip**
    - **Validates: Requirements 1.2, 1.4, 4.1, 4.2, 4.4**
    - Install `fast-check` and `vitest` as dev dependencies
    - Create `tests/property/password-utils.property.ts`
    - Test that `isProtectedPost` returns true for any non-empty, non-whitespace string
    - Test that `isProtectedPost` returns false for undefined, empty, or whitespace-only strings
    - Test session storage round-trip: saveUnlockState → isUnlockedInSession returns true, stored value is "true"

  - [x]* 1.4 Write unit tests for `password-utils.ts`
    - Create `tests/unit/password-utils.test.ts`
    - Test specific examples: undefined, "", "  ", "abc", "pass word", 128-char string
    - Test `getStorageKey` format
    - Test graceful handling when sessionStorage throws
    - _Requirements: 1.2, 1.4, 4.1, 4.4, 4.5_

- [x] 2. Implement PasswordGate Svelte component
  - [x] 2.1 Create `src/components/PasswordGate.svelte`
    - Accept props: `password`, `slug`, `title`, `date`, `category?`, `tags?`
    - Use Svelte 5 runes: `$state` for `unlocked`, `error`, `inputValue`, `checking`
    - On mount: check sessionStorage via `isUnlockedInSession(slug)`, set `checking = false` after check
    - While `checking` is true, render nothing (prevent flash of password form)
    - When locked: render password form with lock icon, title, metadata, input field, submit button
    - When unlocked: render slot content (use `{#snippet}` or `<slot>` pattern)
    - Display metadata (category, tags) only if defined — skip undefined fields without empty placeholders
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.6, 2.7, 7.1, 7.2_

  - [x] 2.2 Implement password validation logic in PasswordGate
    - On submit: compare `inputValue` with `password` prop using strict equality (`===`), no trimming
    - On correct password: call `saveUnlockState(slug)`, set `unlocked = true`
    - On incorrect password: set error message "Mật khẩu không đúng. Vui lòng thử lại.", clear input, keep focus on input
    - Disable submit button when input is empty or whitespace-only
    - Support Enter key submission via form element
    - Support Escape key to clear input field
    - Set `maxlength="128"` on input field
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5, 3.6_

  - [x] 2.3 Implement accessibility features in PasswordGate
    - Add `aria-label="Nhập mật khẩu để mở khóa bài viết này"` on password input
    - Add `aria-label="Mở khóa bài viết"` on submit button
    - Add error container with `role="alert"` and `aria-live="assertive"`
    - Auto-focus password input on mount (after checking state resolves)
    - Ensure full keyboard navigation: Tab between input/button, Enter to submit, Escape to clear
    - Use colors with minimum 4.5:1 contrast ratio in both light and dark themes
    - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5_

  - [x]* 2.4 Write property tests for PasswordGate validation
    - **Property 2: Correct password unlocks content**
    - **Property 3: Incorrect password is rejected**
    - **Property 4: Whitespace-only input disables submit**
    - **Validates: Requirements 3.1, 3.2, 3.4, 3.5**
    - Create `tests/property/password-gate.property.ts`
    - Test that any matching string unlocks content
    - Test that any non-matching string shows error (including strings differing only by whitespace)
    - Test that any whitespace-only string disables submit button

  - [x]* 2.5 Write component tests for PasswordGate
    - Create `tests/component/PasswordGate.test.ts`
    - Test form rendering: lock icon, input field, submit button, metadata display
    - Test aria attributes presence and values
    - Test auto-focus behavior
    - Test flash prevention (checking state)
    - Test conditional metadata rendering (skip undefined category/tags)
    - _Requirements: 2.2, 2.3, 2.4, 2.6, 2.7, 6.1, 6.2, 6.5_

- [x] 3. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 4. Integrate PasswordGate into blog post page
  - [x] 4.1 Update `src/pages/blog/[...slug].astro` to conditionally render PasswordGate
    - Import `PasswordGate` component and `isProtectedPost` utility
    - Check `post.data.password` using `isProtectedPost()`
    - If protected: render `<PasswordGate client:load>` wrapping the article content, passing password, slug, title, formatted date, category, tags as props
    - If not protected: render existing content as-is (no changes to current behavior)
    - Pass rendered content as slot/children to PasswordGate
    - Ensure content is NOT in DOM when locked (Svelte conditional rendering)
    - _Requirements: 1.2, 1.3, 1.4, 1.5, 7.1, 7.2, 7.3, 7.4_

  - [x]* 4.2 Write property test for DOM content exclusion
    - **Property 6: Content exclusion from DOM when locked**
    - **Validates: Requirements 7.1**
    - Create test verifying content string does not appear in rendered DOM when component is in locked state

- [x] 5. Add lock icon indicator to blog listings
  - [x] 5.1 Update `src/components/BlogCard.astro` to show lock icon
    - Add `isProtected` prop (boolean) to BlogCard interface
    - Display lock icon (SVG) next to post title when `isProtected` is true
    - Keep all other display (title, description, date, tags) unchanged
    - _Requirements: 5.2, 5.3_

  - [x] 5.2 Update blog listing pages to pass `isProtected` prop
    - Update `src/pages/blog/index.astro` to import `isProtectedPost` and pass `isProtected={isProtectedPost(post.data.password)}` to BlogCard
    - Update `src/pages/blog/[...page].astro` with same logic
    - Update `src/pages/index.astro` if it renders BlogCard components
    - Update `src/pages/categories/` pages to pass `isProtected` prop
    - Update `src/pages/tags/` pages to pass `isProtected` prop
    - Verify Search component results include lock icon (if applicable)
    - _Requirements: 5.1, 5.2, 5.5, 5.6_

  - [x]* 5.3 Write unit tests for BlogCard lock icon
    - Test that lock icon renders when `isProtected` is true
    - Test that lock icon does not render when `isProtected` is false
    - Test that protected posts still show title and description normally
    - _Requirements: 5.2, 5.3_

- [x] 6. Create sample protected blog post
  - [x] 6.1 Add a sample protected blog post for testing
    - Create a markdown file in `src/content/blog/` with `password` field in frontmatter
    - Use a simple test password (e.g., "test123")
    - Include standard content to verify rendering after unlock
    - _Requirements: 1.1, 1.2_

- [x] 7. Final checkpoint - Ensure all tests pass and build succeeds
  - Run `astro build` to verify static site generation works with protected posts
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation
- Property tests validate universal correctness properties from the design document
- Unit tests validate specific examples and edge cases
- The project uses Svelte 5 runes (`$state`, `$effect`) — follow existing patterns in the codebase
- Testing requires installing `vitest` and `fast-check` as dev dependencies (not currently in package.json)
- All password comparison is strict equality — no trimming, no hashing
- SessionStorage key format: `"protected-post:{slug}"`

## Task Dependency Graph

```json
{
  "waves": [
    { "id": 0, "tasks": ["1.1", "1.2"] },
    { "id": 1, "tasks": ["1.3", "1.4", "2.1"] },
    { "id": 2, "tasks": ["2.2", "2.3", "5.1"] },
    { "id": 3, "tasks": ["2.4", "2.5", "4.1", "5.2"] },
    { "id": 4, "tasks": ["4.2", "5.3", "6.1"] }
  ]
}
```
