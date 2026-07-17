import { getPageImage, getPageMarkdownUrl, source } from '@/lib/source';
import {
  DocsBody,
  DocsDescription,
  DocsPage,
  DocsTitle,
  MarkdownCopyButton,
  ViewOptionsPopover,
} from 'fumadocs-ui/layouts/docs/page';
import { notFound } from 'next/navigation';
import Link from 'next/link';
import { getMDXComponents } from '@/components/mdx';
import type { Metadata } from 'next';
import { createRelativeLink } from 'fumadocs-ui/mdx';
import { gitConfig } from '@/lib/shared';
import {
  repoUrl,
  serializeJsonLd,
  siteUrl,
  withAgentReadyProjectContext,
  withProjectIndependence,
} from '@/lib/seo';
import { site } from '@/lib/site';

const relatedDocs: Record<string, { label: string; href: string }[]> = {
  '': [
    { label: 'Installation', href: '/docs/installation' },
    { label: 'Quick start', href: '/docs/quickstart' },
  ],
  installation: [
    { label: 'Authentication & connections', href: '/docs/authentication' },
    { label: 'Quick start', href: '/docs/quickstart' },
  ],
  authentication: [
    { label: 'Quick start', href: '/docs/quickstart' },
    { label: 'Agents', href: '/docs/agents' },
  ],
  quickstart: [
    { label: 'Commands', href: '/docs/commands' },
    { label: 'Output formats', href: '/docs/output-formats' },
  ],
  'output-formats': [
    { label: 'Agents', href: '/docs/agents' },
    { label: 'Commands', href: '/docs/commands' },
  ],
  agents: [
    { label: 'Output formats', href: '/docs/output-formats' },
    { label: 'Authentication & connections', href: '/docs/authentication' },
  ],
  commands: [
    { label: 'Traces', href: '/docs/commands/traces' },
    { label: 'Metrics', href: '/docs/commands/metrics' },
    { label: 'Logs', href: '/docs/commands/logs' },
    { label: 'Ingest & configuration', href: '/docs/commands/ingest-config' },
  ],
  'commands/traces': [
    { label: 'Commands overview', href: '/docs/commands' },
    { label: 'Metrics', href: '/docs/commands/metrics' },
    { label: 'Logs', href: '/docs/commands/logs' },
  ],
  'commands/metrics': [
    { label: 'Commands overview', href: '/docs/commands' },
    { label: 'Traces', href: '/docs/commands/traces' },
    { label: 'Logs', href: '/docs/commands/logs' },
  ],
  'commands/logs': [
    { label: 'Commands overview', href: '/docs/commands' },
    { label: 'Traces', href: '/docs/commands/traces' },
    { label: 'Ingest & configuration', href: '/docs/commands/ingest-config' },
  ],
  'commands/ingest-config': [
    { label: 'Commands overview', href: '/docs/commands' },
    { label: 'Authentication & connections', href: '/docs/authentication' },
    { label: 'Agents', href: '/docs/agents' },
  ],
};

export default async function Page(props: PageProps<'/docs/[[...slug]]'>) {
  const params = await props.params;
  const page = source.getPage(params.slug);
  if (!page) notFound();

  const MDX = page.data.body;
  const markdownUrl = getPageMarkdownUrl(page).url;
  const slug = page.slugs.join('/');
  const pageUrl = `${siteUrl}${page.url}`;
  const description = withProjectIndependence(page.data.description);
  const breadcrumbPages = page.slugs
    .map((_, index) => source.getPage(page.slugs.slice(0, index + 1)))
    .filter((entry) => entry !== undefined);
  const breadcrumbs = [
    {
      '@type': 'ListItem',
      position: 1,
      name: 'Home',
      item: siteUrl,
    },
    {
      '@type': 'ListItem',
      position: 2,
      name: 'Documentation',
      item: `${siteUrl}/docs`,
    },
    ...breadcrumbPages.map((entry, index) => ({
      '@type': 'ListItem',
      position: index + 3,
      name: entry.data.title,
      item: `${siteUrl}${entry.url}`,
    })),
  ];
  const structuredData = {
    '@context': 'https://schema.org',
    '@graph': [
      {
        '@type': 'BreadcrumbList',
        itemListElement: breadcrumbs,
      },
      {
        '@type': 'TechArticle',
        headline: page.data.title,
        description,
        url: pageUrl,
        mainEntityOfPage: pageUrl,
        license: `${repoUrl}/blob/main/LICENSE`,
        author: {
          '@type': 'Person',
          name: 'Piyush Gambhir',
          url: 'https://github.com/piyush-gambhir',
        },
        publisher: {
          '@type': 'Person',
          name: 'Piyush Gambhir',
          url: 'https://github.com/piyush-gambhir',
        },
        isPartOf: { '@id': `${siteUrl}/#website` },
        sameAs: `${repoUrl}/blob/${gitConfig.branch}/content/docs/${page.path}`,
      },
    ],
  };

  return (
    <>
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: serializeJsonLd(structuredData) }}
      />
      <DocsPage toc={page.data.toc} full={page.data.full}>
        <DocsTitle>{page.data.title}</DocsTitle>
        <DocsDescription className="mb-0">{page.data.description}</DocsDescription>
        <div className="flex flex-row gap-2 items-center pb-6">
          <MarkdownCopyButton markdownUrl={markdownUrl} />
          <ViewOptionsPopover
            markdownUrl={markdownUrl}
            githubUrl={`https://github.com/${gitConfig.user}/${gitConfig.repo}/blob/${gitConfig.branch}/content/docs/${page.path}`}
          />
        </div>
        <DocsBody>
          <MDX
            components={getMDXComponents({
              // this allows you to link to other pages with relative file paths
              a: createRelativeLink(source, page),
            })}
          />
          <nav className="docs-related" aria-label="Related documentation">
            <p>Related documentation</p>
            <ul>
              {(relatedDocs[slug] ?? relatedDocs['']).map((item) => (
                <li key={item.href}>
                  <Link href={item.href}>{item.label}</Link>
                </li>
              ))}
            </ul>
          </nav>
        </DocsBody>
      </DocsPage>
    </>
  );
}

export async function generateStaticParams() {
  return source.generateParams();
}

export async function generateMetadata(props: PageProps<'/docs/[[...slug]]'>): Promise<Metadata> {
  const params = await props.params;
  const page = source.getPage(params.slug);
  if (!page) notFound();
  const description = withAgentReadyProjectContext(page.data.description);
  const image = {
    url: getPageImage(page).url,
    width: 1200,
    height: 630,
    alt: `${page.data.title} | ${site.name}`,
  };

  return {
    title: page.data.title,
    description,
    alternates: { canonical: page.url },
    openGraph: {
      type: 'article',
      title: page.data.title,
      description,
      url: page.url,
      siteName: site.name,
      locale: 'en_US',
      images: [image],
    },
    twitter: {
      card: 'summary_large_image',
      title: page.data.title,
      description,
      images: [image],
    },
  };
}
