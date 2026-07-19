import type { Metadata } from 'next';
import Link from 'next/link';
import { LegalPage } from '@/components/legal-page';
import { createPageMetadata, repoUrl } from '@/lib/metadata';

export const metadata: Metadata = createPageMetadata({
  title: 'Contact',
  description:
    'Contact information for CubeAPM CLI, an independent, unofficial open-source CLI that is not affiliated with CubeAPM or its vendor.',
  path: '/contact',
});

export default function ContactPage() {
  return (
    <LegalPage
      title="Contact"
      lede={
        <p>
          CubeAPM CLI is a free, open-source project maintained by{' '}
          <strong>Piyush Gambhir</strong>. Support is best-effort, here are the
          best ways to get in touch.
        </p>
      }
    >
      <div className="contact-grid">
        <article className="contact-card">
          <p className="contact-card__label">Email</p>
          <p className="contact-card__value">
            <a href="mailto:developer.piyushgambhir@gmail.com">
              developer.piyushgambhir@gmail.com
            </a>
          </p>
          <p>General questions, privacy, and security reports.</p>
        </article>
        <article className="contact-card">
          <p className="contact-card__label">Bugs &amp; features</p>
          <p className="contact-card__value">
            <a href={`${repoUrl}/issues`} target="_blank" rel="noreferrer">
              GitHub Issues ↗
            </a>
          </p>
          <p>The fastest way to report a bug or request a feature.</p>
        </article>
        <article className="contact-card">
          <p className="contact-card__label">Source</p>
          <p className="contact-card__value">
            <a href={repoUrl} target="_blank" rel="noreferrer">
              piyush-gambhir/cubeapm-cli ↗
            </a>
          </p>
          <p>Read the code, open a pull request, or fork it.</p>
        </article>
      </div>
      <section>
        <h2>Security issues</h2>
        <p>
          If you believe you&apos;ve found a security vulnerability, please email{' '}
          <a href="mailto:developer.piyushgambhir@gmail.com">
            developer.piyushgambhir@gmail.com
          </a>{' '}
          with the details rather than opening a public issue. CubeAPM CLI stores
          credentials only on your own device and operates no servers, but
          responsible disclosure is always appreciated.
        </p>
      </section>
      <section>
        <h2>Response time</h2>
        <p>
          This is an independent side project, not a commercial product. The
          maintainer aims to respond when possible, but no response time or
          level of support is guaranteed. See the{' '}
          <Link href="/terms">Terms of Service</Link> for the full no-warranty terms.
        </p>
      </section>
      <section>
        <h2>Not affiliated with CubeAPM</h2>
        <p>
          CubeAPM CLI is an independent, unofficial tool and is not affiliated
          with, endorsed by, or sponsored by CubeAPM or its vendor. For issues
          with CubeAPM itself, contact that vendor&apos;s own support channels.
        </p>
      </section>
    </LegalPage>
  );
}
