import Link from 'next/link';
import { ArrowRight } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { InstallCommand } from '@/components/install-command';
import { HeroTerminal } from '@/components/hero-terminal';
import { SiteFooter } from '@/components/site-footer';
import { site } from '@/lib/site';

export default function HomePage() {
  const repoUrl = `https://github.com/${site.repo}`;

  return (
    <main className="flex flex-1 flex-col">
      {/* Hero */}
      <section className="relative overflow-hidden">
        {/* soft gradient aurora */}
        <div aria-hidden className="pointer-events-none absolute inset-0 -z-10">
          <div
            className="absolute left-1/2 top-[-12%] size-[36rem] -translate-x-1/2 rounded-full blur-[80px]"
            style={{
              background:
                'radial-gradient(circle, color-mix(in oklab, var(--color-amber-300) 22%, transparent), transparent 70%)',
            }}
          />
          <div
            className="absolute right-[8%] top-[4%] size-[22rem] rounded-full blur-[80px]"
            style={{
              background:
                'radial-gradient(circle, color-mix(in oklab, var(--color-sky-400) 18%, transparent), transparent 72%)',
            }}
          />
        </div>

        <div className="mx-auto flex max-w-5xl flex-col items-center px-4 pt-36 pb-20 text-center sm:pt-44">
          <h1 className="max-w-4xl text-balance text-5xl font-semibold leading-[1.05] tracking-tight sm:text-7xl">
            {site.tagline}
          </h1>
          <p className="mt-6 max-w-2xl text-pretty text-lg leading-relaxed text-muted-foreground">
            {site.description}
          </p>

          <div className="mt-9 flex flex-col items-center gap-5">
            <div className="flex flex-wrap items-center justify-center gap-3">
              <Button size="lg" render={<Link href="/docs" />}>
                Get started
                <ArrowRight className="size-4" />
              </Button>
              <Button
                size="lg"
                variant="secondary"
                render={<Link href={repoUrl} />}
              >
                View on GitHub
              </Button>
            </div>
            <InstallCommand command={site.installCommand} />
          </div>

          {/* Signature terminal visual */}
          <HeroTerminal
            title={site.exampleTitle}
            command={site.example}
            className="mt-16 w-full max-w-3xl text-left"
          />
        </div>
      </section>

      {/* Stack strip */}
      {site.compatible && site.compatible.length > 0 ? (
        <section className="mx-auto w-full max-w-5xl px-4 py-12">
          <p className="text-center text-xs font-medium uppercase tracking-widest text-muted-foreground/80">
            Speaks the language of your stack
          </p>
          <div className="mt-7 flex flex-wrap items-center justify-center gap-x-10 gap-y-4">
            {site.compatible.map((item) => (
              <span
                key={item}
                className="font-mono text-sm font-medium text-fd-foreground/60"
              >
                {item}
              </span>
            ))}
          </div>
        </section>
      ) : null}

      {/* Features */}
      <section className="mx-auto w-full max-w-5xl px-4 py-24">
        <div className="mx-auto max-w-2xl text-center">
          <h2 className="text-3xl font-semibold tracking-tight sm:text-4xl">
            Everything, from one binary
          </h2>
          <p className="mt-4 text-lg text-muted-foreground">
            Built for humans at the keyboard and coding agents alike.
          </p>
        </div>

        <div className="mt-16 grid gap-x-10 gap-y-12 sm:grid-cols-2 lg:grid-cols-3">
          {site.features.map(({ icon: Icon, title, body }) => (
            <div key={title} className="group">
              <div className="mb-4 flex size-11 items-center justify-center rounded-2xl bg-fd-muted text-fd-foreground transition-colors group-hover:bg-fd-primary group-hover:text-fd-primary-foreground">
                <Icon className="size-5" />
              </div>
              <h3 className="text-base font-semibold">{title}</h3>
              <p className="mt-2 text-sm leading-relaxed text-muted-foreground">
                {body}
              </p>
            </div>
          ))}
        </div>
      </section>

      {/* CTA band */}
      <section className="mx-auto w-full max-w-5xl px-4 pb-28">
        <div className="relative overflow-hidden rounded-[2rem] bg-fd-muted/50 px-6 py-20 text-center">
          <div
            aria-hidden
            className="pointer-events-none absolute inset-x-0 top-0 -z-10 h-48"
            style={{
              background:
                'radial-gradient(60% 100% at 50% 0%, color-mix(in oklab, var(--color-amber-200) 40%, transparent), transparent)',
            }}
          />
          <h2 className="mx-auto max-w-xl text-3xl font-semibold tracking-tight sm:text-4xl">
            Ready in one command
          </h2>
          <p className="mx-auto mt-4 max-w-md text-lg text-muted-foreground">
            Install the binary, authenticate, and start querying. No runtime, no
            dependencies.
          </p>
          <div className="mt-9 flex flex-col items-center gap-5">
            <InstallCommand command={site.installCommand} />
            <Button render={<Link href="/docs" />}>
              Read the docs
              <ArrowRight className="size-4" />
            </Button>
          </div>
        </div>
      </section>

      <SiteFooter />
    </main>
  );
}
