import type { ReactNode } from 'react';
import { FloatingHeader } from '@/components/floating-header';
import { SiteFooter } from '@/components/site-footer';

interface LegalPageProps {
  title: string;
  lede: ReactNode;
  children: ReactNode;
}

export function LegalPage({ title, lede, children }: LegalPageProps) {
  return (
    <div className="marketing-shell route-shell">
      <FloatingHeader />
      <main className="legal-page">
        <header className="legal-page__header">
          <h1>{title}</h1>
          <p className="legal-page__effective">Effective June 14, 2026</p>
        </header>
        <div className="legal-page__lede">{lede}</div>
        <div className="legal-page__body">{children}</div>
      </main>
      <SiteFooter />
    </div>
  );
}
