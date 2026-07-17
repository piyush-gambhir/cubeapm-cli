import Link from 'next/link';
import type { Metadata } from 'next';
import { site } from '@/lib/site';

export const metadata: Metadata = {
  title: 'Command not found',
  description: `The requested ${site.name} page could not be found.`,
};

export default function NotFound() {
  return (
    <main className="marketing-shell not-found-page">
      <div className="not-found-page__inner">
        <Link className="not-found-page__brand" href="/">
          <span aria-hidden>&gt;_</span> {site.binary}
        </Link>
        <div className="not-found-page__status" aria-hidden="true">
          404
        </div>
        <h1>Command not found.</h1>
        <div className="not-found-page__terminal" role="status">
          <p>
            <span aria-hidden>$</span> {site.binary} open requested-page
          </p>
          <p>
            <span aria-hidden>&gt;</span> No matching route. Try a known command.
          </p>
        </div>
        <div className="not-found-page__actions">
          <Link className="not-found-page__link is-primary" href="/">
            Return home
          </Link>
          <Link className="not-found-page__link" href="/docs">
            Read the docs
          </Link>
        </div>
      </div>
    </main>
  );
}
