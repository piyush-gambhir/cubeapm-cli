import { cn } from '@/lib/utils';

function Line({ line }: { line: string }) {
  if (line.trim() === '') return <span>{'\n'}</span>;

  // Comment line
  if (line.trimStart().startsWith('#')) {
    return <span className="hero-terminal__comment">{line}</span>;
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
            <span key={i} className="hero-terminal__binary">
              {tok}
            </span>
          );
        }
        if (tok.startsWith('-')) {
          return (
            <span key={i} className="hero-terminal__flag">
              {tok}
            </span>
          );
        }
        if (
          /^["'].*["']$/.test(tok) ||
          tok.startsWith('"') ||
          tok.startsWith("'")
        ) {
          return (
            <span key={i} className="hero-terminal__string">
              {tok}
            </span>
          );
        }
        return (
          <span key={i}>
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
    <div className={cn('hero-terminal', className)}>
      <div className="hero-terminal__titlebar">
        <span className="osmo-eyebrow">{title}</span>
      </div>
      <pre className="hero-terminal__body">
        <code>
          {lines.map((line, i) => (
            <span key={i} className="hero-terminal__line">
              {!line.trimStart().startsWith('#') && line.trim() !== '' ? (
                <span className="hero-terminal__prompt" aria-hidden="true">
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
