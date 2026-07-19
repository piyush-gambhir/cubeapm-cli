import { socialCard } from '@/lib/og';
import { site } from '@/lib/site';
import { siteUrl } from '@/lib/shared';

export const revalidate = false;

export function GET() {
  return socialCard({
    title: site.tagline,
    description: site.description,
    context: siteUrl.replace(/^https?:\/\//, ''),
  });
}
