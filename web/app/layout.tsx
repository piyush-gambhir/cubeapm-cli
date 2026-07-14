import type { Metadata } from 'next';
import localFont from 'next/font/local';
import { Provider } from '@/components/provider';
import { site } from '@/lib/site';
import './global.css';

const satoshi = localFont({
  src: '../fonts/Satoshi-Variable.woff2',
  weight: '300 900',
  display: 'swap',
  variable: '--font-satoshi',
});

export const metadata: Metadata = {
  metadataBase: new URL(`https://${site.repo.split('/')[1]}.pages.dev`),
  title: {
    default: `${site.name} — ${site.tagline}`,
    template: `%s · ${site.name}`,
  },
  description: site.description,
};

export default function Layout({ children }: LayoutProps<'/'>) {
  return (
    <html
      lang="en"
      className={`${satoshi.variable} ${satoshi.className}`}
      suppressHydrationWarning
    >
      <body className="flex flex-col min-h-screen">
        <Provider>{children}</Provider>
      </body>
    </html>
  );
}
