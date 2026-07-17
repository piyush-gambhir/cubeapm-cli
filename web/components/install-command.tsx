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
          setTimeout(() => setCopied(false), 1200);
        });
      }}
      className={cn(
        'install-command group',
        className,
      )}
      aria-label="Copy install command"
    >
      <span className="install-command__prompt" aria-hidden="true">$</span>
      <code className="install-command__code">{command}</code>
      <span className="install-command__icon">
        {copied ? (
          <Check className="is-copied" />
        ) : (
          <Copy />
        )}
      </span>
    </button>
  );
}
