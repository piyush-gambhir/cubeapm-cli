import type { Metadata } from 'next';
import localFont from 'next/font/local';
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

const hafferVF = localFont({
  src: '../fonts/haffer-vf-thin-2.ttf',
  variable: '--font-haffer-vf',
  weight: '100 1000',
  style: 'normal',
  display: 'swap',
});

const hafferXH = localFont({
  src: '../fonts/haffer-xh-regular-2.woff2',
  variable: '--font-haffer-xh',
  weight: '400',
  style: 'normal',
  display: 'swap',
});

const hafferMono = localFont({
  src: [
    {
      path: '../fonts/haffer-mono-regular-2.woff2',
      weight: '400',
      style: 'normal',
    },
    {
      path: '../fonts/haffer-mono-medium-2.woff2',
      weight: '500',
      style: 'normal',
    },
  ],
  variable: '--font-haffer-mono',
  display: 'swap',
});

const brisa = localFont({
  src: '../fonts/brisa-pro-regular-2.woff2',
  variable: '--font-brisa',
  weight: '400',
  style: 'normal',
  display: 'swap',
});

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

  const fontVariables = [
    hafferVF.variable,
    hafferXH.variable,
    hafferMono.variable,
    brisa.variable,
  ].join(' ');

  return (
    <html
      lang="en"
      className={`${fontVariables} ${hafferVF.className}`}
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
