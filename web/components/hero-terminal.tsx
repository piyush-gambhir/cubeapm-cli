import { cn } from '@/lib/utils';

// Lightweight, dependency-free shell highlighter for the hero example.
function Line({ line }: { line: string }) {
  if (line.trim() === '') return <span>{'\n'}</span>;

  // Comment line
  if (line.trimStart().startsWith('#')) {
    return <span className="text-fd-muted-foreground/70">{line}</span>;
  }

  const tokens = line.split(/(\s+)/);
  let seenBinary = false;

  return (
    <span>
      {tokens.map((tok, i) => {
        if (/^\s+$/.test(tok)) return <span key={i}>{tok}</span>;

        // first non-space token = the binary
        if (!seenBinary) {
          seenBinary = true;
          return (
            <span key={i} className="text-violet-500 dark:text-violet-400">
              {tok}
            </span>
          );
        }
        if (tok.startsWith('-')) {
          return (
            <span key={i} className="text-amber-600 dark:text-amber-400">
              {tok}
            </span>
          );
        }
        if (/^["'].*["']$/.test(tok) || tok.startsWith('"') || tok.startsWith("'")) {
          return (
            <span key={i} className="text-emerald-600 dark:text-emerald-400">
              {tok}
            </span>
          );
        }
        return (
          <span key={i} className="text-fd-foreground/90">
            {tok}
          </span>
        );
      })}
    </span>
  );
}

export function HeroTerminal({
  title,
  command,
  className,
}: {
  title: string;
  command: string;
  className?: string;
}) {
  const lines = command.split('\n');

  return (
    <div
      className={cn(
        'overflow-hidden rounded-2xl bg-fd-card shadow-[0_24px_80px_-24px_rgba(0,0,0,0.25)]',
        className,
      )}
    >
      {/* titlebar */}
      <div className="flex items-center gap-2 bg-fd-muted/50 px-4 py-3">
        <span className="size-3 rounded-full bg-red-400/90" />
        <span className="size-3 rounded-full bg-amber-400/90" />
        <span className="size-3 rounded-full bg-emerald-400/90" />
        <span className="ml-3 text-xs font-medium text-fd-muted-foreground">
          {title}
        </span>
      </div>
      {/* body */}
      <pre className="overflow-x-auto px-5 py-4 text-left font-mono text-[13px] leading-relaxed sm:text-sm">
        <code>
          {lines.map((line, i) => (
            <span key={i} className="block">
              {!line.trimStart().startsWith('#') && line.trim() !== '' ? (
                <span className="mr-2 select-none text-fd-muted-foreground/50">
                  $
                </span>
              ) : null}
              <Line line={line} />
            </span>
          ))}
        </code>
      </pre>
    </div>
  );
}
