'use client';

import { Fragment, useRef } from 'react';
import { ArrowRight } from 'lucide-react';
import { HeroTerminal } from '@/components/hero-terminal';
import { InstallCommand } from '@/components/install-command';
import { OsmoButton } from '@/components/ui/osmo-button';
import { gsap } from '@/lib/motion/gsap';
import { useGsap } from '@/lib/motion/useGsap';
import { site } from '@/lib/site';

export function HomeHero() {
  const rootRef = useRef<HTMLElement>(null);
  const taglineWords = site.tagline.split(/\s+/);
  const repoUrl = `https://github.com/${site.repo}`;

  useGsap(
    () => {
      const root = rootRef.current;
      if (!root) return;

      const words = gsap.utils.toArray<HTMLElement>('[data-hero-word]', root);
      gsap.set(words, {
        yPercent: 100,
        rotation: 10,
        transformOrigin: 'bottom left',
      });

      gsap.to(words, {
        yPercent: 0,
        rotation: 0,
        autoAlpha: 1,
        duration: 1.2,
        stagger: 0.05,
        ease: 'expo.out',
      });
    },
    [],
    rootRef,
  );

  return (
    <section ref={rootRef} className="osmo-home-hero">
      <div className="osmo-container osmo-home-hero__inner">
        <h1
          className="osmo-home-hero__title"
          aria-label={site.tagline}
        >
          <span className="home-motion__text-mask" aria-hidden="true">
            <span className="home-motion__text-line">
              {taglineWords.slice(0, -1).map((word, index) => (
                <Fragment key={`${word}-${index}`}>
                  <span data-hero-word>{word}</span>{' '}
                </Fragment>
              ))}
              <span className="osmo-home-hero__tail">
                <span data-hero-word>{taglineWords[taglineWords.length - 1]}</span>
                <span className="osmo-home-hero__cursor" />
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
