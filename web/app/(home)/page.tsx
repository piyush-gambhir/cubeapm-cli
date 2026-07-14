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
            className="absolute left-1/2 top-[-14%] size-[42rem] -translate-x-1/2 rounded-full blur-[100px]"
            style={{
              background:
                'radial-gradient(circle, color-mix(in oklab, var(--color-amber-200) 38%, transparent), transparent 68%)',
            }}
          />
          <div
            className="absolute right-[6%] top-[8%] size-[24rem] rounded-full blur-[100px]"
            style={{
              background:
                'radial-gradient(circle, color-mix(in oklab, var(--color-sky-300) 30%, transparent), transparent 70%)',
            }}
          />
        </div>

        <div className="mx-auto flex max-w-5xl flex-col items-center px-4 pt-32 pb-20 text-center sm:pt-40">
          <Link
            href={`${repoUrl}/releases`}
            className="mb-8 inline-flex items-center gap-2 rounded-full border border-border bg-fd-card/70 px-3.5 py-1.5 text-xs font-medium text-muted-foreground shadow-sm backdrop-blur transition-colors hover:text-fd-foreground"
          >
            <span className="inline-block size-1.5 rounded-full bg-emerald-500" />
            {site.badge}
          </Link>

          <h1 className="max-w-4xl text-balance font-serif text-5xl font-normal leading-[1.03] tracking-tight sm:text-7xl">
            {site.tagline}
          </h1>
          <p className="mt-6 max-w-2xl text-pretty text-lg leading-relaxed text-muted-foreground">
            {site.description}
          </p>

          <div className="mt-9 flex flex-col items-center gap-4">
            <div className="flex flex-wrap items-center justify-center gap-3">
              <Button size="lg" render={<Link href="/docs" />}>
                Get started
                <ArrowRight className="size-4" />
              </Button>
              <Button size="lg" variant="outline" render={<Link href={repoUrl} />}>
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
        <section className="border-y border-border bg-fd-muted/30">
          <div className="mx-auto max-w-5xl px-4 py-10">
            <p className="text-center text-xs font-medium uppercase tracking-widest text-muted-foreground">
              Speaks the language of your stack
            </p>
            <div className="mt-6 flex flex-wrap items-center justify-center gap-x-8 gap-y-4">
              {site.compatible.map((item) => (
                <span
                  key={item}
                  className="font-mono text-sm font-medium text-fd-foreground/70"
                >
                  {item}
                </span>
              ))}
            </div>
          </div>
        </section>
      ) : null}

      {/* Features */}
      <section className="mx-auto w-full max-w-5xl px-4 py-24">
        <div className="mx-auto max-w-2xl text-center">
          <h2 className="font-serif text-3xl font-normal tracking-tight sm:text-4xl">
            Everything, from one binary
          </h2>
          <p className="mt-4 text-muted-foreground">
            Built for humans at the keyboard and coding agents alike.
          </p>
        </div>

        <div className="mt-14 grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
          {site.features.map(({ icon: Icon, title, body }) => (
            <div
              key={title}
              className="group relative rounded-2xl border border-border bg-fd-card/40 p-6 transition-all hover:-translate-y-0.5 hover:border-fd-primary/30 hover:bg-fd-card/70"
            >
              <div className="mb-4 flex size-10 items-center justify-center rounded-xl bg-fd-muted text-fd-foreground ring-1 ring-inset ring-border/60">
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
      <section className="mx-auto w-full max-w-5xl px-4 pb-24">
        <div className="relative overflow-hidden rounded-3xl border border-border bg-fd-muted/40 px-6 py-16 text-center">
          <div
            aria-hidden
            className="pointer-events-none absolute inset-x-0 top-0 -z-10 h-40"
            style={{
              background:
                'radial-gradient(60% 100% at 50% 0%, color-mix(in oklab, var(--color-amber-200) 45%, transparent), transparent)',
            }}
          />
          <h2 className="mx-auto max-w-xl font-serif text-3xl font-normal tracking-tight sm:text-4xl">
            Ready in one command
          </h2>
          <p className="mx-auto mt-4 max-w-md text-muted-foreground">
            Install the binary, authenticate, and start querying. No runtime, no
            dependencies.
          </p>
          <div className="mt-8 flex flex-col items-center gap-4">
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
