import { ArrowRight } from 'lucide-react';
import Link from 'next/link';
import { HomeHero } from '@/components/home-hero';
import { InstallCommand } from '@/components/install-command';
import { Reveal } from '@/components/reveal';
import { SiteFooter } from '@/components/site-footer';
import { OsmoButton } from '@/components/ui/osmo-button';
import { site } from '@/lib/site';
import { repoUrl, serializeJsonLd, siteUrl } from '@/lib/seo';
import { getOtherSuiteProjects } from '@/lib/suite';

const revealDelays = ['0s', '0.075s', '0.15s'] as const;

const featureDocLinks = [
  { label: 'Jaeger-compatible traces', href: '/docs/commands/traces' },
  { label: 'named profiles', href: '/docs/authentication' },
  { label: 'structured errors', href: '/docs/output-formats' },
  { label: 'PromQL metrics', href: '/docs/commands/metrics' },
  { label: 'search logs', href: '/docs/commands/logs' },
  {
    label: 'dedicated ingest endpoints',
    href: '/docs/commands/ingest-config',
  },
] as const;

function FeatureBody({ body, index }: { body: string; index: number }) {
  const docLink = featureDocLinks[index];
  if (!docLink) return body;

  const linkStart = body.indexOf(docLink.label);
  if (linkStart === -1) return body;

  return (
    <>
      {body.slice(0, linkStart)}
      <Link href={docLink.href}>{docLink.label}</Link>
      {body.slice(linkStart + docLink.label.length)}
    </>
  );
}

export default function HomePage() {
  const relatedLinks = getOtherSuiteProjects(site.repo).map(
    ({ website }) => website,
  );
  const structuredData = {
    '@context': 'https://schema.org',
    '@graph': [
      {
        '@type': 'SoftwareApplication',
        '@id': `${siteUrl}/#software`,
        name: site.name,
        applicationCategory: 'DeveloperApplication',
        operatingSystem: 'macOS, Linux, Windows',
        license: `${repoUrl}/blob/main/LICENSE`,
        description: site.description,
        featureList: [
          'Structured JSON and YAML output for coding agents',
          'Read-only safety mode for protected mutating API commands',
          'Non-interactive automation with no-input and quiet flags',
          'Works with any coding agent or agent harness that can run shell commands',
        ],
        keywords: [
          'coding agent',
          'AI agent CLI',
          'agent harness',
          'MCP-free shell integration',
          'terminal automation',
          'cubeapm automation',
        ],
        url: siteUrl,
        sameAs: repoUrl,
        relatedLink: relatedLinks,
      },
      {
        '@type': 'WebSite',
        '@id': `${siteUrl}/#website`,
        name: site.name,
        url: siteUrl,
        description: site.description,
        sameAs: repoUrl,
        relatedLink: relatedLinks,
      },
    ],
  };

  return (
    <main className="osmo-home flex flex-1 flex-col">
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: serializeJsonLd(structuredData) }}
      />
      <HomeHero />

      {/* Stack strip */}
      {site.compatible && site.compatible.length > 0 ? (
        <section className="osmo-section osmo-section--compatible">
          <div className="osmo-container">
            <Reveal className="compatible-marquee">
              <div className="compatible-marquee__track">
                {[false, true].map((hidden) => (
                  <span
                    className="compatible-marquee__list"
                    aria-hidden={hidden || undefined}
                    key={String(hidden)}
                  >
                    {site.compatible?.map((item) => (
                      <span className="compatible-marquee__item" key={item}>
                        {item}
                        <span aria-hidden>{' · '}</span>
                      </span>
                    ))}
                  </span>
                ))}
              </div>
            </Reveal>
          </div>
        </section>
      ) : null}

      {/* Features */}
      <section
        className="osmo-section osmo-section--features"
        data-theme-section="dark"
      >
        <div className="osmo-container">
          <Reveal className="osmo-section__header">
            <h2 className="osmo-section__title">
              {site.featuresTitle ?? 'Everything, from one binary'}
            </h2>
            <p className="osmo-section__description">
              {site.featuresSubtitle ??
                'Built for humans at the keyboard and coding agents alike.'}
            </p>
          </Reveal>

          <div className="osmo-card-grid osmo-card-grid--features">
            {site.features.map(({ title, body }, index) => (
              <Reveal
                key={title}
                delay={revealDelays[index % revealDelays.length]}
                className="osmo-card osmo-feature-card"
              >
                <span className="osmo-eyebrow osmo-card__number">
                  {String(index + 1).padStart(2, '0')}
                </span>
                <h3 className="osmo-card__title">{title}</h3>
                <p className="osmo-card__body">
                  <FeatureBody body={body} index={index} />
                </p>
              </Reveal>
            ))}
          </div>
        </div>
      </section>

      {/* CTA band */}
      <section className="osmo-section osmo-section--cta">
        <div className="osmo-container">
          <Reveal className="osmo-cta">
            <h2 className="osmo-cta__title">Ready in one command</h2>
            <p className="osmo-cta__body">
              {site.ctaBody ??
                'Install the binary, authenticate, and start querying. No runtime, no dependencies.'}
            </p>
            <div className="osmo-cta__actions">
              <InstallCommand command={site.installCommand} />
              <OsmoButton
                href="/docs"
                aria-label="Read the docs"
                icon={<ArrowRight />}
              >
                Read the docs
              </OsmoButton>
            </div>
          </Reveal>
        </div>
      </section>

      <SiteFooter />
    </main>
  );
}
