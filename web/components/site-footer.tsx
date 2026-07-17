'use client';

import { useState } from 'react';
import Link from 'next/link';
import { ArrowRight } from 'lucide-react';
import { InstallCommand } from '@/components/install-command';
import { OsmoButton } from '@/components/ui/osmo-button';
import { site } from '@/lib/site';

const documentationLinks = [
  { label: 'Introduction', href: '/docs' },
  { label: 'Installation', href: '/docs/installation' },
  { label: 'Authentication', href: '/docs/authentication' },
  { label: 'Quick start', href: '/docs/quickstart' },
];

export function SiteFooter() {
  const repoUrl = `https://github.com/${site.repo}`;
  const year = new Date().getFullYear();
  const groups = [
    {
      title: 'Documentation',
      links: documentationLinks,
    },
    {
      title: 'Project',
      links: [
        { label: 'GitHub', href: repoUrl },
        { label: 'Releases', href: `${repoUrl}/releases` },
        { label: 'Issues', href: `${repoUrl}/issues` },
        { label: 'License', href: `${repoUrl}/blob/main/LICENSE` },
      ],
    },
  ];
  const [activeGroup, setActiveGroup] = useState(groups[0]?.title ?? '');

  return (
    <footer className="site-footer">
      <div className="site-footer__inner">
        <div className="site-footer__top">
          <div className="site-footer__start">
            <h2>Get started in seconds</h2>
            <p className="site-footer__intro">
              Install {site.binary}, then follow the quick start to connect your
              first profile.
            </p>
            <InstallCommand
              command={site.installCommand}
              className="site-footer__install"
            />
            <div className="site-footer__actions">
              <OsmoButton
                href="/docs/quickstart"
                aria-label="Get started"
                icon={<ArrowRight />}
              >
                Get started
              </OsmoButton>
              <a
                className="site-footer__github"
                href={repoUrl}
                target="_blank"
                rel="noreferrer"
                aria-label={`${site.name} on GitHub`}
              >
                <svg viewBox="0 0 24 24" aria-hidden="true">
                  <path
                    fill="currentColor"
                    d="M12 .7A11.5 11.5 0 0 0 8.36 23.1c.58.1.79-.25.79-.56v-2.23c-3.24.7-3.92-1.38-3.92-1.38-.53-1.35-1.3-1.71-1.3-1.71-1.06-.73.08-.71.08-.71 1.17.08 1.79 1.2 1.79 1.2 1.04 1.79 2.73 1.27 3.4.97.1-.76.4-1.27.74-1.56-2.59-.3-5.31-1.3-5.31-5.69 0-1.26.45-2.28 1.2-3.09-.12-.29-.52-1.47.11-3.05 0 0 .98-.31 3.16 1.18a10.9 10.9 0 0 1 5.76 0c2.19-1.49 3.16-1.18 3.16-1.18.63 1.58.23 2.76.11 3.05.75.81 1.2 1.83 1.2 3.09 0 4.4-2.73 5.39-5.32 5.68.42.36.79 1.07.79 2.17v3.22c0 .31.21.67.8.56A11.5 11.5 0 0 0 12 .7Z"
                  />
                </svg>
              </a>
            </div>
          </div>

          <div className="site-footer__directory">
            <div
              className="site-footer__groups"
              data-accordion-close-siblings="true"
            >
              {groups.map((group) => {
                const expanded = group.title === activeGroup;
                const panelId = `footer-${group.title
                  .toLowerCase()
                  .replace(/\s+/g, '-')}`;

                return (
                  <div
                    className="site-footer__group"
                    data-accordion-status={expanded ? 'active' : 'not-active'}
                    key={group.title}
                  >
                    <button
                      type="button"
                      className="site-footer__group-toggle"
                      aria-expanded={expanded}
                      aria-controls={panelId}
                      onClick={() =>
                        setActiveGroup(expanded ? '' : group.title)
                      }
                    >
                      <span>{group.title}</span>
                      <span className="site-footer__plus" aria-hidden="true" />
                    </button>
                    <div className="site-footer__group-panel" id={panelId}>
                      <ul>
                        {group.links.map((link) => {
                          const external = link.href.startsWith('http');

                          return (
                            <li key={link.label}>
                              {external ? (
                                <a
                                  href={link.href}
                                  target="_blank"
                                  rel="noreferrer"
                                >
                                  {link.label}
                                </a>
                              ) : (
                                <Link href={link.href}>{link.label}</Link>
                              )}
                            </li>
                          );
                        })}
                      </ul>
                    </div>
                  </div>
                );
              })}
            </div>

            <p className="site-footer__disclaimer">
              {site.name} is an independent, open-source project. It is{' '}
              <span>not affiliated with, endorsed by, or sponsored by</span>{' '}
              the makers of the underlying software. All product names, logos,
              and trademarks are the property of their respective owners and
              are used for identification purposes only.
            </p>
          </div>
        </div>

        <div className="site-footer__bottom">
          <div className="site-footer__wordmark" aria-hidden="true">
            {site.binary}
          </div>
          <div className="site-footer__details">
            <div className="site-footer__legal">
              <Link href="/privacy">Privacy</Link>
              <Link href="/terms">Terms</Link>
              <Link href="/contact">Contact</Link>
            </div>
            <p>
              © {year} {site.name}
            </p>
            <p className="site-footer__credit">
              <a
                href="https://github.com/piyush-gambhir"
                target="_blank"
                rel="noreferrer"
              >
                Built by Piyush Gambhir
              </a>
            </p>
          </div>
        </div>
      </div>
    </footer>
  );
}
