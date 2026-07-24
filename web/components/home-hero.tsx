'use client';

import { ArrowRight } from 'lucide-react';
import { HeroTerminal } from '@/components/hero-terminal';
import { InstallCommand } from '@/components/install-command';
import { OsmoButton } from '@/components/ui/osmo-button';
import { site } from '@/lib/site';

export function HomeHero() {
  const taglineWords = site.tagline.split(/\s+/);
  const repoUrl = `https://github.com/${site.repo}`;

  return (
    <section className="osmo-home-hero">
      <div className="osmo-container osmo-home-hero__inner">
        <h1 className="osmo-home-hero__title">
          <span className="home-motion__text-mask">
            <span className="home-motion__text-line">
              {`${taglineWords.slice(0, -1).join(' ')} `}
              <span className="osmo-home-hero__tail">
                {taglineWords[taglineWords.length - 1]}
                <span className="osmo-home-hero__cursor" aria-hidden="true" />
              </span>
            </span>
          </span>
        </h1>
        <p
          className="osmo-home-hero__description"
        >
          {site.description}
        </p>

        <div
          className="osmo-home-hero__actions"
        >
          <OsmoButton
            href="/docs/quickstart"
            aria-label="Get started"
            icon={<ArrowRight />}
          >
            Get started
          </OsmoButton>
          <OsmoButton
            href={repoUrl}
            theme="neutral"
            aria-label="View on GitHub"
          >
            View on GitHub
          </OsmoButton>
        </div>
        <div className="osmo-home-hero__install">
          <InstallCommand command={site.installCommand} />
        </div>

        <div className="osmo-home-hero__terminal">
          <HeroTerminal title={site.exampleTitle} command={site.example} />
        </div>
      </div>
    </section>
  );
}
