import type { Metadata } from 'next';
import { site } from '@/lib/site';

export const siteUrl = 'https://projects.piyushgambhir.com/cubeapm-cli';
export const repoUrl = `https://github.com/${site.repo}`;
export const siteMetadataDescription =
  'Independent, unofficial CubeAPM CLI for any coding agent or shell harness to manage traces, metrics, and logs with JSON/YAML, read-only, no-input automation.';
export const socialMetadataDescription =
  'Agent-ready and harness-agnostic, this independent, unofficial CubeAPM CLI manages traces, metrics, and logs with JSON/YAML, read-only, no-input automation.';
export const defaultSocialImage = {
  url: `${siteUrl}/og/docs/image.png`,
  width: 1200,
  height: 630,
  alt: `${site.name} documentation`,
};

export function withProjectIndependence(description?: string) {
  const summary = description?.trim() || `${site.name} documentation.`;

  return `${summary} ${site.name} is an independent, unofficial open-source CLI and is not affiliated with CubeAPM or its vendor.`;
}

export function withAgentReadyProjectContext(description?: string) {
  const summary = description?.trim() || `${site.name} documentation.`;

  return `${summary} Agent-ready for any coding agent or shell harness. Independent and unofficial.`;
}

export function createPageMetadata(
  title: string,
  description: string,
  path: string,
): Metadata {
  return {
    title,
    description,
    alternates: { canonical: `${siteUrl}${path}` },
    openGraph: {
      type: 'website',
      title,
      description,
      url: `${siteUrl}${path}`,
      siteName: site.name,
      locale: 'en_US',
      images: [defaultSocialImage],
    },
    twitter: {
      card: 'summary_large_image',
      title,
      description,
      images: [defaultSocialImage],
    },
  };
}

export function serializeJsonLd(value: unknown) {
  return JSON.stringify(value).replace(/</g, '\\u003c');
}
