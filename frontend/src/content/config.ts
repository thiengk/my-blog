import { defineCollection, z } from 'astro:content';

const blogCollection = defineCollection({
  type: 'content',
  schema: z.object({
    title: z.string().min(1, 'Title is required'),
    description: z.string().min(1, 'Description is required'),
    date: z.coerce.date(),
    updatedDate: z.coerce.date().optional(),
    category: z.string().min(1, 'Category is required'),
    tags: z.array(z.string()).default([]),
    coverImage: z.string().optional(),
    draft: z.boolean().default(false),
    slug: z.string().optional(),
  }),
});

export const collections = {
  blog: blogCollection,
};
