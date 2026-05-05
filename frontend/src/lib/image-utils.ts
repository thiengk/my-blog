/**
 * Image Optimization Utilities
 *
 * Provides helper functions for responsive image handling,
 * blur placeholders (LQIP), and width constraints.
 *
 * Validates: Requirements 9.1, 9.2, 9.5, 9.6
 */

/** Default responsive image widths for srcset generation */
export const RESPONSIVE_WIDTHS = [400, 800, 1200, 1600, 2000] as const;

/** Maximum allowed image width */
export const MAX_IMAGE_WIDTH = 2000;

/** Default image quality for WebP output */
export const DEFAULT_QUALITY = 80;

/**
 * Generate a srcset string for responsive images.
 * Produces entries for each width that doesn't exceed the original image width.
 *
 * @param baseSrc - The base image source path or URL
 * @param originalWidth - The original width of the image (optional)
 * @param widths - Array of target widths for srcset
 * @returns A srcset-compatible string
 *
 * @example
 * generateSrcSet('/images/hero.jpg', 1800)
 * // => '/images/hero.jpg?w=400 400w, /images/hero.jpg?w=800 800w, ...'
 */
export function generateSrcSet(
  baseSrc: string,
  originalWidth?: number,
  widths: readonly number[] = RESPONSIVE_WIDTHS
): string {
  const maxWidth = originalWidth
    ? Math.min(originalWidth, MAX_IMAGE_WIDTH)
    : MAX_IMAGE_WIDTH;

  const applicableWidths = widths.filter((w) => w <= maxWidth);

  // If no widths are applicable (image smaller than smallest breakpoint),
  // use the original width
  if (applicableWidths.length === 0 && originalWidth) {
    return `${baseSrc} ${originalWidth}w`;
  }

  return applicableWidths
    .map((w) => {
      const separator = baseSrc.includes('?') ? '&' : '?';
      return `${baseSrc}${separator}w=${w} ${w}w`;
    })
    .join(', ');
}

/**
 * Generate responsive sizes attribute for common layout patterns.
 *
 * @param layout - The layout context: 'full', 'content', or 'card'
 * @returns A sizes attribute string
 */
export function generateSizes(
  layout: 'full' | 'content' | 'card' = 'content'
): string {
  switch (layout) {
    case 'full':
      return '100vw';
    case 'content':
      return '(max-width: 768px) 100vw, (max-width: 1200px) 80vw, 720px';
    case 'card':
      return '(max-width: 768px) 100vw, (max-width: 1200px) 50vw, 400px';
    default:
      return '100vw';
  }
}

/**
 * Generate CSS-based blur placeholder styles (LQIP technique).
 * Uses a solid background color with CSS blur filter to simulate
 * a low-quality image placeholder while the real image loads.
 *
 * @param dominantColor - Dominant color of the image (hex or CSS color)
 * @param aspectRatio - Aspect ratio as width/height (e.g., 16/9)
 * @returns CSS style object for the placeholder
 */
export function getBlurPlaceholder(
  dominantColor: string = '#e2e8f0',
  aspectRatio?: number
): Record<string, string> {
  const styles: Record<string, string> = {
    'background-color': dominantColor,
    'background-size': 'cover',
    filter: 'blur(20px)',
    transform: 'scale(1.1)',
    position: 'absolute',
    inset: '0',
    'z-index': '0',
    transition: 'opacity 0.3s ease-in-out',
  };

  if (aspectRatio) {
    styles['aspect-ratio'] = String(aspectRatio);
  }

  return styles;
}

/**
 * Convert a style object to an inline CSS string.
 *
 * @param styles - Object with CSS property-value pairs
 * @returns Inline CSS string
 */
export function stylesToString(styles: Record<string, string>): string {
  return Object.entries(styles)
    .map(([key, value]) => `${key}: ${value}`)
    .join('; ');
}

/**
 * Constrain image width to the maximum allowed width.
 * If the image is wider than MAX_IMAGE_WIDTH, returns MAX_IMAGE_WIDTH.
 * Otherwise returns the original width.
 *
 * @param width - Original image width in pixels
 * @returns Constrained width (max 2000px)
 */
export function constrainWidth(width: number): number {
  if (width <= 0) {
    return 0;
  }
  return Math.min(width, MAX_IMAGE_WIDTH);
}

/**
 * Constrain image dimensions while preserving aspect ratio.
 * If width exceeds MAX_IMAGE_WIDTH, both dimensions are scaled down proportionally.
 *
 * @param width - Original width in pixels
 * @param height - Original height in pixels
 * @returns Object with constrained width and height
 */
export function constrainDimensions(
  width: number,
  height: number
): { width: number; height: number } {
  if (width <= 0 || height <= 0) {
    return { width: 0, height: 0 };
  }

  if (width <= MAX_IMAGE_WIDTH) {
    return { width, height };
  }

  const ratio = MAX_IMAGE_WIDTH / width;
  return {
    width: MAX_IMAGE_WIDTH,
    height: Math.round(height * ratio),
  };
}

/**
 * Determine if an image should use eager or lazy loading
 * based on its position context.
 *
 * @param priority - Whether the image is above the fold / high priority
 * @returns 'eager' for priority images, 'lazy' for others
 */
export function getLoadingStrategy(priority: boolean): 'eager' | 'lazy' {
  return priority ? 'eager' : 'lazy';
}

/**
 * Get the appropriate fetchpriority value for an image.
 *
 * @param priority - Whether the image is high priority (above the fold)
 * @returns 'high' for priority images, 'auto' for others
 */
export function getFetchPriority(priority: boolean): 'high' | 'auto' {
  return priority ? 'high' : 'auto';
}
