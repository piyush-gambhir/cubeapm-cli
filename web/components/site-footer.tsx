import Link from 'next/link';
import { Terminal } from 'lucide-react';
import { site } from '@/lib/site';

export function SiteFooter() {
  const repoUrl = `https://github.com/${site.repo}`;

  return (
    <footer className="bg-fd-muted/30">
      <div className="mx-auto grid max-w-5xl gap-8 px-4 py-14 sm:grid-cols-2">
        <div className="max-w-xs">
          <div className="flex items-center gap-2 font-semibold">
            <Terminal className="size-4" />
            {site.name}
          </div>
          <p className="mt-3 text-sm text-muted-foreground">
            {site.description}
          </p>
        </div>

        <div className="grid grid-cols-2 gap-8 sm:justify-items-end">
          <div className="flex flex-col gap-2 text-sm">
            <span className="font-medium">Docs</span>
            <Link href="/docs" className="text-muted-foreground hover:text-fd-foreground">
              Introduction
            </Link>
            <Link href="/docs/installation" className="text-muted-foreground hover:text-fd-foreground">
              Installation
            </Link>
            <Link href="/docs/quickstart" className="text-muted-foreground hover:text-fd-foreground">
              Quick start
            </Link>
          </div>
          <div className="flex flex-col gap-2 text-sm">
            <span className="font-medium">Project</span>
            <Link href={repoUrl} className="text-muted-foreground hover:text-fd-foreground">
              GitHub
            </Link>
            <Link href={`${repoUrl}/releases`} className="text-muted-foreground hover:text-fd-foreground">
              Releases
            </Link>
            <Link href={`${repoUrl}/blob/main/LICENSE`} className="text-muted-foreground hover:text-fd-foreground">
              License
            </Link>
          </div>
        </div>
      </div>
      <div>
        <div className="mx-auto max-w-5xl px-4 pb-10 text-xs leading-relaxed text-muted-foreground">
          <p>
            {site.name} is an independent, open-source project. It is{' '}
            <span className="font-medium text-fd-foreground/80">
              not affiliated with, endorsed by, or sponsored by
            </span>{' '}
            the makers of the underlying software. All product names, logos, and
            trademarks are the property of their respective owners and are used
            for identification purposes only.
          </p>
          <p className="mt-2">© {site.name} · Open-source · MIT-licensed.</p>
        </div>
      </div>
    </footer>
  );
}
