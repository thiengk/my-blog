/**
 * TOC (Table of Contents) Generator
 *
 * Extracts headings (h2, h3, h4) from rendered HTML content
 * and generates a nested structure for table of contents navigation.
 *
 * Validates: Requirements 1.3
 */

export interface TOCItem {
  id: string;
  text: string;
  depth: number;
  children: TOCItem[];
}

/**
 * Slugify a heading text to create a URL-friendly anchor ID.
 * Handles Unicode characters (Vietnamese, etc.) and special characters.
 */
export function slugify(text: string): string {
  return text
    .toLowerCase()
    .trim()
    .replace(/[^\p{L}\p{N}\s-]/gu, '') // Remove non-letter, non-number, non-space, non-hyphen (Unicode-aware)
    .replace(/[\s_]+/g, '-') // Replace spaces and underscores with hyphens
    .replace(/-+/g, '-') // Collapse multiple hyphens
    .replace(/^-|-$/g, ''); // Remove leading/trailing hyphens
}

/**
 * Extract headings from rendered HTML string.
 * Parses h2, h3, h4 elements and returns flat list of heading info.
 */
export function extractHeadings(html: string): { id: string; text: string; depth: number }[] {
  const headingRegex = /<h([2-4])(?:\s+[^>]*id="([^"]*)"[^>]*)?>(.*?)<\/h[2-4]>/gi;
  const headings: { id: string; text: string; depth: number }[] = [];

  let match: RegExpExecArray | null;
  while ((match = headingRegex.exec(html)) !== null) {
    const depth = parseInt(match[1], 10);
    const existingId = match[2] || '';
    // Strip HTML tags from heading content to get plain text
    const text = match[3].replace(/<[^>]*>/g, '').trim();
    const id = existingId || slugify(text);

    if (text) {
      headings.push({ id, text, depth });
    }
  }

  return headings;
}

/**
 * Generate a nested TOC structure from a flat list of headings.
 * Headings are nested based on their depth level (h2 > h3 > h4).
 */
export function generateTOC(headings: { id: string; text: string; depth: number }[]): TOCItem[] {
  const toc: TOCItem[] = [];

  if (headings.length === 0) {
    return toc;
  }

  for (const heading of headings) {
    const item: TOCItem = {
      id: heading.id,
      text: heading.text,
      depth: heading.depth,
      children: [],
    };

    if (heading.depth === 2) {
      toc.push(item);
    } else if (heading.depth === 3) {
      const parent = toc[toc.length - 1];
      if (parent) {
        parent.children.push(item);
      } else {
        // No h2 parent, add as top-level
        toc.push(item);
      }
    } else if (heading.depth === 4) {
      const parentH2 = toc[toc.length - 1];
      if (parentH2 && parentH2.children.length > 0) {
        const parentH3 = parentH2.children[parentH2.children.length - 1];
        parentH3.children.push(item);
      } else if (parentH2) {
        parentH2.children.push(item);
      } else {
        toc.push(item);
      }
    }
  }

  return toc;
}

/**
 * Generate TOC directly from HTML content.
 * Convenience function combining extraction and nesting.
 */
export function generateTOCFromHTML(html: string): TOCItem[] {
  const headings = extractHeadings(html);
  return generateTOC(headings);
}
