import Link from 'next/link';
import { Terminal } from 'lucide-react';
import { site } from '@/lib/site';

export function SiteFooter() {
  const repoUrl = `https://github.com/${site.repo}`;

  return (
    <footer className="border-t border-border">
      <div className="mx-auto grid max-w-5xl gap-8 px-4 py-12 sm:grid-cols-2">
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
      <div className="border-t border-border">
        <div className="mx-auto max-w-5xl px-4 py-6 text-xs text-muted-foreground">
          © {site.name} · Open-source · Not affiliated with the upstream project.
        </div>
      </div>
    </footer>
  );
}
