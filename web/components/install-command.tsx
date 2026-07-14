'use client';

import { Check, Copy } from 'lucide-react';
import { useState } from 'react';
import { cn } from '@/lib/utils';

export function InstallCommand({
  command,
  className,
}: {
  command: string;
  className?: string;
}) {
  const [copied, setCopied] = useState(false);

  return (
    <button
      type="button"
      onClick={() => {
        void navigator.clipboard.writeText(command).then(() => {
          setCopied(true);
          setTimeout(() => setCopied(false), 2000);
        });
      }}
      className={cn(
        'group flex w-full max-w-xl items-center gap-3 rounded-xl border border-border bg-fd-card/60 px-4 py-3 text-left font-mono text-sm shadow-sm backdrop-blur transition-colors hover:border-fd-primary/40',
        className,
      )}
      aria-label="Copy install command"
    >
      <span className="select-none text-muted-foreground">$</span>
      <code className="flex-1 truncate text-fd-foreground">{command}</code>
      <span className="shrink-0 text-muted-foreground transition-colors group-hover:text-fd-foreground">
        {copied ? (
          <Check className="size-4 text-emerald-500" />
        ) : (
          <Copy className="size-4" />
        )}
      </span>
    </button>
  );
}
