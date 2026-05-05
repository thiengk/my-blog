/**
 * Reading Time Calculator
 *
 * Calculates estimated reading time based on word count.
 * Uses average reading speed of 200 words per minute.
 *
 * Validates: Requirements 1.4
 */

const WORDS_PER_MINUTE = 200;

/**
 * Strip HTML tags from content to get plain text.
 */
export function stripHTML(html: string): string {
  return html.replace(/<[^>]*>/g, ' ').replace(/\s+/g, ' ').trim();
}

/**
 * Strip Markdown syntax from content to get plain text.
 * Removes common Markdown elements: headings, links, images, bold, italic, code, etc.
 */
export function stripMarkdown(markdown: string): string {
  return markdown
    // Remove code blocks (fenced)
    .replace(/```[\s\S]*?```/g, '')
    // Remove inline code
    .replace(/`[^`]*`/g, '')
    // Remove images
    .replace(/!\[.*?\]\(.*?\)/g, '')
    // Remove links but keep text
    .replace(/\[([^\]]*)\]\(.*?\)/g, '$1')
    // Remove headings markers
    .replace(/^#{1,6}\s+/gm, '')
    // Remove bold/italic markers
    .replace(/(\*{1,3}|_{1,3})(.*?)\1/g, '$2')
    // Remove strikethrough
    .replace(/~~(.*?)~~/g, '$1')
    // Remove blockquotes
    .replace(/^\s*>\s?/gm, '')
    // Remove horizontal rules
    .replace(/^[-*_]{3,}\s*$/gm, '')
    // Remove list markers
    .replace(/^\s*[-*+]\s+/gm, '')
    .replace(/^\s*\d+\.\s+/gm, '')
    // Remove HTML tags
    .replace(/<[^>]*>/g, ' ')
    // Remove frontmatter
    .replace(/^---[\s\S]*?---/m, '')
    // Collapse whitespace
    .replace(/\s+/g, ' ')
    .trim();
}

/**
 * Count words in a text string.
 * Handles multiple languages including Vietnamese.
 */
export function countWords(text: string): number {
  if (!text || text.trim().length === 0) {
    return 0;
  }

  // Split by whitespace and filter empty strings
  const words = text.trim().split(/\s+/).filter((word) => word.length > 0);
  return words.length;
}

/**
 * Calculate reading time in minutes from word count.
 * Minimum reading time is 1 minute.
 */
export function calculateReadingTime(wordCount: number): number {
  if (wordCount <= 0) {
    return 0;
  }
  return Math.max(1, Math.ceil(wordCount / WORDS_PER_MINUTE));
}

/**
 * Format reading time as a Vietnamese string.
 */
export function formatReadingTime(minutes: number): string {
  if (minutes <= 0) {
    return '0 phút đọc';
  }
  return `${minutes} phút đọc`;
}

/**
 * Get reading time from raw Markdown content.
 * Returns formatted string like "5 phút đọc".
 */
export function getReadingTimeFromMarkdown(markdown: string): string {
  const plainText = stripMarkdown(markdown);
  const wordCount = countWords(plainText);
  const minutes = calculateReadingTime(wordCount);
  return formatReadingTime(minutes);
}

/**
 * Get reading time from rendered HTML content.
 * Returns formatted string like "5 phút đọc".
 */
export function getReadingTimeFromHTML(html: string): string {
  const plainText = stripHTML(html);
  const wordCount = countWords(plainText);
  const minutes = calculateReadingTime(wordCount);
  return formatReadingTime(minutes);
}

/**
 * Get reading time details (for when you need both minutes and word count).
 */
export function getReadingTimeDetails(content: string, isHTML = false): {
  wordCount: number;
  minutes: number;
  text: string;
} {
  const plainText = isHTML ? stripHTML(content) : stripMarkdown(content);
  const wordCount = countWords(plainText);
  const minutes = calculateReadingTime(wordCount);
  const text = formatReadingTime(minutes);
  return { wordCount, minutes, text };
}
