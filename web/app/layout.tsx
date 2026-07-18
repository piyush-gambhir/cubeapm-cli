import type { Metadata } from 'next';
import '@fontsource-variable/inter';
import '@fontsource-variable/jetbrains-mono';
import '@fontsource/instrument-serif';
import type { CSSProperties } from 'react';
import { Provider } from '@/components/provider';
import { site } from '@/lib/site';
import {
  defaultSocialImage,
  siteMetadataDescription,
  siteUrl,
  socialMetadataDescription,
} from '@/lib/seo';
import './global.css';

export const metadata: Metadata = {
  metadataBase: new URL(siteUrl),
  title: {
    default: `${site.name}: ${site.tagline}`,
    template: `%s · ${site.name}`,
  },
  description: siteMetadataDescription,
  alternates: { canonical: '/' },
  authors: [{ name: 'Piyush Gambhir', url: 'https://github.com/piyush-gambhir' }],
  creator: 'Piyush Gambhir',
  publisher: 'Piyush Gambhir',
  icons: {
    icon: [{ url: '/cubeapm-cli/favicon.svg', type: 'image/svg+xml' }],
  },
  openGraph: {
    type: 'website',
    url: '/',
    siteName: site.name,
    title: `${site.name}: ${site.tagline}`,
    description: socialMetadataDescription,
    locale: 'en_US',
    images: [defaultSocialImage],
  },
  twitter: {
    card: 'summary_large_image',
    title: `${site.name}: ${site.tagline}`,
    description: socialMetadataDescription,
    images: [defaultSocialImage],
  },
};

export default function Layout({ children }: LayoutProps<'/'>) {
  const rootStyle = {
    ...(site.accent ? { '--site-accent': site.accent } : {}),
  } as CSSProperties;

  return (
    <html
      lang="en"
      data-accent={site.accentName}
      style={rootStyle}
      suppressHydrationWarning
    >
      <head>
        <script
          dangerouslySetInnerHTML={{
            __html: "document.documentElement.classList.add('js')",
          }}
        />
      </head>
      <body className="flex flex-col min-h-screen">
        <Provider>{children}</Provider>
      </body>
    </html>
  );
}
