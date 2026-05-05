import type { APIContext } from 'astro';
import { getCollection } from 'astro:content';

const SITE_URL = 'https://blog.example.com';
const SITE_TITLE = 'Personal Blog';
const SITE_DESCRIPTION = 'Chia sẻ kinh nghiệm, review khóa học, và câu chuyện cuộc sống';
const MAX_ITEMS = 20;

function escapeXml(str: string): string {
  return str
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&apos;');
}

function formatRFC822Date(date: Date): string {
  return date.toUTCString();
}

export async function GET(context: APIContext): Promise<Response> {
  const allPosts = await getCollection('blog', ({ data }) => !data.draft);

  const sortedPosts = allPosts
    .sort((a, b) => b.data.date.valueOf() - a.data.date.valueOf())
    .slice(0, MAX_ITEMS);

  const lastBuildDate = sortedPosts.length > 0
    ? formatRFC822Date(sortedPosts[0].data.date)
    : formatRFC822Date(new Date());

  const items = sortedPosts.map((post) => {
    const slug = post.data.slug || post.slug;
    const link = `${SITE_URL}/blog/${slug}/`;
    return `    <item>
      <title>${escapeXml(post.data.title)}</title>
      <description>${escapeXml(post.data.description)}</description>
      <link>${link}</link>
      <guid isPermaLink="true">${link}</guid>
      <pubDate>${formatRFC822Date(post.data.date)}</pubDate>
      <category>${escapeXml(post.data.category)}</category>
    </item>`;
  });

  const rss = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">
  <channel>
    <title>${escapeXml(SITE_TITLE)}</title>
    <description>${escapeXml(SITE_DESCRIPTION)}</description>
    <link>${SITE_URL}</link>
    <language>vi</language>
    <lastBuildDate>${lastBuildDate}</lastBuildDate>
    <generator>Astro</generator>
    <atom:link href="${SITE_URL}/rss.xml" rel="self" type="application/rss+xml"/>
${items.join('\n')}
  </channel>
</rss>`;

  return new Response(rss.trim(), {
    headers: {
      'Content-Type': 'application/xml; charset=utf-8',
    },
  });
}
