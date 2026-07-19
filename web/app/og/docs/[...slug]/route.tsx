import { notFound } from 'next/navigation';
import { socialCard } from '@/lib/og';
import { getPageImage, source } from '@/lib/source';

export const revalidate = false;

export async function GET(
  _req: Request,
  { params }: RouteContext<'/og/docs/[...slug]'>,
) {
  const { slug } = await params;
  const page = source.getPage(slug.slice(0, -1));
  if (!page) notFound();

  return socialCard({
    title: page.data.title,
    description: page.data.description,
    context: `docs / ${page.slugs.join(' / ') || 'introduction'}`,
  });
}

export function generateStaticParams() {
  return source.getPages().map((page) => ({
    lang: page.locale,
    slug: getPageImage(page).segments,
  }));
}
