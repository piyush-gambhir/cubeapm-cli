export const dynamic = 'force-static';
import type { MetadataRoute } from 'next';
import { siteUrl } from '@/lib/metadata';
import { source } from '@/lib/source';

export default function sitemap(): MetadataRoute.Sitemap {
  const routes = [
    '/',
    ...source.getPages().map((page) => page.url),
    '/privacy',
    '/terms',
    '/contact',
  ];

  return [...new Set(routes)].map((route) => ({
    url: route === '/' ? siteUrl : `${siteUrl}${route}`,
    changeFrequency: route.startsWith('/docs') ? 'weekly' : 'monthly',
    priority: route === '/' ? 1 : route === '/docs' ? 0.9 : 0.6,
  }));
}
