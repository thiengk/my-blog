import type { APIRoute } from 'astro';
import { getCollection } from 'astro:content';

export const GET: APIRoute = async () => {
  const posts = await getCollection('blog', ({ data }) => !data.draft);

  const searchIndex = {
    posts: posts.map((post) => ({
      slug: post.slug,
      title: post.data.title,
      description: post.data.description,
      tags: post.data.tags,
      category: post.data.category,
      content: stripMarkdown(post.body || ''),
      date: post.data.date.toISOString().split('T')[0],
    })),
  };

  return new Response(JSON.stringify(searchIndex), {
    headers: {
      'Content-Type': 'application/json',
    },
  });
};

/**
 * Strip Markdown/HTML syntax to get plain text for search indexing.
 */
function stripMarkdown(content: string): string {
  return (
    content
      // Remove frontmatter
      .replace(/^---[\s\S]*?---\n?/, '')
      // Remove HTML tags
      .replace(/<[^>]*>/g, '')
      // Remove headings markers
      .replace(/^#{1,6}\s+/gm, '')
      // Remove bold/italic markers
      .replace(/\*\*([^*]+)\*\*/g, '$1')
      .replace(/\*([^*]+)\*/g, '$1')
      .replace(/__([^_]+)__/g, '$1')
      .replace(/_([^_]+)_/g, '$1')
      // Remove links but keep text
      .replace(/\[([^\]]+)\]\([^)]+\)/g, '$1')
      // Remove images
      .replace(/!\[([^\]]*)\]\([^)]+\)/g, '$1')
      // Remove code blocks
      .replace(/```[\s\S]*?```/g, '')
      // Remove inline code
      .replace(/`([^`]+)`/g, '$1')
      // Remove blockquotes
      .replace(/^>\s+/gm, '')
      // Remove horizontal rules
      .replace(/^[-*_]{3,}\s*$/gm, '')
      // Remove list markers
      .replace(/^[\s]*[-*+]\s+/gm, '')
      .replace(/^[\s]*\d+\.\s+/gm, '')
      // Collapse multiple newlines
      .replace(/\n{3,}/g, '\n\n')
      .trim()
  );
}
